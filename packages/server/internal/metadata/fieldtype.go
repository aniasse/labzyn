// Package metadata defines the catalog primitives (objects, fields, views) that
// drive Labzyn's dynamic data model — the philosophy borrowed from Twenty.
package metadata

// FieldType enumerates the supported field kinds. Modelled on Twenty's
// FieldMetadataType (24 types) but kept to the subset Labzyn needs first;
// composite/extra types are added as the product grows.
type FieldType string

const (
	FieldText        FieldType = "TEXT"
	FieldNumber      FieldType = "NUMBER"  // integer
	FieldNumeric     FieldType = "NUMERIC" // decimal
	FieldBoolean     FieldType = "BOOLEAN"
	FieldDate        FieldType = "DATE"
	FieldDateTime    FieldType = "DATE_TIME"
	FieldUUID        FieldType = "UUID"
	FieldSelect      FieldType = "SELECT"
	FieldMultiSelect FieldType = "MULTI_SELECT"
	FieldRelation    FieldType = "RELATION"
	FieldRichText    FieldType = "RICH_TEXT"
	FieldRawJSON     FieldType = "RAW_JSON"
	FieldRating      FieldType = "RATING"
	FieldFiles       FieldType = "FILES"
	FieldArray       FieldType = "ARRAY"
)

// allFieldTypes is the validity set for incoming field definitions.
var allFieldTypes = map[FieldType]struct{}{
	FieldText: {}, FieldNumber: {}, FieldNumeric: {}, FieldBoolean: {},
	FieldDate: {}, FieldDateTime: {}, FieldUUID: {}, FieldSelect: {},
	FieldMultiSelect: {}, FieldRelation: {}, FieldRichText: {}, FieldRawJSON: {},
	FieldRating: {}, FieldFiles: {}, FieldArray: {},
}

// IsValid reports whether t is a supported field type.
func (t FieldType) IsValid() bool {
	_, ok := allFieldTypes[t]
	return ok
}

// SortCast returns the Postgres cast to apply when ordering by this field in
// the JSONB store, so typed fields sort by value rather than lexically.
// An empty string means sort as text.
func (t FieldType) SortCast() string {
	switch t {
	case FieldNumber, FieldNumeric, FieldRating:
		return "numeric"
	case FieldDate:
		return "date"
	case FieldDateTime:
		return "timestamptz"
	case FieldBoolean:
		return "boolean"
	default:
		return ""
	}
}

// PostgresColumnType maps a field type to a Postgres column type. Used by the
// future v2 engine that promotes hot objects from JSONB storage to real
// per-workspace tables with native columns (Twenty's approach).
func (t FieldType) PostgresColumnType() string {
	switch t {
	case FieldNumber:
		return "bigint"
	case FieldNumeric, FieldRating:
		return "numeric"
	case FieldBoolean:
		return "boolean"
	case FieldDate:
		return "date"
	case FieldDateTime:
		return "timestamptz"
	case FieldUUID, FieldRelation:
		return "uuid"
	case FieldMultiSelect, FieldArray:
		return "text[]"
	case FieldRawJSON, FieldFiles:
		return "jsonb"
	default: // TEXT, SELECT, RICH_TEXT
		return "text"
	}
}
