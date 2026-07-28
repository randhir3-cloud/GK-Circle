package models

import (
	"bytes"
	"encoding/json"
	"strings"
)

// Canonical empty LearningItem metadata document (COURSE-P2-T03).
var defaultLearningItemMetadata = json.RawMessage(`{"version":1,"blocks":[]}`)

type InformationBlockType string

const (
	InformationBlockTypeText    InformationBlockType = "TEXT"
	InformationBlockTypeHeading InformationBlockType = "HEADING"
	InformationBlockTypeImage   InformationBlockType = "IMAGE"
	InformationBlockTypeVideo   InformationBlockType = "VIDEO"
	InformationBlockTypePDF     InformationBlockType = "PDF"
	InformationBlockTypeLink    InformationBlockType = "LINK"
	InformationBlockTypeCallout InformationBlockType = "CALLOUT"
	InformationBlockTypeCode    InformationBlockType = "CODE"
	InformationBlockTypeTable   InformationBlockType = "TABLE"
	InformationBlockTypeDivider InformationBlockType = "DIVIDER"
)

// LearningItemMetadata is the structural envelope stored in learning_items.metadata.
// Per-type content-field contracts are intentionally deferred.
type LearningItemMetadata struct {
	Version int                 `json:"version"`
	Blocks  []LearningItemBlock `json:"blocks"`
}

// LearningItemBlock is one ordered information block inside LearningItem metadata.
type LearningItemBlock struct {
	ID         string                      `json:"id"`
	Type       InformationBlockType        `json:"type"`
	Data       json.RawMessage             `json:"data"`
	Visibility LearningItemBlockVisibility `json:"visibility"`
}

// learningItemBlockInput preserves raw visibility so omitted vs null can be distinguished.
type learningItemBlockInput struct {
	ID         string               `json:"id"`
	Type       InformationBlockType `json:"type"`
	Data       json.RawMessage      `json:"data"`
	Visibility json.RawMessage      `json:"visibility"`
}

func normalizeLearningItemMetadata(value json.RawMessage) (json.RawMessage, error) {
	if len(value) == 0 {
		return append(json.RawMessage(nil), defaultLearningItemMetadata...), nil
	}
	trimmed := bytes.TrimSpace(value)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return nil, ErrLearningItemMetadataInvalid
	}
	if !json.Valid(trimmed) {
		return nil, ErrLearningItemMetadataInvalid
	}

	var root interface{}
	if err := json.Unmarshal(trimmed, &root); err != nil {
		return nil, ErrLearningItemMetadataInvalid
	}
	rootObject, ok := root.(map[string]interface{})
	if !ok {
		return nil, ErrLearningItemMetadataInvalid
	}
	if _, hasBlocks := rootObject["blocks"]; !hasBlocks {
		return nil, ErrLearningItemMetadataInvalid
	}

	var envelope struct {
		Version int                      `json:"version"`
		Blocks  []learningItemBlockInput `json:"blocks"`
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.UseNumber()
	if err := decoder.Decode(&envelope); err != nil {
		return nil, ErrLearningItemMetadataInvalid
	}
	if envelope.Version < 1 {
		return nil, ErrLearningItemMetadataVersionInvalid
	}
	if envelope.Blocks == nil {
		return nil, ErrLearningItemMetadataInvalid
	}

	seenIDs := make(map[string]struct{}, len(envelope.Blocks))
	normalizedBlocks := make([]LearningItemBlock, 0, len(envelope.Blocks))
	for _, block := range envelope.Blocks {
		id := strings.TrimSpace(block.ID)
		if id == "" {
			return nil, ErrLearningItemMetadataInvalid
		}
		if _, exists := seenIDs[id]; exists {
			return nil, ErrLearningItemBlockDuplicate
		}
		seenIDs[id] = struct{}{}
		if !isValidInformationBlockType(block.Type) {
			return nil, ErrLearningItemBlockTypeInvalid
		}
		data, err := validateLearningItemBlockData(block.Data)
		if err != nil {
			return nil, err
		}
		visibility, err := normalizeLearningItemBlockVisibility(block.Visibility)
		if err != nil {
			return nil, err
		}
		normalizedBlocks = append(normalizedBlocks, LearningItemBlock{
			ID:         id,
			Type:       block.Type,
			Data:       data,
			Visibility: visibility,
		})
	}

	canonical, err := json.Marshal(LearningItemMetadata{
		Version: envelope.Version,
		Blocks:  normalizedBlocks,
	})
	if err != nil {
		return nil, ErrLearningItemMetadataInvalid
	}
	return json.RawMessage(canonical), nil
}

func validateLearningItemBlockData(value json.RawMessage) (json.RawMessage, error) {
	if len(value) == 0 {
		return nil, ErrLearningItemMetadataInvalid
	}
	trimmed := bytes.TrimSpace(value)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return nil, ErrLearningItemMetadataInvalid
	}
	if !json.Valid(trimmed) {
		return nil, ErrLearningItemMetadataInvalid
	}
	var probe interface{}
	if err := json.Unmarshal(trimmed, &probe); err != nil {
		return nil, ErrLearningItemMetadataInvalid
	}
	if _, ok := probe.(map[string]interface{}); !ok {
		return nil, ErrLearningItemMetadataInvalid
	}
	if err := validateLearningItemPlaceholdersInData(trimmed); err != nil {
		return nil, err
	}
	// Preserve original object bytes; do not remashal to {} or reorder keys.
	return json.RawMessage(trimmed), nil
}

func isValidInformationBlockType(blockType InformationBlockType) bool {
	switch blockType {
	case InformationBlockTypeText,
		InformationBlockTypeHeading,
		InformationBlockTypeImage,
		InformationBlockTypeVideo,
		InformationBlockTypePDF,
		InformationBlockTypeLink,
		InformationBlockTypeCallout,
		InformationBlockTypeCode,
		InformationBlockTypeTable,
		InformationBlockTypeDivider:
		return true
	default:
		return false
	}
}
