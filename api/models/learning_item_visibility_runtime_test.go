package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestLearningItemVisibilityRuntime_ALLRetained(t *testing.T) {
	item := learningItemWithBlocks(t, visibilityBlock("all", LearningItemVisibilityModeAll))
	got, err := ProjectLearningItemForLearner(item, LearnerVisibilityAccess{})
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	assertProjectedBlockIDs(t, got.Metadata, []string{"all"})
}

func TestLearningItemVisibilityRuntime_AuthenticatedRetainedWhenAuthenticated(t *testing.T) {
	item := learningItemWithBlocks(t, visibilityBlock("auth", LearningItemVisibilityModeAuthenticated))
	got, err := ProjectLearningItemForLearner(item, LearnerVisibilityAccess{Authenticated: true})
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	assertProjectedBlockIDs(t, got.Metadata, []string{"auth"})
}

func TestLearningItemVisibilityRuntime_AuthenticatedOmittedWhenNotAuthenticated(t *testing.T) {
	item := learningItemWithBlocks(t, visibilityBlock("auth", LearningItemVisibilityModeAuthenticated))
	got, err := ProjectLearningItemForLearner(item, LearnerVisibilityAccess{Authenticated: false})
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	assertProjectedBlockIDs(t, got.Metadata, []string{})
}

func TestLearningItemVisibilityRuntime_HiddenAndInstructorRemoved(t *testing.T) {
	item := learningItemWithBlocks(t,
		visibilityBlock("hidden", LearningItemVisibilityModeHidden),
		visibilityBlock("instructor", LearningItemVisibilityModeInstructor),
		visibilityBlock("all", LearningItemVisibilityModeAll),
	)
	got, err := ProjectLearningItemForLearner(item, AuthenticatedLearnerVisibilityAccess())
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	assertProjectedBlockIDs(t, got.Metadata, []string{"all"})
}

func TestLearningItemVisibilityRuntime_PremiumOmittedWhenUnauthorized(t *testing.T) {
	item := learningItemWithBlocks(t, visibilityBlock("premium", LearningItemVisibilityModePremium))
	got, err := ProjectLearningItemForLearner(item, AuthenticatedLearnerVisibilityAccess())
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	assertProjectedBlockIDs(t, got.Metadata, []string{})
}

func TestLearningItemVisibilityRuntime_PremiumRetainedWhenAuthorized(t *testing.T) {
	item := learningItemWithBlocks(t, visibilityBlock("premium", LearningItemVisibilityModePremium))
	got, err := ProjectLearningItemForLearner(item, LearnerVisibilityAccess{
		Authenticated:     true,
		PremiumAuthorized: true,
	})
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	assertProjectedBlockIDs(t, got.Metadata, []string{"premium"})
}

func TestLearningItemVisibilityRuntime_MixedDocumentPreservesOrder(t *testing.T) {
	item := learningItemWithBlocks(t,
		visibilityBlock("a", LearningItemVisibilityModeAll),
		visibilityBlock("b", LearningItemVisibilityModeHidden),
		visibilityBlock("c", LearningItemVisibilityModeAuthenticated),
		visibilityBlock("d", LearningItemVisibilityModePremium),
		visibilityBlock("e", LearningItemVisibilityModeInstructor),
		visibilityBlock("f", LearningItemVisibilityModeAll),
	)
	got, err := ProjectLearningItemForLearner(item, AuthenticatedLearnerVisibilityAccess())
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	assertProjectedBlockIDs(t, got.Metadata, []string{"a", "c", "f"})
}

