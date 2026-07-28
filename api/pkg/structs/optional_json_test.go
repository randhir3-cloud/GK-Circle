package structs

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestOptionalJSONUnmarshalPresence(t *testing.T) {
	type body struct {
		Field OptionalJSON `json:"field"`
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
	t.Run("object", func(t *testing.T) {
		var parsed body
		raw := []byte(`{"field":{"version":1,"blocks":[]}}`)
		if err := json.Unmarshal(raw, &parsed); err != nil {
			t.Fatal(err)
		}
		if !parsed.Field.Present || parsed.Field.Null {
			t.Fatalf("object = %+v", parsed.Field)
		}
		if !bytes.Equal(parsed.Field.Value, []byte(`{"version":1,"blocks":[]}`)) {
			t.Fatalf("value = %s", parsed.Field.Value)
		}
	})
}
