package models

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestNormalizeLearningItemBlockVisibility(t *testing.T) {
	t.Run("omitted defaults to ALL", func(t *testing.T) {
		got, err := normalizeLearningItemBlockVisibility(nil)
		if err != nil {
			t.Fatalf("normalize: %v", err)
		}
		if got.Mode != LearningItemVisibilityModeAll {
			t.Fatalf("mode = %s, want ALL", got.Mode)
		}
	})

	validModes := []LearningItemVisibilityMode{
		LearningItemVisibilityModeAll,
		LearningItemVisibilityModeAuthenticated,
		LearningItemVisibilityModePremium,
		LearningItemVisibilityModeInstructor,
		LearningItemVisibilityModeHidden,
	}
	for _, mode := range validModes {
		t.Run(string(mode), func(t *testing.T) {
			raw := json.RawMessage(`{"mode":"` + string(mode) + `"}`)
			got, err := normalizeLearningItemBlockVisibility(raw)
			if err != nil {
				t.Fatalf("normalize: %v", err)
			}
			if got.Mode != mode {
				t.Fatalf("mode = %s, want %s", got.Mode, mode)
			}
		})
	}

	tests := []struct {
		name string
		raw  json.RawMessage
		want error
	}{
		{name: "null", raw: json.RawMessage(`null`), want: ErrLearningItemVisibilityInvalid},
		{name: "array", raw: json.RawMessage(`[]`), want: ErrLearningItemVisibilityInvalid},
		{name: "scalar", raw: json.RawMessage(`"ALL"`), want: ErrLearningItemVisibilityInvalid},
		{name: "missing mode", raw: json.RawMessage(`{}`), want: ErrLearningItemVisibilityInvalid},
		{name: "unknown mode", raw: json.RawMessage(`{"mode":"PUBLIC"}`), want: ErrLearningItemVisibilityModeInvalid},
		{name: "null mode", raw: json.RawMessage(`{"mode":null}`), want: ErrLearningItemVisibilityModeInvalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := normalizeLearningItemBlockVisibility(test.raw)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestNormalizeLearningItemMetadataVisibility(t *testing.T) {
	t.Run("omitted visibility persisted as ALL", func(t *testing.T) {
		input := json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{"text":"Welcome"}}]}`)
		got, err := normalizeLearningItemMetadata(input)
		if err != nil {
			t.Fatalf("normalize: %v", err)
		}
		var decoded LearningItemMetadata
		if err := json.Unmarshal(got, &decoded); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if len(decoded.Blocks) != 1 {
			t.Fatalf("blocks = %d", len(decoded.Blocks))
		}
		if decoded.Blocks[0].Visibility.Mode != LearningItemVisibilityModeAll {
			t.Fatalf("visibility = %+v", decoded.Blocks[0].Visibility)
		}
	})

	t.Run("explicit PREMIUM preserved", func(t *testing.T) {
		input := json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{},"visibility":{"mode":"PREMIUM"}}]}`)
		got, err := normalizeLearningItemMetadata(input)
		if err != nil {
			t.Fatalf("normalize: %v", err)
		}
		var decoded LearningItemMetadata
		if err := json.Unmarshal(got, &decoded); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if decoded.Blocks[0].Visibility.Mode != LearningItemVisibilityModePremium {
			t.Fatalf("mode = %s", decoded.Blocks[0].Visibility.Mode)
		}
	})

	t.Run("null visibility rejected", func(t *testing.T) {
		input := json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{},"visibility":null}]}`)
		_, err := normalizeLearningItemMetadata(input)
		if !errors.Is(err, ErrLearningItemVisibilityInvalid) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemVisibilityInvalid)
		}
	})

	t.Run("unknown mode rejected", func(t *testing.T) {
		input := json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{},"visibility":{"mode":"PUBLIC"}}]}`)
		_, err := normalizeLearningItemMetadata(input)
		if !errors.Is(err, ErrLearningItemVisibilityModeInvalid) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemVisibilityModeInvalid)
		}
	})
}