func TestLearningItemVisibilityRuntime_EmptyBlocksPreserved(t *testing.T) {
	item := LearningItem{
		Metadata: json.RawMessage(`{"version":3,"blocks":[]}`),
	}
	got, err := ProjectLearningItemForLearner(item, AuthenticatedLearnerVisibilityAccess())
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	var decoded LearningItemMetadata
	if err := json.Unmarshal(got.Metadata, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Version != 3 {
		t.Fatalf("version = %d, want 3", decoded.Version)
	}
	if decoded.Blocks == nil {
		t.Fatal("blocks must be non-nil empty array")
	}
	if len(decoded.Blocks) != 0 {
		t.Fatalf("blocks len = %d", len(decoded.Blocks))
	}
	if bytes.Contains(got.Metadata, []byte(`null`)) && bytes.Contains(got.Metadata, []byte(`"blocks":null`)) {
		t.Fatalf("blocks serialized as null: %s", got.Metadata)
	}
}

func TestLearningItemVisibilityRuntime_AllHiddenYieldsNonNilEmptyBlocks(t *testing.T) {
	item := learningItemWithBlocks(t,
		visibilityBlock("h1", LearningItemVisibilityModeHidden),
		visibilityBlock("h2", LearningItemVisibilityModeInstructor),
	)
	item.Metadata = mustReplaceVersion(t, item.Metadata, 4)
	got, err := ProjectLearningItemForLearner(item, AuthenticatedLearnerVisibilityAccess())
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	var decoded LearningItemMetadata
	if err := json.Unmarshal(got.Metadata, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Version != 4 {
		t.Fatalf("version = %d, want 4", decoded.Version)
	}
	if decoded.Blocks == nil || len(decoded.Blocks) != 0 {
		t.Fatalf("blocks = %#v, want non-nil empty", decoded.Blocks)
	}
}

func TestLearningItemVisibilityRuntime_DeepCopyDoesNotMutateOriginal(t *testing.T) {
	originalData := json.RawMessage(`{"text":"secret"}`)
	item := LearningItem{
		Metadata: mustMarshalMetadata(t, LearningItemMetadata{
			Version: 1,
			Blocks: []LearningItemBlock{
				{
					ID:         "mutable",
					Type:       InformationBlockTypeText,
					Data:       originalData,
					Visibility: LearningItemBlockVisibility{Mode: LearningItemVisibilityModeAll},
				},
				{
					ID:         "hidden",
					Type:       InformationBlockTypeText,
					Data:       json.RawMessage(`{"text":"gone"}`),
					Visibility: LearningItemBlockVisibility{Mode: LearningItemVisibilityModeHidden},
				},
			},
		}),
	}
	originalMetadataCopy := append(json.RawMessage(nil), item.Metadata...)

	got, err := ProjectLearningItemForLearner(item, AuthenticatedLearnerVisibilityAccess())
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	if !bytes.Equal(item.Metadata, originalMetadataCopy) {
		t.Fatalf("original metadata mutated\nbefore=%s\nafter=%s", originalMetadataCopy, item.Metadata)
	}

	var decoded LearningItemMetadata
	if err := json.Unmarshal(got.Metadata, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Blocks) != 1 {
		t.Fatalf("blocks = %d", len(decoded.Blocks))
	}
	decoded.Blocks[0].Data[2] = 'X'
	if bytes.Equal(decoded.Blocks[0].Data, originalData) {
		t.Fatal("projected data shares backing with original input data")
	}
	if !bytes.Equal(originalData, json.RawMessage(`{"text":"secret"}`)) {
		t.Fatalf("original block data mutated: %s", originalData)
	}
}

func TestLearningItemVisibilityRuntime_UnknownModeOmitted(t *testing.T) {
	raw := json.RawMessage(`{"version":1,"blocks":[{"id":"known","type":"TEXT","data":{},"visibility":{"mode":"ALL"}},{"id":"weird","type":"TEXT","data":{},"visibility":{"mode":"PUBLIC"}}]}`)
	item := LearningItem{Metadata: raw}
	got, err := ProjectLearningItemForLearner(item, AuthenticatedLearnerVisibilityAccess())
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	assertProjectedBlockIDs(t, got.Metadata, []string{"known"})
}

func TestLearningItemVisibilityRuntime_MalformedMetadataWrappedError(t *testing.T) {
	item := LearningItem{Metadata: json.RawMessage(`{"version":1,"blocks":`)}
	_, err := ProjectLearningItemForLearner(item, AuthenticatedLearnerVisibilityAccess())
	if !errors.Is(err, ErrLearningItemMetadataInvalid) {
		t.Fatalf("error = %v, want wrapped %v", err, ErrLearningItemMetadataInvalid)
	}
	if !errors.Is(err, ErrLearningItemMetadataInvalid) {
		t.Fatalf("unwrap missing metadata invalid: %v", err)
	}
}

func TestLearningItemVisibilityRuntime_NullBlocksWrappedError(t *testing.T) {
	item := LearningItem{Metadata: json.RawMessage(`{"version":1,"blocks":null}`)}
	_, err := ProjectLearningItemForLearner(item, AuthenticatedLearnerVisibilityAccess())
	if !errors.Is(err, ErrLearningItemMetadataInvalid) {
		t.Fatalf("error = %v, want %v", err, ErrLearningItemMetadataInvalid)
	}
}

func TestLearningItemVisibilityRuntime_ProductionAccessDefaults(t *testing.T) {
	access := AuthenticatedLearnerVisibilityAccess()
	if !access.Authenticated {
		t.Fatal("production learner must be Authenticated=true")
	}
	if access.PremiumAuthorized {
		t.Fatal("production PremiumAuthorized must stay false until a real premium source exists")
	}
}

func TestLearningItemVisibilityRuntime_ListProjectionAtomicFailure(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	goodID := uuid.MustParse("019c02a0-1111-7000-8000-000000000201")
	badID := uuid.MustParse("019c02a0-1111-7000-8000-000000000202")
	now := time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)
	goodMeta := mustMarshalMetadata(t, LearningItemMetadata{
		Version: 1,
		Blocks:  []LearningItemBlock{visibilityBlock("ok", LearningItemVisibilityModeAll)},
	})
	badMeta := []byte(`{"version":1,"blocks":`)

	model, mock := newLearningItemModelTest(t)
	expectLearningItemNodeLookup(mock, courseID, nodeID, true)
	rows := learningItemRowsWithPublishState(
		goodID, courseID, nodeID, "Good", LearningItemTypeArticle,
		nil, goodMeta, 0, LearningItemPublishStatePublished,
	)
	rows.AddRow(
		badID, courseID, nodeID, "Bad", LearningItemTypeArticle,
		nil, badMeta, 1, LearningItemPublishStatePublished, now, now,
	)
	mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
		WithArgs(courseID, nodeID, LearningItemPublishStatePublished).
		WillReturnRows(rows)

	items, err := model.ListPublishedLearningItemsByNode(courseID, nodeID)
	if err == nil {
		t.Fatal("expected atomic projection failure")
	}
	if !errors.Is(err, ErrLearningItemMetadataInvalid) {
		t.Fatalf("error = %v, want metadata invalid", err)
	}
	if items != nil {
		t.Fatalf("partial list returned: %#v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLearningItemVisibilityRuntime_AdminRetrievalUnaffected(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000203")
	now := time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)
	metadata := mustMarshalMetadata(t, LearningItemMetadata{
		Version: 1,
		Blocks: []LearningItemBlock{
			visibilityBlock("hidden", LearningItemVisibilityModeHidden),
			visibilityBlock("instructor", LearningItemVisibilityModeInstructor),
			visibilityBlock("premium", LearningItemVisibilityModePremium),
		},
	})

	model, mock := newLearningItemModelTest(t)
	mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
		WithArgs(courseID, nodeID, itemID, uint(1)).
		WillReturnRows(learningItemRowsWithPublishState(
			itemID, courseID, nodeID, "Admin Full", LearningItemTypeArticle,
			nil, metadata, 0, LearningItemPublishStatePublished,
		))

	item, err := model.GetLearningItemByID(courseID, nodeID, itemID)
	if err != nil {
		t.Fatalf("admin get: %v", err)
	}
	assertProjectedBlockIDs(t, item.Metadata, []string{"hidden", "instructor", "premium"})
	_ = now
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLearningItemVisibilityRuntime_SQLPublishFilterStill404ForDrafts(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	draftID := uuid.MustParse("019c02a0-1111-7000-8000-000000000204")

	model, mock := newLearningItemModelTest(t)
	mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
		WithArgs(courseID, nodeID, draftID, LearningItemPublishStatePublished, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "course_id", "course_node_id", "title", "item_type",
			"description", "metadata", "position", "publish_state", "created_at", "updated_at",
		}))

	_, err := model.GetPublishedLearningItemByID(courseID, nodeID, draftID)
	if !errors.Is(err, ErrLearningItemNotFound) {
		t.Fatalf("error = %v, want not found", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLearningItemVisibilityRuntime_PublishedGetProjectsBlocks(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000205")
	metadata := mustMarshalMetadata(t, LearningItemMetadata{
		Version: 2,
		Blocks: []LearningItemBlock{
			visibilityBlock("keep", LearningItemVisibilityModeAuthenticated),
			visibilityBlock("drop", LearningItemVisibilityModeHidden),
			visibilityBlock("premium", LearningItemVisibilityModePremium),
		},
	})

	model, mock := newLearningItemModelTest(t)
	mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state`).
		WithArgs(courseID, nodeID, itemID, LearningItemPublishStatePublished, uint(1)).
		WillReturnRows(learningItemRowsWithPublishState(
			itemID, courseID, nodeID, "Projected", LearningItemTypeArticle,
			nil, metadata, 0, LearningItemPublishStatePublished,
		))

	item, err := model.GetPublishedLearningItemByID(courseID, nodeID, itemID)
	if err != nil {
		t.Fatalf("get published: %v", err)
	}
	assertProjectedBlockIDs(t, item.Metadata, []string{"keep"})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func visibilityBlock(id string, mode LearningItemVisibilityMode) LearningItemBlock {
	return LearningItemBlock{
		ID:         id,
		Type:       InformationBlockTypeText,
		Data:       json.RawMessage(`{"text":"` + id + `"}`),
		Visibility: LearningItemBlockVisibility{Mode: mode},
	}
}

func learningItemWithBlocks(t *testing.T, blocks ...LearningItemBlock) LearningItem {
	t.Helper()
	return LearningItem{
		Metadata: mustMarshalMetadata(t, LearningItemMetadata{
			Version: 1,
			Blocks:  blocks,
		}),
	}
}

func mustMarshalMetadata(t *testing.T, meta LearningItemMetadata) json.RawMessage {
	t.Helper()
	if meta.Blocks == nil {
		meta.Blocks = []LearningItemBlock{}
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal metadata: %v", err)
	}
	return raw
}

func mustReplaceVersion(t *testing.T, metadata json.RawMessage, version int) json.RawMessage {
	t.Helper()
	var decoded LearningItemMetadata
	if err := json.Unmarshal(metadata, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	decoded.Version = version
	return mustMarshalMetadata(t, decoded)
}

func assertProjectedBlockIDs(t *testing.T, metadata json.RawMessage, want []string) {
	t.Helper()
	var decoded LearningItemMetadata
	if err := json.Unmarshal(metadata, &decoded); err != nil {
		t.Fatalf("unmarshal projected metadata: %v", err)
	}
	if decoded.Blocks == nil {
		t.Fatal("projected blocks is nil")
	}
	if len(decoded.Blocks) != len(want) {
		t.Fatalf("block count = %d (%v), want %d (%v)", len(decoded.Blocks), blockIDs(decoded.Blocks), len(want), want)
	}
	for i, id := range want {
		if decoded.Blocks[i].ID != id {
			t.Fatalf("block[%d]=%s, want %s (got %v)", i, decoded.Blocks[i].ID, id, blockIDs(decoded.Blocks))
		}
	}
}

func blockIDs(blocks []LearningItemBlock) []string {
	ids := make([]string, 0, len(blocks))
	for _, block := range blocks {
		ids = append(ids, block.ID)
	}
	return ids
}
