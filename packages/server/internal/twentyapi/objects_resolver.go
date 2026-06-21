package twentyapi

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/aniasse/labzyn/server/internal/metadata"
)

// ── objects connection types (UseFieldResolvers) ─────────────────────────────

type cursorPagingInput struct {
	First  *int32
	After  *string
	Last   *int32
	Before *string
}

type objectConnectionT struct {
	Edges    []*objectEdgeT
	PageInfo pageInfoT
}

type objectEdgeT struct{ Node *objectT }

type pageInfoT struct {
	HasNextPage     *bool
	HasPreviousPage *bool
	StartCursor     *string
	EndCursor       *string
}

type objectT struct {
	Id                             UUID
	UniversalIdentifier            *string
	NameSingular                   string
	NamePlural                     string
	LabelSingular                  string
	LabelPlural                    string
	Color                          *string
	Description                    *string
	Icon                           *string
	IsCustom                       bool
	IsRemote                       bool
	IsActive                       bool
	IsSystem                       bool
	IsUIEditable                   bool
	IsUICreatable                  bool
	CreatedAt                      DateTime
	UpdatedAt                      DateTime
	LabelIdentifierFieldMetadataId *UUID
	ImageIdentifierFieldMetadataId *UUID
	ApplicationId                  *string
	Shortcut                       *string
	IsLabelSyncedWithName          bool
	IsSearchable                   bool
	DuplicateCriteria              *JSON
	IndexMetadataList              []*indexMetadataT
	FieldsList                     []*fieldT
}

type indexMetadataT struct {
	Id                     UUID
	CreatedAt              DateTime
	UpdatedAt              DateTime
	Name                   string
	IndexWhereClause       *string
	IndexType              string
	IsUnique               bool
	IsCustom               bool
	IndexFieldMetadataList []*indexFieldMetadataT
}

type indexFieldMetadataT struct {
	Id              UUID
	FieldMetadataId UUID
	SubFieldName    *string
	CreatedAt       DateTime
	UpdatedAt       DateTime
	Order           int32
}

type fieldT struct {
	Id                    UUID
	UniversalIdentifier   *string
	Type                  string
	Name                  string
	Label                 string
	Description           *string
	Icon                  *string
	IsCustom              bool
	IsActive              bool
	IsSystem              bool
	IsUIEditable          bool
	IsNullable            bool
	IsUnique              *bool
	CreatedAt             DateTime
	UpdatedAt             DateTime
	DefaultValue          *JSON
	Options               *JSON
	Settings              *JSON
	IsLabelSyncedWithName bool
	MorphId               *string
	ApplicationId         *string
	Relation              *fieldRelationT
	MorphRelations        []*fieldRelationT
}

type fieldRelationT struct {
	Type                 string
	SourceObjectMetadata relationObjectRefT
	TargetObjectMetadata relationObjectRefT
	SourceFieldMetadata  relationFieldRefT
	TargetFieldMetadata  relationFieldRefT
}

type relationObjectRefT struct {
	Id           UUID
	NameSingular string
	NamePlural   string
}

type relationFieldRefT struct {
	Id   UUID
	Name string
}

// Objects returns the full object metadata (with fields) as a Relay connection.
func (r *MetadataResolver) Objects(ctx context.Context, _ struct{ Paging *cursorPagingInput }) (*objectConnectionT, error) {
	wsID, err := r.core.DefaultWorkspaceID(ctx, userIDFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	objs, err := r.meta.ListObjects(ctx, wsID)
	if err != nil {
		return nil, err
	}

	edges := []*objectEdgeT{}
	for _, o := range objs {
		fields, err := r.meta.ListFields(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		node := buildObjectNode(o, fields)
		edges = append(edges, &objectEdgeT{Node: node})
	}
	// Twenty's standard objects (workspaceMember, …).
	for _, so := range standardObjects() {
		edges = append(edges, &objectEdgeT{Node: buildObjectNode(so.Object, so.Fields)})
	}

	f := false
	empty := ""
	return &objectConnectionT{
		Edges:    edges,
		PageInfo: pageInfoT{HasNextPage: &f, HasPreviousPage: &f, StartCursor: &empty, EndCursor: &empty},
	}, nil
}

func buildObjectNode(o metadata.Object, fields []metadata.Field) *objectT {
	icon := o.Icon
	created := DateTime{o.CreatedAt}
	updated := DateTime{o.UpdatedAt}

	list := []*fieldT{}
	// System fields Twenty always expects.
	list = append(list,
		systemField(o.ID, "id", "Id", "UUID", created, updated),
		systemField(o.ID, "createdAt", "Creation date", "DATE_TIME", created, updated),
		systemField(o.ID, "updatedAt", "Last update", "DATE_TIME", created, updated),
	)

	var labelID *UUID
	for _, fld := range fields {
		id := UUID(fld.ID)
		if labelID == nil && fld.Type == "TEXT" {
			labelID = &id
		}
		list = append(list, userField(fld))
	}
	if labelID == nil && len(list) > 0 {
		labelID = &list[0].Id
	}

	return &objectT{
		Id:                             UUID(o.ID),
		NameSingular:                   o.NameSingular,
		NamePlural:                     o.NamePlural,
		LabelSingular:                  o.LabelSingular,
		LabelPlural:                    o.LabelPlural,
		Icon:                           &icon,
		IsCustom:                       !o.IsSystem,
		IsRemote:                       false,
		IsActive:                       o.IsActive,
		IsSystem:                       o.IsSystem,
		IsUIEditable:                   true,
		IsUICreatable:                  true,
		CreatedAt:                      created,
		UpdatedAt:                      updated,
		LabelIdentifierFieldMetadataId: labelID,
		IsLabelSyncedWithName:          false,
		IsSearchable:                   true,
		IndexMetadataList:              []*indexMetadataT{},
		FieldsList:                     list,
	}
}

func systemField(objectID, name, label, ftype string, created, updated DateTime) *fieldT {
	id := synthFieldID(objectID, name)
	nullable := name != "id"
	return &fieldT{
		Id:             id,
		Type:           ftype,
		Name:           name,
		Label:          label,
		IsCustom:       false,
		IsActive:       true,
		IsSystem:       true,
		IsUIEditable:   false,
		IsNullable:     nullable,
		CreatedAt:      created,
		UpdatedAt:      updated,
		MorphRelations: []*fieldRelationT{},
	}
}

func userField(f metadata.Field) *fieldT {
	var desc *string
	if f.Description != "" {
		desc = &f.Description
	}
	icon := f.Icon
	uniq := f.IsUnique
	out := &fieldT{
		Id:             UUID(f.ID),
		Type:           string(f.Type),
		Name:           f.Name,
		Label:          f.Label,
		Description:    desc,
		Icon:           &icon,
		IsCustom:       !f.IsSystem,
		IsActive:       f.IsActive,
		IsSystem:       f.IsSystem,
		IsUIEditable:   true,
		IsNullable:     f.IsNullable,
		IsUnique:       &uniq,
		CreatedAt:      DateTime{f.CreatedAt},
		UpdatedAt:      DateTime{f.UpdatedAt},
		MorphRelations: []*fieldRelationT{},
	}
	if len(f.DefaultValue) > 0 {
		out.DefaultValue = &JSON{Value: f.DefaultValue}
	}
	if len(f.Options) > 0 {
		out.Options = &JSON{Value: f.Options}
	}
	return out
}

// synthFieldID derives a stable UUID for a synthesized system field.
func synthFieldID(objectID, name string) UUID {
	h := sha256.Sum256([]byte(objectID + ":" + name))
	b := h[:16]
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return UUID(fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]))
}
