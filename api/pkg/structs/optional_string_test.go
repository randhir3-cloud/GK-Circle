package structs

import (
	"encoding/json"
	"testing"
)

func TestOptionalStringUnmarshalPresence(t *testing.T) {
	type body struct {
		Field OptionalString `json:"field"`
	}

	t.Run("omitted", func(t *testing.T) {
		var parsed body
		if err := json.Unmarshal([]byte(`{}`), &parsed); err != nil {
			t.Fatal(err)
		}
		if parsed.Field.Present || parsed.Field.Null {
			t.Fatalf("omitted = %+v", parsed.Field)
		}
	})
	t.Run("null", func(t *testing.T) {
		var parsed body
		if err := json.Unmarshal([]byte(`{"field":null}`), &parsed); err != nil {
			t.Fatal(err)
		}
		if !parsed.Field.Present || !parsed.Field.Null {
			t.Fatalf("null = %+v", parsed.Field)
		}
	})
	t.Run("empty string", func(t *testing.T) {
		var parsed body
		if err := json.Unmarshal([]byte(`{"field":""}`), &parsed); err != nil {
			t.Fatal(err)
		}
		if !parsed.Field.Present || parsed.Field.Null || parsed.Field.Value != "" {
			t.Fatalf("empty = %+v", parsed.Field)
		}
	})
	t.Run("value", func(t *testing.T) {
		var parsed body
		if err := json.Unmarshal([]byte(`{"field":"en"}`), &parsed); err != nil {
			t.Fatal(err)
		}
		if !parsed.Field.Present || parsed.Field.Null || parsed.Field.Value != "en" {
			t.Fatalf("value = %+v", parsed.Field)
		}
	})
}
