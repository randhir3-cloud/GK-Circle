package structs

import (
	"encoding/json"
	"testing"
)

func TestOptionalIntegerUnmarshalPresence(t *testing.T) {
	type body struct {
		Field OptionalInteger `json:"field"`
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
	t.Run("zero", func(t *testing.T) {
		var parsed body
		if err := json.Unmarshal([]byte(`{"field":0}`), &parsed); err != nil {
			t.Fatal(err)
		}
		if !parsed.Field.Present || parsed.Field.Null || parsed.Field.Value != 0 {
			t.Fatalf("zero = %+v", parsed.Field)
		}
	})
	t.Run("positive", func(t *testing.T) {
		var parsed body
		if err := json.Unmarshal([]byte(`{"field":3}`), &parsed); err != nil {
			t.Fatal(err)
		}
		if !parsed.Field.Present || parsed.Field.Null || parsed.Field.Value != 3 {
			t.Fatalf("positive = %+v", parsed.Field)
		}
	})
}
