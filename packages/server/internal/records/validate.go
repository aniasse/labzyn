package records

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aniasse/labzyn/server/internal/metadata"
)

// ValidationError aggregates field-level problems.
type ValidationError struct{ Messages []string }

func (e *ValidationError) Error() string { return strings.Join(e.Messages, "; ") }

// Validate checks data against the object's field catalog. When isCreate is
// true, required (non-nullable, no default) fields must be present.
func Validate(fields []metadata.Field, data map[string]any, isCreate bool) error {
	byName := make(map[string]metadata.Field, len(fields))
	for _, f := range fields {
		byName[f.Name] = f
	}

	var msgs []string

	// Unknown keys + type checks on provided values.
	for key, val := range data {
		f, ok := byName[key]
		if !ok {
			msgs = append(msgs, fmt.Sprintf("unknown field %q", key))
			continue
		}
		if val == nil {
			if !f.IsNullable {
				msgs = append(msgs, fmt.Sprintf("field %q cannot be null", key))
			}
			continue
		}
		if err := checkType(f, val); err != nil {
			msgs = append(msgs, err.Error())
		}
	}

	// Required fields on create.
	if isCreate {
		for _, f := range fields {
			if f.IsSystem || f.IsNullable || len(f.DefaultValue) > 0 {
				continue
			}
			if _, present := data[f.Name]; !present {
				msgs = append(msgs, fmt.Sprintf("field %q is required", f.Name))
			}
		}
	}

	if len(msgs) > 0 {
		return &ValidationError{Messages: msgs}
	}
	return nil
}

func checkType(f metadata.Field, val any) error {
	bad := func(want string) error {
		return fmt.Errorf("field %q expects %s", f.Name, want)
	}
	switch f.Type {
	case metadata.FieldText, metadata.FieldRichText, metadata.FieldUUID,
		metadata.FieldDate, metadata.FieldDateTime, metadata.FieldRelation:
		if _, ok := val.(string); !ok {
			return bad("a string")
		}
	case metadata.FieldSelect:
		s, ok := val.(string)
		if !ok {
			return bad("a string")
		}
		if opts := selectValues(f.Options); len(opts) > 0 {
			if _, ok := opts[s]; !ok {
				return fmt.Errorf("field %q value %q is not an allowed option", f.Name, s)
			}
		}
	case metadata.FieldNumber, metadata.FieldNumeric, metadata.FieldRating:
		if _, ok := val.(float64); !ok { // JSON numbers decode to float64
			return bad("a number")
		}
	case metadata.FieldBoolean:
		if _, ok := val.(bool); !ok {
			return bad("a boolean")
		}
	case metadata.FieldMultiSelect, metadata.FieldArray:
		if _, ok := val.([]any); !ok {
			return bad("an array")
		}
	case metadata.FieldRawJSON, metadata.FieldFiles:
		// any JSON value is acceptable
	default:
		// unknown types are permissive
	}
	return nil
}

// selectValues parses a field's options into a set of allowed values.
// Accepts either ["a","b"] or [{"value":"a"},{"value":"b"}].
func selectValues(raw json.RawMessage) map[string]struct{} {
	if len(raw) == 0 {
		return nil
	}
	out := map[string]struct{}{}

	var strs []string
	if err := json.Unmarshal(raw, &strs); err == nil {
		for _, s := range strs {
			out[s] = struct{}{}
		}
		return out
	}

	var objs []struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(raw, &objs); err == nil {
		for _, o := range objs {
			if o.Value != "" {
				out[o.Value] = struct{}{}
			}
		}
		return out
	}
	return nil
}
