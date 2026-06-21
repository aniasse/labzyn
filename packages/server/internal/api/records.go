package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/aniasse/labzyn/server/internal/httpx"
	"github.com/aniasse/labzyn/server/internal/metadata"
	"github.com/aniasse/labzyn/server/internal/records"
	"github.com/go-chi/chi/v5"
)

// reservedQuery params are not treated as data filters.
var reservedQuery = map[string]struct{}{
	"limit": {}, "offset": {}, "orderBy": {}, "order": {},
}

// resolveObject loads the object named by the {objectPlural} URL param and its
// fields. It writes a 404/500 and returns ok=false on failure.
func (s *Server) resolveObject(w http.ResponseWriter, r *http.Request) (metadata.Object, []metadata.Field, bool) {
	plural := chi.URLParam(r, "objectPlural")
	obj, err := s.meta.GetObjectByPlural(r.Context(), workspaceID(r), plural)
	if err != nil {
		objectErrorResponse(w, err)
		return metadata.Object{}, nil, false
	}
	fields, err := s.meta.ListFields(r.Context(), obj.ID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not load fields")
		return metadata.Object{}, nil, false
	}
	return obj, fields, true
}

func (s *Server) handleListRecords(w http.ResponseWriter, r *http.Request) {
	obj, fields, ok := s.resolveObject(w, r)
	if !ok {
		return
	}
	byName := make(map[string]metadata.Field, len(fields))
	for _, f := range fields {
		byName[f.Name] = f
	}

	q := r.URL.Query()
	opt := records.ListOptions{
		Limit:   atoiDefault(q.Get("limit"), 50),
		Offset:  atoiDefault(q.Get("offset"), 0),
		Desc:    strings.EqualFold(q.Get("order"), "desc"),
		Filters: map[string]string{},
	}

	if ob := q.Get("orderBy"); ob != "" {
		switch ob {
		case "created_at", "updated_at", "position":
			opt.OrderBy = ob
		default:
			f, isField := byName[ob]
			if !isField {
				httpx.Error(w, http.StatusBadRequest, "unknown orderBy field: "+ob)
				return
			}
			opt.OrderBy = ob
			opt.OrderByCast = f.Type.SortCast()
		}
	}

	for key, vals := range q {
		if _, reserved := reservedQuery[key]; reserved {
			continue
		}
		if _, isField := byName[key]; !isField {
			httpx.Error(w, http.StatusBadRequest, "unknown filter field: "+key)
			return
		}
		opt.Filters[key] = vals[0]
	}

	list, total, err := s.recList(r.Context(), obj, fields, workspaceID(r), opt)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not list records")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"data":   list,
		"total":  total,
		"limit":  opt.Limit,
		"offset": opt.Offset,
	})
}

func (s *Server) handleCreateRecord(w http.ResponseWriter, r *http.Request) {
	obj, fields, ok := s.resolveObject(w, r)
	if !ok {
		return
	}
	var data map[string]any
	if err := decodeJSON(r, &data); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := records.Validate(fields, data, true); err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	rec, err := s.recCreate(r.Context(), obj, workspaceID(r), data)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not create record")
		return
	}
	httpx.JSON(w, http.StatusCreated, rec)
}

func (s *Server) handleGetRecord(w http.ResponseWriter, r *http.Request) {
	obj, _, ok := s.resolveObject(w, r)
	if !ok {
		return
	}
	rec, err := s.recGet(r.Context(), obj, workspaceID(r), chi.URLParam(r, "id"))
	if err != nil {
		recordErrorResponse(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, rec)
}

func (s *Server) handleUpdateRecord(w http.ResponseWriter, r *http.Request) {
	obj, fields, ok := s.resolveObject(w, r)
	if !ok {
		return
	}
	var patch map[string]any
	if err := decodeJSON(r, &patch); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := records.Validate(fields, patch, false); err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	rec, err := s.recUpdate(r.Context(), obj, workspaceID(r), chi.URLParam(r, "id"), patch)
	if err != nil {
		recordErrorResponse(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, rec)
}

func (s *Server) handleDeleteRecord(w http.ResponseWriter, r *http.Request) {
	obj, _, ok := s.resolveObject(w, r)
	if !ok {
		return
	}
	if err := s.recDelete(r.Context(), obj, workspaceID(r), chi.URLParam(r, "id")); err != nil {
		recordErrorResponse(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func recordErrorResponse(w http.ResponseWriter, err error) {
	if errors.Is(err, records.ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "record not found")
		return
	}
	httpx.Error(w, http.StatusInternalServerError, "record operation failed")
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}
