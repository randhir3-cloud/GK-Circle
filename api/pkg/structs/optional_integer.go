package structs

import (
	"bytes"
	"encoding/json"
)

// OptionalInteger distinguishes omitted, JSON null, and integer values.
type OptionalInteger struct {
	Value   int
	Present bool
	Null    bool
}

func (o *OptionalInteger) UnmarshalJSON(data []byte) error {
	o.Present = true
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		o.Null = true
		o.Value = 0
		return nil
	}
	var value int
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	o.Null = false
	o.Value = value
	return nil
}

func (o OptionalInteger) MarshalJSON() ([]byte, error) {
	if !o.Present || o.Null {
		return []byte("null"), nil
	}
	return json.Marshal(o.Value)
}
