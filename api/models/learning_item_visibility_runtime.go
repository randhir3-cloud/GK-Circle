package models

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// LearnerVisibilityAccess is the learner access context used for runtime
// visibility projection. It does not invent Course premium entitlements.
type LearnerVisibilityAccess struct {
	Authenticated     bool
	PremiumAuthorized bool
}

// AuthenticatedLearnerVisibilityAccess is the production learner delivery
// context: authenticated Kratos identity with no Course premium source yet.
func AuthenticatedLearnerVisibilityAccess() LearnerVisibilityAccess {
	return LearnerVisibilityAccess{
		Authenticated:     true,
		PremiumAuthorized: false,
	}
}

// ProjectLearningItemForLearner returns a deep-copied LearningItem whose
// metadata blocks are filtered for learner delivery. The caller's original
// metadata buffer is never mutated.
func ProjectLearningItemForLearner(item LearningItem, access LearnerVisibilityAccess) (LearningItem, error) {
	out := item

	trimmed := bytes.TrimSpace(item.Metadata)
	if len(trimmed) == 0 {
		out.Metadata = append(json.RawMessage(nil), defaultLearningItemMetadata...)
		return out, nil
	}

	var envelope LearningItemMetadata
	if err := json.Unmarshal(trimmed, &envelope); err != nil {
		return LearningItem{}, newLearningItemSemanticError(
			ErrLearningItemMetadataInvalid,
			fmt.Errorf("project learning item metadata: %w", err),
		)
	}
	if envelope.Version < 1 {
		return LearningItem{}, newLearningItemSemanticError(
			ErrLearningItemMetadataVersionInvalid,
			ErrLearningItemMetadataVersionInvalid,
		)
	}
	if envelope.Blocks == nil {
		return LearningItem{}, newLearningItemSemanticError(
			ErrLearningItemMetadataInvalid,
			ErrLearningItemMetadataInvalid,
		)
	}

	filtered := filterBlocksForLearner(envelope.Blocks, access)
	if filtered == nil {
		filtered = []LearningItemBlock{}
	}

	encoded, err := json.Marshal(LearningItemMetadata{
		Version: envelope.Version,
		Blocks:  filtered,
	})
	if err != nil {
		return LearningItem{}, newLearningItemSemanticError(
			ErrLearningItemMetadataInvalid,
			fmt.Errorf("marshal projected learning item metadata: %w", err),
		)
	}

	out.Metadata = json.RawMessage(encoded)
	return out, nil
}

// filterBlocksForLearner applies stable visibility rules and deep-copies retained
// blocks. Unknown modes are omitted without placeholders.
func filterBlocksForLearner(blocks []LearningItemBlock, access LearnerVisibilityAccess) []LearningItemBlock {
	out := make([]LearningItemBlock, 0, len(blocks))
	for _, block := range blocks {
		if !learnerMaySeeVisibilityMode(block.Visibility.Mode, access) {
			continue
		}
		out = append(out, LearningItemBlock{
			ID:   block.ID,
			Type: block.Type,
			Data: append(json.RawMessage(nil), block.Data...),
			Visibility: LearningItemBlockVisibility{
				Mode: block.Visibility.Mode,
			},
		})
	}
	return out
}

func learnerMaySeeVisibilityMode(mode LearningItemVisibilityMode, access LearnerVisibilityAccess) bool {
	switch mode {
	case LearningItemVisibilityModeAll:
		return true
	case LearningItemVisibilityModeAuthenticated:
		return access.Authenticated
	case LearningItemVisibilityModePremium:
		return access.PremiumAuthorized
	case LearningItemVisibilityModeInstructor, LearningItemVisibilityModeHidden:
		return false
	default:
		// Unknown modes are omitted and projection continues.
		return false
	}
}
