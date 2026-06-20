package metadata

import (
	"encoding/json"
	"time"
)

// Object is a catalog entry describing a logical "table" (a user-defined or
// system object) within a workspace.
type Object struct {
	ID            string          `json:"id"`
	WorkspaceID   string          `json:"workspaceId"`
	NameSingular  string          `json:"nameSingular"`
	NamePlural    string          `json:"namePlural"`
	LabelSingular string          `json:"labelSingular"`
	LabelPlural   string          `json:"labelPlural"`
	Labels        json.RawMessage `json:"labels"`
	Icon          string          `json:"icon"`
	Description   string          `json:"description"`
	IsSystem      bool            `json:"isSystem"`
	IsActive      bool            `json:"isActive"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`

	// Fields is populated only when an object is fetched with its fields.
	Fields []Field `json:"fields,omitempty"`
}

// Field is a catalog entry describing a single attribute of an object.
type Field struct {
	ID           string          `json:"id"`
	ObjectID     string          `json:"objectId"`
	Name         string          `json:"name"`
	Type         FieldType       `json:"type"`
	Label        string          `json:"label"`
	Labels       json.RawMessage `json:"labels"`
	Icon         string          `json:"icon"`
	Description  string          `json:"description"`
	IsNullable   bool            `json:"isNullable"`
	IsUnique     bool            `json:"isUnique"`
	IsSystem     bool            `json:"isSystem"`
	IsActive     bool            `json:"isActive"`
	DefaultValue json.RawMessage `json:"defaultValue,omitempty"`
	Options      json.RawMessage `json:"options,omitempty"`
	Position     int             `json:"position"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
}
