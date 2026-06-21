package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/aniasse/labzyn/server/internal/httpx"
	"github.com/aniasse/labzyn/server/internal/metadata"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (s *Server) handleCreateObject(w http.ResponseWriter, r *http.Request) {
	var in struct {
		NameSingular  string          `json:"nameSingular"`
		NamePlural    string          `json:"namePlural"`
		LabelSingular string          `json:"labelSingular"`
		LabelPlural   string          `json:"labelPlural"`
		Labels        json.RawMessage `json:"labels"`
		Icon          string          `json:"icon"`
		Description   string          `json:"description"`
		Storage       string          `json:"storage"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if !validName(in.NameSingular) || !validName(in.NamePlural) {
		httpx.Error(w, http.StatusBadRequest,
			"nameSingular and namePlural must match ^[a-zA-Z][a-zA-Z0-9_]*$")
		return
	}
	if in.Storage != "" && in.Storage != "jsonb" && in.Storage != "table" {
		httpx.Error(w, http.StatusBadRequest, "storage must be \"jsonb\" or \"table\"")
		return
	}
	if in.LabelSingular == "" {
		in.LabelSingular = in.NameSingular
	}
	if in.LabelPlural == "" {
		in.LabelPlural = in.NamePlural
	}

	obj, err := s.meta.CreateObject(r.Context(), workspaceID(r), metadata.CreateObjectInput{
		NameSingular:  in.NameSingular,
		NamePlural:    in.NamePlural,
		LabelSingular: in.LabelSingular,
		LabelPlural:   in.LabelPlural,
		Labels:        in.Labels,
		Icon:          in.Icon,
		Description:   in.Description,
		Storage:       in.Storage,
	})
	if err != nil {
		if isUniqueViolation(err) {
			httpx.Error(w, http.StatusConflict, "an object with this name already exists")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "could not create object")
		return
	}

	// Materialize a real table for table-backed objects.
	if isTableBacked(obj) {
		if err := s.ddl.EnsureTable(r.Context(), workspaceID(r), obj, nil); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not create table: "+err.Error())
			return
		}
	}
	httpx.JSON(w, http.StatusCreated, obj)
}

func (s *Server) handleListObjects(w http.ResponseWriter, r *http.Request) {
	list, err := s.meta.ListObjects(r.Context(), workspaceID(r))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not list objects")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"data": list})
}

func (s *Server) handleGetObject(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "nameSingular")
	obj, err := s.meta.GetObjectBySingular(r.Context(), workspaceID(r), name)
	if err != nil {
		objectErrorResponse(w, err)
		return
	}
	fields, err := s.meta.ListFields(r.Context(), obj.ID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not load fields")
		return
	}
	obj.Fields = fields
	httpx.JSON(w, http.StatusOK, obj)
}

func (s *Server) handleCreateField(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "nameSingular")
	obj, err := s.meta.GetObjectBySingular(r.Context(), workspaceID(r), name)
	if err != nil {
		objectErrorResponse(w, err)
		return
	}

	var in struct {
		Name         string          `json:"name"`
		Type         string          `json:"type"`
		Label        string          `json:"label"`
		Labels       json.RawMessage `json:"labels"`
		Icon         string          `json:"icon"`
		Description  string          `json:"description"`
		IsNullable   *bool           `json:"isNullable"`
		IsUnique     bool            `json:"isUnique"`
		DefaultValue json.RawMessage `json:"defaultValue"`
		Options      json.RawMessage `json:"options"`
		Position     int             `json:"position"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if !validName(in.Name) {
		httpx.Error(w, http.StatusBadRequest, "name must match ^[a-zA-Z][a-zA-Z0-9_]*$")
		return
	}
	ft := metadata.FieldType(strings.ToUpper(in.Type))
	if !ft.IsValid() {
		httpx.Error(w, http.StatusBadRequest, "invalid field type: "+in.Type)
		return
	}
	if in.Label == "" {
		in.Label = in.Name
	}
	nullable := true
	if in.IsNullable != nil {
		nullable = *in.IsNullable
	}

	field, err := s.meta.CreateField(r.Context(), obj.ID, metadata.CreateFieldInput{
		Name:         in.Name,
		Type:         ft,
		Label:        in.Label,
		Labels:       in.Labels,
		Icon:         in.Icon,
		Description:  in.Description,
		IsNullable:   nullable,
		IsUnique:     in.IsUnique,
		DefaultValue: in.DefaultValue,
		Options:      in.Options,
		Position:     in.Position,
	})
	if err != nil {
		if isUniqueViolation(err) {
			httpx.Error(w, http.StatusConflict, "a field with this name already exists on the object")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "could not create field")
		return
	}

	// Add the column on table-backed objects.
	if isTableBacked(obj) {
		if err := s.ddl.AddColumn(r.Context(), workspaceID(r), obj, field); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not add column: "+err.Error())
			return
		}
	}
	httpx.JSON(w, http.StatusCreated, field)
}

