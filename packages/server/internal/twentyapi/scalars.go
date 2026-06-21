package twentyapi

import (
	"encoding/json"
	"fmt"
	"time"
)

// Custom GraphQL scalars matching Twenty's schema (UUID, DateTime, JSON).

// UUID scalar.
type UUID string

func (UUID) ImplementsGraphQLType(name string) bool { return name == "UUID" }
func (u *UUID) UnmarshalGraphQL(input any) error {
	s, ok := input.(string)
	if !ok {
		return fmt.Errorf("UUID must be a string")
	}
	*u = UUID(s)
	return nil
}
func (u UUID) MarshalJSON() ([]byte, error) { return json.Marshal(string(u)) }

// DateTime scalar (RFC3339).
type DateTime struct{ time.Time }

func (DateTime) ImplementsGraphQLType(name string) bool { return name == "DateTime" }
func (d *DateTime) UnmarshalGraphQL(input any) error {
	switch v := input.(type) {
	case string:
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return err
		}
		d.Time = t
		return nil
	case time.Time:
		d.Time = v
		return nil
	default:
		return fmt.Errorf("DateTime must be a string")
	}
}
func (d DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Time.UTC().Format(time.RFC3339))
}

// JSON scalar (arbitrary JSON value).
type JSON struct{ Value any }

func (JSON) ImplementsGraphQLType(name string) bool { return name == "JSON" }
func (j *JSON) UnmarshalGraphQL(input any) error    { j.Value = input; return nil }
func (j JSON) MarshalJSON() ([]byte, error)         { return json.Marshal(j.Value) }
