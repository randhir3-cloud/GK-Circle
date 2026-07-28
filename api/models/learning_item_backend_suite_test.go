package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestLearningItemBackendSuiteAcceptedMetadataBlockTypes(t *testing.T) {
	accepted := []InformationBlockType{
		InformationBlockTypeText,
		InformationBlockTypeHeading,
		InformationBlockTypeImage,
		InformationBlockTypeVideo,
		InformationBlockTypePDF,
		InformationBlockTypeLink,
		InformationBlockTypeCallout,
		InformationBlockTypeCode,
		InformationBlockTypeTable,
		InformationBlockTypeDivider,
	}

	for index, blockType := range accepted {
		t.Run(string(blockType), func(t *testing.T) {
			input := json.RawMessage(fmt.Sprintf(
				`{"version":1,"blocks":[{"id":"block-%d","type":%q,"data":{}}]}`,
				index,
				blockType,
			))
			normalized, err := normalizeLearningItemMetadata(input)
			if err != nil {
				t.Fatalf("normalize %s: %v", blockType, err)
			}
			var metadata LearningItemMetadata
			if err := json.Unmarshal(normalized, &metadata); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if len(metadata.Blocks) != 1 ||
				metadata.Blocks[0].Type != blockType ||
				metadata.Blocks[0].Visibility.Mode != LearningItemVisibilityModeAll {
				t.Fatalf("normalized metadata = %+v", metadata)
			}
		})
	}

	unsupported := json.RawMessage(
		`{"version":1,"blocks":[{"id":"future","type":"FUTURE_BLOCK","data":{}}]}`,
	)
	if _, err := normalizeLearningItemMetadata(unsupported); !errors.Is(err, ErrLearningItemBlockTypeInvalid) {
		t.Fatalf("unsupported error = %v, want %v", err, ErrLearningItemBlockTypeInvalid)
	}
}

func TestLearningItemBackendSuiteCreateDefaultsAndUpdatePresence(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	title, metadata, publishState, err := validateCreateLearningItemParams(
		CreateLearningItemParams{
			CourseID:     courseID,
			CourseNodeID: nodeID,
			Title:        "  Defaulted title  ",
			ItemType:     LearningItemTypeArticle,
		},
	)
	if err != nil {
		t.Fatalf("validate create: %v", err)
	}
	if title != "Defaulted title" {
		t.Fatalf("title = %q", title)
	}
	if !bytes.Equal(metadata, defaultLearningItemMetadata) {
		t.Fatalf("metadata = %s, want %s", metadata, defaultLearningItemMetadata)
	}
	if publishState != LearningItemPublishStateDraft {
		t.Fatalf("publish_state = %q", publishState)
	}

	replacementInput := json.RawMessage(
		`{"version":1,"blocks":[{"id":"callout","type":"CALLOUT","data":{"text":"Notice"}}]}`,
	)
	replacementStored := []byte(
		`{"version":1,"blocks":[{"id":"callout","type":"CALLOUT","data":{"text":"Notice"},"visibility":{"mode":"ALL"}}]}`,
	)
	published := LearningItemPublishStatePublished
	record, err := buildLearningItemUpdateRecord(UpdateLearningItemParams{
		Description:  OptionalNullableString{Present: true, Null: true},
		Metadata:     OptionalJSONBytes{Present: true, Value: replacementInput},
		PublishState: &published,
	})
	if err != nil {
		t.Fatalf("build update: %v", err)
	}
	if len(record) != 4 {
		t.Fatalf("record = %#v", record)
	}
	if value, exists := record["title"]; exists || value != nil {
		t.Fatalf("omitted title present: %#v", value)
	}
	if value, exists := record["item_type"]; exists || value != nil {
		t.Fatalf("omitted item_type present: %#v", value)
	}
	if value, exists := record["description"]; !exists || value != nil {
		t.Fatalf("description = %#v, exists=%v", value, exists)
	}
	if value, exists := record["metadata"]; !exists || !bytes.Equal(value.([]byte), replacementStored) {
		t.Fatalf("metadata = %#v, exists=%v", value, exists)
	}
	if value, exists := record["publish_state"]; !exists || value != LearningItemPublishStatePublished {
		t.Fatalf("publish_state = %#v, exists=%v", value, exists)
	}
	if _, exists := record["updated_at"]; !exists {
		t.Fatalf("updated_at missing: %#v", record)
	}

	if _, err := buildLearningItemUpdateRecord(UpdateLearningItemParams{}); !errors.Is(err, ErrLearningItemUpdateRequired) {
		t.Fatalf("empty update error = %v", err)
	}
	if _, err := buildLearningItemUpdateRecord(UpdateLearningItemParams{
		Metadata: OptionalJSONBytes{Present: true, Null: true},
	}); !errors.Is(err, ErrLearningItemMetadataInvalid) {
		t.Fatalf("metadata null error = %v", err)
	}
}