func (s *Server) handleUpdateObject(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "nameSingular")
	var in struct {
		LabelSingular *string `json:"labelSingular"`
		LabelPlural   *string `json:"labelPlural"`
		Icon          *string `json:"icon"`
		Description   *string `json:"description"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	obj, err := s.meta.UpdateObject(r.Context(), workspaceID(r), name, metadata.UpdateObjectInput{
		LabelSingular: in.LabelSingular, LabelPlural: in.LabelPlural, Icon: in.Icon, Description: in.Description,
	})
	if err != nil {
		objectErrorResponse(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, obj)
}

func (s *Server) handleDeleteObject(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "nameSingular")
	obj, err := s.meta.GetObjectBySingular(r.Context(), workspaceID(r), name)
	if err != nil {
		objectErrorResponse(w, err)
		return
	}
	if err := s.meta.DeleteObject(r.Context(), workspaceID(r), name); err != nil {
		objectErrorResponse(w, err)
		return
	}
	if isTableBacked(obj) {
		if err := s.ddl.DropTable(r.Context(), workspaceID(r), obj); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "object deleted but table drop failed: "+err.Error())
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleUpdateField(w http.ResponseWriter, r *http.Request) {
	obj, err := s.meta.GetObjectBySingular(r.Context(), workspaceID(r), chi.URLParam(r, "nameSingular"))
	if err != nil {
		objectErrorResponse(w, err)
		return
	}
	var in struct {
		Label string `json:"label"`
	}
	if err := decodeJSON(r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	field, err := s.meta.UpdateField(r.Context(), obj.ID, chi.URLParam(r, "fieldId"), in.Label)
	if err != nil {
		objectErrorResponse(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, field)
}

func (s *Server) handleDeleteField(w http.ResponseWriter, r *http.Request) {
	obj, err := s.meta.GetObjectBySingular(r.Context(), workspaceID(r), chi.URLParam(r, "nameSingular"))
	if err != nil {
		objectErrorResponse(w, err)
		return
	}
	fieldID := chi.URLParam(r, "fieldId")

	// Resolve the field name (needed to drop the column on table-backed objects).
	fieldName := ""
	if isTableBacked(obj) {
		fields, err := s.meta.ListFields(r.Context(), obj.ID)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not load fields")
			return
		}
		for _, f := range fields {
			if f.ID == fieldID {
				fieldName = f.Name
				break
			}
		}
	}

	if err := s.meta.DeleteField(r.Context(), obj.ID, fieldID); err != nil {
		objectErrorResponse(w, err)
		return
	}
	if isTableBacked(obj) && fieldName != "" {
		if err := s.ddl.DropColumn(r.Context(), workspaceID(r), obj, fieldName); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "field deleted but column drop failed: "+err.Error())
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func objectErrorResponse(w http.ResponseWriter, err error) {
	if errors.Is(err, metadata.ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "object not found")
		return
	}
	httpx.Error(w, http.StatusInternalServerError, "could not load object")
}
