package structs

import (
	"bytes"
	"encoding/json"
)

// OptionalJSON distinguishes omitted, JSON null, and present raw JSON values.
type OptionalJSON struct {
	Value   json.RawMessage
	Present bool
	Null    bool
}

func (o *OptionalJSON) UnmarshalJSON(data []byte) error {
	o.Present = true
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		o.Null = true
		o.Value = nil
		return nil
	}
	o.Null = false
	o.Value = append(json.RawMessage(nil), trimmed...)
	return nil
}

func (o OptionalJSON) MarshalJSON() ([]byte, error) {
	if !o.Present || o.Null {
		return []byte("null"), nil
	}
	if len(o.Value) == 0 {
		return []byte("null"), nil
	}
	return append([]byte(nil), o.Value...), nil
}
