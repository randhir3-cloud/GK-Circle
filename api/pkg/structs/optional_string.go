package structs

import (
	"bytes"
	"encoding/json"
)

// OptionalString distinguishes omitted, JSON null, and string values for PATCH/create bodies.
type OptionalString struct {
	Value   string
	Present bool
	Null    bool
}

func (o *OptionalString) UnmarshalJSON(data []byte) error {
	o.Present = true
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		o.Null = true
		o.Value = ""
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	o.Null = false
	o.Value = value
	return nil
}

func (o OptionalString) MarshalJSON() ([]byte, error) {
	if !o.Present || o.Null {
		return []byte("null"), nil
	}
	return json.Marshal(o.Value)
}
