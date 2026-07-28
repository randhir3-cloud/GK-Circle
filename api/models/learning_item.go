package models

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

const LearningItemsTable = "learning_items"

const learningItemsNodePositionConstraint = "learning_items_node_position_unique"

type LearningItemType string

const (
	LearningItemTypeArticle LearningItemType = "ARTICLE"
	LearningItemTypeVideo   LearningItemType = "VIDEO"
	LearningItemTypePDF     LearningItemType = "PDF"
	LearningItemTypeLink    LearningItemType = "LINK"
	LearningItemTypeQuizRef LearningItemType = "QUIZ_REFERENCE"
)

type LearningItemPublishState string

const (
	LearningItemPublishStateDraft     LearningItemPublishState = "DRAFT"
	LearningItemPublishStatePublished LearningItemPublishState = "PUBLISHED"
)

var (
	ErrLearningItemNotFound               = errors.New("learning item not found")
	ErrLearningItemConflict               = errors.New("learning item conflict")
	ErrLearningItemNodeNotFound           = errors.New("learning item course node not found")
	ErrLearningItemCrossCourse            = errors.New("learning item course node belongs to another course")
	ErrLearningItemTitleRequired          = errors.New("learning item title is required")
	ErrLearningItemTypeInvalid            = errors.New("learning item type is invalid")
	ErrLearningItemPositionInvalid        = errors.New("learning item position is invalid")
	ErrLearningItemUpdateRequired         = errors.New("learning item update requires at least one field")
	ErrLearningItemMetadataInvalid        = errors.New("learning item metadata is invalid")
	ErrLearningItemBlockDuplicate         = errors.New("learning item block id is duplicate")
	ErrLearningItemBlockTypeInvalid       = errors.New("learning item block type is invalid")
	ErrLearningItemMetadataVersionInvalid = errors.New("learning item metadata version is invalid")
	ErrLearningItemPlaceholderInvalid     = errors.New("learning item placeholder is invalid")
	ErrLearningItemPlaceholderSyntax      = errors.New("learning item placeholder syntax is invalid")
	ErrLearningItemVisibilityInvalid      = errors.New("learning item visibility is invalid")
	ErrLearningItemVisibilityModeInvalid  = errors.New("learning item visibility mode is invalid")
	ErrLearningItemPublishStateInvalid    = errors.New("learning item publish state is invalid")
	ErrLearningItemReorderDuplicate       = errors.New("learning item reorder duplicate ID")
	ErrLearningItemReorderMismatch        = errors.New("learning item reorder mismatched ID set")
	ErrLearningItemReorderConflict        = errors.New("learning item reorder concurrency conflict")
	ErrLearningItemMoveDuplicate          = errors.New("learning item move duplicate ID")
	ErrLearningItemMoveMismatch           = errors.New("learning item move mismatched ID set")
	ErrLearningItemMoveSameNode           = errors.New("learning item move source and destination are the same")
	ErrLearningItemMoveConflict           = errors.New("learning item move concurrency conflict")
	ErrLearningItemQuizRequired           = errors.New("learning item quiz reference is required")
	ErrLearningItemQuizForbidden          = errors.New("learning item quiz reference is forbidden")
	ErrLearningItemQuizNotFound           = errors.New("learning item quiz reference was not found")
	ErrLearningItemPersistence            = errors.New("learning item persistence failure")
)

var learningItemColumns = []interface{}{
	"id",
	"course_id",
	"course_node_id",
	"title",
	"item_type",
	"description",
	"metadata",
	"position",
	"publish_state",
	"created_at",
	"updated_at",
}

// LearningItem is a CourseNode-scoped ordered educational unit.
// Learner published reads are available via GetPublished/ListPublished helpers.
// Sibling reorder is available via ReorderLearningItems (COURSE-P2-T10).
// Cross-node move within a Course is available via MoveLearningItems (COURSE-P2-T11).
// Node-local previous/next resolution is available via GetAdjacentLearningItems (COURSE-P2-T21).
type LearningItem struct {
	ID           uuid.UUID                `json:"id" db:"id"`
	CourseID     uuid.UUID                `json:"course_id" db:"course_id"`
	CourseNodeID uuid.UUID                `json:"course_node_id" db:"course_node_id"`
	Title        string                   `json:"title" db:"title"`
	ItemType     LearningItemType         `json:"item_type" db:"item_type"`
	Description  sql.NullString           `json:"description,omitempty" db:"description"`
	Metadata     json.RawMessage          `json:"metadata" db:"metadata"`
	Position     int                      `json:"position" db:"position"`
	PublishState LearningItemPublishState `json:"publish_state" db:"publish_state"`
	CreatedAt    time.Time                `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at" db:"updated_at"`
}

type CreateLearningItemParams struct {
	CourseID     uuid.UUID
	CourseNodeID uuid.UUID
	Title        string
	ItemType     LearningItemType
	Description  string
	Metadata     json.RawMessage
	QuizID       *uuid.UUID
	ActorID      string
	PublishState LearningItemPublishState
}

// OptionalJSONBytes carries presence for partial JSON updates.
// Present=false means omitted; Present+Null is invalid for LearningItem metadata.
type OptionalJSONBytes struct {
	Present bool
	Null    bool
	Value   json.RawMessage
}

type OptionalNullableUUID struct {
	Present bool
	Null    bool
	Value   uuid.UUID
}

type UpdateLearningItemParams struct {
	Title        *string
	ItemType     *LearningItemType
	Description  OptionalNullableString
	Metadata     OptionalJSONBytes
	QuizID       OptionalNullableUUID
	ActorID      string
	PublishState *LearningItemPublishState
}

// ReorderLearningItemsResult is the frozen reorder outcome for admin APIs.
type ReorderLearningItemsResult struct {
	CourseNodeID      uuid.UUID
	LearningItemCount int
	PositionsUpdated  int
	Noop              bool
}

// MoveLearningItemsResult is the frozen move outcome for admin APIs.
type MoveLearningItemsResult struct {
	SourceNodeID         uuid.UUID
	DestinationNodeID    uuid.UUID
	ItemsMoved           int
	SourceItemCount      int
	DestinationItemCount int
	Noop                 bool
}

// LearningItemAdjacentResult holds node-local previous/next siblings.
// Nil Previous and/or Next means a chain boundary (not an error).
type LearningItemAdjacentResult struct {
	Previous *LearningItem
	Next     *LearningItem
}

type LearningItemModel struct {
	db      *goqu.Database
	newUUID func() (uuid.UUID, error)
}

func InitLearningItemModel(goquDB *goqu.Database) *LearningItemModel {
	return &LearningItemModel{
		db:      goquDB,
		newUUID: uuid.NewUUID,
	}
}

func (model *LearningItemModel) CreateLearningItem(params CreateLearningItemParams) (LearningItem, error) {
	var item LearningItem

	title, metadata, publishState, err := validateCreateLearningItemParams(params)
	if err != nil {
		return item, err
	}

	transaction, err := model.db.Begin()
	if err != nil {
		return item, newLearningItemPersistenceError(err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = transaction.Rollback()
		}
	}()

	if err := ensureLearningItemNodeInCourse(transaction, params.CourseID, params.CourseNodeID, true); err != nil {
		return item, err
	}
	if params.QuizID != nil {
		if err := ensureManageableLearningItemQuiz(transaction, *params.QuizID, params.ActorID); err != nil {
			return item, err
		}
	}

	var maxPosition sql.NullInt64
	found, err := transaction.From(LearningItemsTable).
		Select(goqu.MAX("position")).
		Where(goqu.Ex{"course_node_id": params.CourseNodeID}).
		Prepared(true).
		ScanVal(&maxPosition)
	if err != nil {
		return item, newLearningItemPersistenceError(err)
	}
	position := 0
	if found && maxPosition.Valid {
		if maxPosition.Int64 >= int64(^uint(0)>>1) {
			return item, ErrLearningItemPositionInvalid
		}
		position = int(maxPosition.Int64) + 1
	}

	itemID, err := model.newUUID()
	if err != nil {
		return item, newLearningItemPersistenceError(err)
	}

	createRecord := goqu.Record{
		"id":             itemID,
		"course_id":      params.CourseID,
		"course_node_id": params.CourseNodeID,
		"title":          title,
		"item_type":      params.ItemType,
		"description":    nullableTrimmedString(params.Description),
		"metadata":       []byte(metadata),
		"position":       position,
		"publish_state":  publishState,
	}
	if params.QuizID != nil {
		createRecord["quiz_id"] = *params.QuizID
	}
	found, err = transaction.Insert(LearningItemsTable).
		Rows(createRecord).
		Returning(learningItemColumns...).
		Prepared(true).
		Executor().
		ScanStruct(&item)
	if err != nil {
		return item, mapLearningItemWriteError(err)
	}
	if !found {
		return item, newLearningItemPersistenceError(sql.ErrNoRows)
	}
	if err := transaction.Commit(); err != nil {
		return item, newLearningItemPersistenceError(err)
	}
	committed = true
	return item, nil
}

func (model *LearningItemModel) GetLearningItemByID(courseID, nodeID, itemID uuid.UUID) (LearningItem, error) {
	return model.getLearningItem(courseID, nodeID, itemID, false)
}

// GetAdjacentLearningItems resolves the previous and next LearningItems for currentItemID
// within the same CourseNode only, using deterministic (position, id) sibling order.
// Chain ends return nil pointers (not errors). Publish filtering is intentionally omitted
// (COURSE-P2-T22 owns learner previous/next with draft skipping).
func (model *LearningItemModel) GetAdjacentLearningItems(
	courseID, nodeID, currentItemID uuid.UUID,
) (LearningItemAdjacentResult, error) {
	var result LearningItemAdjacentResult
	if err := validateGetAdjacentLearningItemsParams(courseID, nodeID, currentItemID); err != nil {
		return result, err
	}
	if err := ensureLearningItemNodeInCourse(model.db, courseID, nodeID, false); err != nil {
		return result, err
	}

	current, err := model.getLearningItem(courseID, nodeID, currentItemID, false)
	if err != nil {
		return result, err
	}

	previous, err := model.findAdjacentLearningItem(courseID, nodeID, current, false, false)
	if err != nil {
		return LearningItemAdjacentResult{}, err
	}
	next, err := model.findAdjacentLearningItem(courseID, nodeID, current, true, false)
	if err != nil {
		return LearningItemAdjacentResult{}, err
	}
	result.Previous = previous
	result.Next = next
	return result, nil
}

// GetAdjacentPublishedLearningItems resolves the previous and next published LearningItems
// for currentItemID within the same CourseNode using deterministic (position, id) sibling order.
// If the current item is a draft or missing, ErrLearningItemNotFound is returned.
func (model *LearningItemModel) GetAdjacentPublishedLearningItems(
	courseID, nodeID, currentItemID uuid.UUID,
) (LearningItemAdjacentResult, error) {
	var result LearningItemAdjacentResult
	if err := validateGetAdjacentLearningItemsParams(courseID, nodeID, currentItemID); err != nil {
		return result, err
	}
	if err := ensureLearningItemNodeInCourse(model.db, courseID, nodeID, false); err != nil {
		return result, err
	}

	// Fetch current item only if published.
	current, err := model.getLearningItem(courseID, nodeID, currentItemID, true)
	if err != nil {
		return result, err
	}

	previous, err := model.findAdjacentLearningItem(courseID, nodeID, current, false, true)
	if err != nil {
		return LearningItemAdjacentResult{}, err
	}
	next, err := model.findAdjacentLearningItem(courseID, nodeID, current, true, true)
	if err != nil {
		return LearningItemAdjacentResult{}, err
	}
	result.Previous = previous
	result.Next = next
	return result, nil
}

// GetPublishedLearningItemByID returns a LearningItem only when publish_state is PUBLISHED.
// Missing and draft items both map to ErrLearningItemNotFound (no draft discovery).
// After the SQL publish filter, metadata blocks are projected for learner visibility.
func (model *LearningItemModel) GetPublishedLearningItemByID(courseID, nodeID, itemID uuid.UUID) (LearningItem, error) {
	item, err := model.getLearningItem(courseID, nodeID, itemID, true)
	if err != nil {
		return item, err
	}
	return ProjectLearningItemForLearner(item, AuthenticatedLearnerVisibilityAccess())
}

func (model *LearningItemModel) getLearningItem(courseID, nodeID, itemID uuid.UUID, publishedOnly bool) (LearningItem, error) {
	var item LearningItem

	where := goqu.Ex{
		"id":             itemID,
		"course_id":      courseID,
		"course_node_id": nodeID,
	}
	if publishedOnly {
		where["publish_state"] = LearningItemPublishStatePublished
	}

	found, err := model.db.From(LearningItemsTable).
		Select(learningItemColumns...).
		Where(where).
		Limit(1).
		Prepared(true).
		ScanStruct(&item)
	if err != nil {
		return item, newLearningItemPersistenceError(err)
	}
	if !found {
		return item, ErrLearningItemNotFound
	}
	return item, nil
}

func (model *LearningItemModel) ListLearningItemsByNode(courseID, nodeID uuid.UUID) ([]LearningItem, error) {
	return model.listLearningItems(courseID, nodeID, false)
}

// ListPublishedLearningItemsByNode returns only PUBLISHED items for a node.
// Publish filtering is enforced in SQL; callers must not post-filter.
// After the SQL publish filter, each item is projected for learner visibility.
// Projection failure is atomic: no partially projected list is returned.
func (model *LearningItemModel) ListPublishedLearningItemsByNode(courseID, nodeID uuid.UUID) ([]LearningItem, error) {
	items, err := model.listLearningItems(courseID, nodeID, true)
	if err != nil {
		return nil, err
	}

	access := AuthenticatedLearnerVisibilityAccess()
	projected := make([]LearningItem, 0, len(items))
	for _, item := range items {
		out, projectErr := ProjectLearningItemForLearner(item, access)
		if projectErr != nil {
			return nil, projectErr
		}
		projected = append(projected, out)
	}
	return projected, nil
}

func (model *LearningItemModel) listLearningItems(courseID, nodeID uuid.UUID, publishedOnly bool) ([]LearningItem, error) {
	if err := ensureLearningItemNodeInCourse(model.db, courseID, nodeID, false); err != nil {
		return nil, err
	}

	where := goqu.Ex{
		"course_id":      courseID,
		"course_node_id": nodeID,
	}
	if publishedOnly {
		where["publish_state"] = LearningItemPublishStatePublished
	}

	items := make([]LearningItem, 0)
	err := model.db.From(LearningItemsTable).
		Select(learningItemColumns...).
		Where(where).
		Order(goqu.I("position").Asc(), goqu.I("id").Asc()).
		Prepared(true).
		ScanStructs(&items)
	if err != nil {
		return nil, newLearningItemPersistenceError(err)
	}
	return items, nil
}

func (model *LearningItemModel) UpdateLearningItem(courseID, nodeID, itemID uuid.UUID, params UpdateLearningItemParams) (LearningItem, error) {
	var item LearningItem

	record, err := buildLearningItemUpdateRecord(params)
	if err != nil {
		return item, err
	}
	if params.QuizID.Present && !params.QuizID.Null {
		if err := ensureManageableLearningItemQuiz(model.db, params.QuizID.Value, params.ActorID); err != nil {
			return item, err
		}
	}

	found, err := model.db.Update(LearningItemsTable).
		Set(record).
		Where(goqu.Ex{
			"id":             itemID,
			"course_id":      courseID,
			"course_node_id": nodeID,
		}).
		Returning(learningItemColumns...).
		Prepared(true).
		Executor().
		ScanStruct(&item)
	if err != nil {
		return item, mapLearningItemWriteError(err)
	}
	if !found {
		return item, ErrLearningItemNotFound
	}
	return item, nil
}

// ReorderLearningItems replaces sibling order under a CourseNode with the exact
// ordered ID list. Positions are normalized to 0..n-1 using a two-phase temporary
// staging update to avoid unique-index collisions. Already-canonical orders are
// committed as a noop without UPDATE statements.
func (model *LearningItemModel) ReorderLearningItems(
	courseID, nodeID uuid.UUID,
	orderedItemIDs []uuid.UUID,
) (ReorderLearningItemsResult, error) {
	result := ReorderLearningItemsResult{CourseNodeID: nodeID}
	if err := validateReorderLearningItemsParams(courseID, nodeID, orderedItemIDs); err != nil {
		return result, err
	}

	transaction, err := model.db.Begin()
	if err != nil {
		return result, newLearningItemPersistenceError(err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = transaction.Rollback()
		}
	}()

	if err := lockCourseForLearningItemReorder(transaction, courseID); err != nil {
		return result, err
	}
	if err := ensureLearningItemNodeInCourse(transaction, courseID, nodeID, true); err != nil {
		return result, err
	}

	lockedSiblings, err := lockLearningItemSiblings(transaction, courseID, nodeID)
	if err != nil {
		return result, err
	}
	result.LearningItemCount = len(lockedSiblings)

	if len(lockedSiblings) == 0 {
		if len(orderedItemIDs) != 0 {
			return result, ErrLearningItemReorderMismatch
		}
		result.Noop = true
		result.PositionsUpdated = 0
		if err := transaction.Commit(); err != nil {
			return result, newLearningItemPersistenceError(err)
		}
		committed = true
		return result, nil
	}
	if err := ensureLearningItemReorderSetMatches(lockedSiblings, orderedItemIDs); err != nil {
		return result, err
	}

	finalPositionByID := make(map[uuid.UUID]int, len(orderedItemIDs))
	for index, itemID := range orderedItemIDs {
		finalPositionByID[itemID] = index
	}
	alreadyCanonical := true
	for _, sibling := range lockedSiblings {
		if sibling.Position != finalPositionByID[sibling.ID] {
			alreadyCanonical = false
			break
		}
	}
	if alreadyCanonical {
		result.Noop = true
		result.PositionsUpdated = 0
		if err := transaction.Commit(); err != nil {
			return result, newLearningItemPersistenceError(err)
		}
		committed = true
		return result, nil
	}

	maxPosition := 0
	for _, sibling := range lockedSiblings {
		if sibling.Position > maxPosition {
			maxPosition = sibling.Position
		}
	}
	siblingCount64 := int64(len(lockedSiblings))
	maxPosition64 := int64(maxPosition)
	maxExisting64 := maxPosition64
	if canonicalMax := siblingCount64 - 1; canonicalMax > maxExisting64 {
		maxExisting64 = canonicalMax
	}
	temporaryBase64 := maxExisting64 + siblingCount64 + 1
	highestTemporary64 := temporaryBase64 + siblingCount64 - 1
	if temporaryBase64 > math.MaxInt32 || highestTemporary64 > math.MaxInt32 {
		return result, ErrLearningItemPositionInvalid
	}
	temporaryBase := int(temporaryBase64)

	expectedIDs := make(map[uuid.UUID]struct{}, len(lockedSiblings))
	for _, sibling := range lockedSiblings {
		expectedIDs[sibling.ID] = struct{}{}
	}

	phase1IDs := make([]uuid.UUID, 0, len(lockedSiblings))
	for index, sibling := range lockedSiblings {
		updatedID, err := updateLearningItemSiblingPosition(
			transaction,
			courseID,
			nodeID,
			sibling.ID,
			temporaryBase+index,
			false,
		)
		if err != nil {
			return result, err
		}
		phase1IDs = append(phase1IDs, updatedID)
	}
	if err := verifyLearningItemReorderUpdatedIDs(expectedIDs, phase1IDs); err != nil {
		return result, err
	}

	phase2IDs := make([]uuid.UUID, 0, len(lockedSiblings))
	positionsUpdated := 0
	for _, sibling := range lockedSiblings {
		finalPosition := finalPositionByID[sibling.ID]
		updatedID, err := updateLearningItemSiblingPosition(
			transaction,
			courseID,
			nodeID,
			sibling.ID,
			finalPosition,
			true,
		)
		if err != nil {
			return result, err
		}
		phase2IDs = append(phase2IDs, updatedID)
		if sibling.Position != finalPosition {
			positionsUpdated++
		}
	}
	if err := verifyLearningItemReorderUpdatedIDs(expectedIDs, phase2IDs); err != nil {
		return result, err
	}

	if err := transaction.Commit(); err != nil {
		return result, newLearningItemPersistenceError(err)
	}
	committed = true
	result.Noop = false
	result.PositionsUpdated = positionsUpdated
	return result, nil
}

// MoveLearningItems moves an ordered subset of LearningItems from sourceNodeID to
// destinationNodeID within the same Course. Both sibling lists are compacted to
// canonical 0..n-1. Empty orderedItemIDs is a noop after course/node validation.
func (model *LearningItemModel) MoveLearningItems(
	courseID, sourceNodeID, destinationNodeID uuid.UUID,
	orderedItemIDs []uuid.UUID,
) (MoveLearningItemsResult, error) {
	result := MoveLearningItemsResult{
		SourceNodeID:      sourceNodeID,
		DestinationNodeID: destinationNodeID,
	}
	if err := validateMoveLearningItemsParams(courseID, sourceNodeID, destinationNodeID, orderedItemIDs); err != nil {
		return result, err
	}

	transaction, err := model.db.Begin()
	if err != nil {
		return result, newLearningItemPersistenceError(err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = transaction.Rollback()
		}
	}()

	if err := lockCourseForLearningItemReorder(transaction, courseID); err != nil {
		return result, err
	}

	firstNodeID, secondNodeID := sourceNodeID, destinationNodeID
	if bytes.Compare(destinationNodeID[:], sourceNodeID[:]) < 0 {
		firstNodeID, secondNodeID = destinationNodeID, sourceNodeID
	}
	if err := ensureLearningItemNodeInCourse(transaction, courseID, firstNodeID, true); err != nil {
		return result, err
	}
	if err := ensureLearningItemNodeInCourse(transaction, courseID, secondNodeID, true); err != nil {
		return result, err
	}

	firstSiblings, err := lockLearningItemSiblings(transaction, courseID, firstNodeID)
	if err != nil {
		return result, err
	}
	secondSiblings, err := lockLearningItemSiblings(transaction, courseID, secondNodeID)
	if err != nil {
		return result, err
	}

	var sourceSiblings, destinationSiblings []LearningItem
	if firstNodeID == sourceNodeID {
		sourceSiblings, destinationSiblings = firstSiblings, secondSiblings
	} else {
		sourceSiblings, destinationSiblings = secondSiblings, firstSiblings
	}
	result.SourceItemCount = len(sourceSiblings)
	result.DestinationItemCount = len(destinationSiblings)

	if err := ensureLearningItemMoveSubsetMatches(sourceSiblings, orderedItemIDs); err != nil {
		return result, err
	}

	if len(orderedItemIDs) == 0 {
		result.Noop = true
		result.ItemsMoved = 0
		if err := transaction.Commit(); err != nil {
			return result, newLearningItemPersistenceError(err)
		}
		committed = true
		return result, nil
	}

	sourceByID := make(map[uuid.UUID]LearningItem, len(sourceSiblings))
	for _, item := range sourceSiblings {
		sourceByID[item.ID] = item
	}
	movedSet := make(map[uuid.UUID]struct{}, len(orderedItemIDs))
	for _, itemID := range orderedItemIDs {
		movedSet[itemID] = struct{}{}
	}

	remainingSource := make([]LearningItem, 0, len(sourceSiblings)-len(orderedItemIDs))
	for _, item := range sourceSiblings {
		if _, moved := movedSet[item.ID]; !moved {
			remainingSource = append(remainingSource, item)
		}
	}
	movedItems := make([]LearningItem, 0, len(orderedItemIDs))
	for _, itemID := range orderedItemIDs {
		movedItems = append(movedItems, sourceByID[itemID])
	}

	maxPosition := 0
	for _, item := range sourceSiblings {
		if item.Position > maxPosition {
			maxPosition = item.Position
		}
	}
	for _, item := range destinationSiblings {
		if item.Position > maxPosition {
			maxPosition = item.Position
		}
	}
	sourceCount64 := int64(len(sourceSiblings))
	destCount64 := int64(len(destinationSiblings))
	maxPosition64 := int64(maxPosition)
	sourceTempBase64 := maxPosition64 + sourceCount64 + destCount64 + 1
	destTempBase64 := sourceTempBase64 + sourceCount64
	highestTemp64 := destTempBase64 + destCount64 - 1
	if destCount64 == 0 {
		highestTemp64 = sourceTempBase64 + sourceCount64 - 1
	}
	if sourceTempBase64 > math.MaxInt32 || highestTemp64 > math.MaxInt32 {
		return result, ErrLearningItemPositionInvalid
	}
	sourceTempBase := int(sourceTempBase64)
	destTempBase := int(destTempBase64)

	// 1) Stage source then destination siblings into disjoint temporary positions.
	for index, item := range sourceSiblings {
		if _, err := updateLearningItemSiblingPosition(
			transaction, courseID, sourceNodeID, item.ID, sourceTempBase+index, false,
		); err != nil {
			return result, mapLearningItemMoveWriteError(err)
		}
	}
	for index, item := range destinationSiblings {
		if _, err := updateLearningItemSiblingPosition(
			transaction, courseID, destinationNodeID, item.ID, destTempBase+index, false,
		); err != nil {
			return result, mapLearningItemMoveWriteError(err)
		}
	}

	// 2) Move selected rows to destination while they remain at source temp positions.
	for _, item := range movedItems {
		if err := updateLearningItemCourseNode(
			transaction, courseID, sourceNodeID, destinationNodeID, item.ID,
		); err != nil {
			return result, err
		}
	}

	// 3) Final source positions 0..n-1 for remaining items.
	for index, item := range remainingSource {
		if _, err := updateLearningItemSiblingPosition(
			transaction, courseID, sourceNodeID, item.ID, index, true,
		); err != nil {
			return result, mapLearningItemMoveWriteError(err)
		}
	}

	// 4) Final destination positions 0..m-1 (existing dest order + moved request order).
	finalDest := make([]LearningItem, 0, len(destinationSiblings)+len(movedItems))
	finalDest = append(finalDest, destinationSiblings...)
	finalDest = append(finalDest, movedItems...)
	for index, item := range finalDest {
		if _, err := updateLearningItemSiblingPosition(
			transaction, courseID, destinationNodeID, item.ID, index, true,
		); err != nil {
			return result, mapLearningItemMoveWriteError(err)
		}
	}

	// 5) Verify ownership and contiguity via locked re-read of both nodes.
	sourceAfter, err := lockLearningItemSiblings(transaction, courseID, sourceNodeID)
	if err != nil {
		return result, err
	}
	destAfter, err := lockLearningItemSiblings(transaction, courseID, destinationNodeID)
	if err != nil {
		return result, err
	}
	if err := verifyLearningItemMoveResult(
		sourceAfter, destAfter, remainingSource, finalDest, sourceNodeID, destinationNodeID,
	); err != nil {
		return result, err
	}

	if err := transaction.Commit(); err != nil {
		return result, newLearningItemPersistenceError(err)
	}
	committed = true
	result.Noop = false
	result.ItemsMoved = len(movedItems)
	result.SourceItemCount = len(sourceAfter)
	result.DestinationItemCount = len(destAfter)
	return result, nil
}

func (model *LearningItemModel) DeleteLearningItem(courseID, nodeID, itemID uuid.UUID) error {
	result, err := model.db.Delete(LearningItemsTable).
		Where(goqu.Ex{
			"id":             itemID,
			"course_id":      courseID,
			"course_node_id": nodeID,
		}).
		Prepared(true).
		Executor().
		Exec()
	if err != nil {
		return newLearningItemPersistenceError(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return newLearningItemPersistenceError(err)
	}
	if affected == 0 {
		return ErrLearningItemNotFound
	}
	return nil
}

func validateCreateLearningItemParams(params CreateLearningItemParams) (string, json.RawMessage, LearningItemPublishState, error) {
	if params.CourseID == uuid.Nil {
		return "", nil, "", ErrLearningItemNodeNotFound
	}
	if params.CourseNodeID == uuid.Nil {
		return "", nil, "", ErrLearningItemNodeNotFound
	}

	title := strings.TrimSpace(params.Title)
	if title == "" {
		return "", nil, "", ErrLearningItemTitleRequired
	}
	if !isValidLearningItemType(params.ItemType) {
		return "", nil, "", ErrLearningItemTypeInvalid
	}
	if params.ItemType == LearningItemTypeQuizRef && params.QuizID == nil {
		return "", nil, "", ErrLearningItemQuizRequired
	}
	if params.ItemType != LearningItemTypeQuizRef && params.QuizID != nil {
		return "", nil, "", ErrLearningItemQuizForbidden
	}

	publishState, err := normalizeCreateLearningItemPublishState(params.PublishState)
	if err != nil {
		return "", nil, "", err
	}

	metadata, err := normalizeLearningItemMetadata(params.Metadata)
	if err != nil {
		return "", nil, "", err
	}
	return title, metadata, publishState, nil
}

func buildLearningItemUpdateRecord(params UpdateLearningItemParams) (goqu.Record, error) {
	record := goqu.Record{}
	if params.Title != nil {
		title := strings.TrimSpace(*params.Title)
		if title == "" {
			return nil, ErrLearningItemTitleRequired
		}
		record["title"] = title
	}
	if params.ItemType != nil {
		if !isValidLearningItemType(*params.ItemType) {
			return nil, ErrLearningItemTypeInvalid
		}
		record["item_type"] = *params.ItemType
	}
	if params.Description.Present {
		record["description"] = optionalNullableToSQL(params.Description)
	}
	if params.Metadata.Present {
		if params.Metadata.Null {
			return nil, ErrLearningItemMetadataInvalid
		}
		metadata, err := normalizeLearningItemMetadata(params.Metadata.Value)
		if err != nil {
			return nil, err
		}
		record["metadata"] = []byte(metadata)
	}
	if params.QuizID.Present {
		if params.QuizID.Null {
			record["quiz_id"] = nil
		} else {
			if params.QuizID.Value == uuid.Nil {
				return nil, ErrLearningItemQuizRequired
			}
			record["quiz_id"] = params.QuizID.Value
		}
	}
	if params.PublishState != nil {
		if !isValidLearningItemPublishState(*params.PublishState) {
			return nil, ErrLearningItemPublishStateInvalid
		}
		record["publish_state"] = *params.PublishState
	}
	if len(record) == 0 {
		return nil, ErrLearningItemUpdateRequired
	}
	record["updated_at"] = goqu.L("now()")
	return record, nil
}

func normalizeCreateLearningItemPublishState(state LearningItemPublishState) (LearningItemPublishState, error) {
	if state == "" {
		return LearningItemPublishStateDraft, nil
	}
	if !isValidLearningItemPublishState(state) {
		return "", ErrLearningItemPublishStateInvalid
	}
	return state, nil
}

func isValidLearningItemType(itemType LearningItemType) bool {
	switch itemType {
	case LearningItemTypeArticle, LearningItemTypeVideo, LearningItemTypePDF, LearningItemTypeLink, LearningItemTypeQuizRef:
		return true
	default:
		return false
	}
}

func isValidLearningItemPublishState(state LearningItemPublishState) bool {
	switch state {
	case LearningItemPublishStateDraft, LearningItemPublishStatePublished:
		return true
	default:
		return false
	}
}

func ensureManageableLearningItemQuiz(
	querier learningItemQuerier,
	quizID uuid.UUID,
	actorID string,
) error {
	if quizID == uuid.Nil {
		return ErrLearningItemQuizRequired
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return ErrLearningItemQuizForbidden
	}
	var foundID uuid.UUID
	found, err := querier.From(QuizzesTable).
		Select("id").
		Where(
			goqu.Ex{"id": quizID},
			goqu.Or(
				goqu.Ex{"creator_id": actorID},
				goqu.Ex{"is_public": true},
			),
		).
		Limit(1).
		Prepared(true).
		ScanVal(&foundID)
	if err != nil {
		return newLearningItemPersistenceError(err)
	}
	if !found {
		return ErrLearningItemQuizNotFound
	}
	return nil
}

type learningItemQuerier interface {
	From(cols ...interface{}) *goqu.SelectDataset
}

func validateReorderLearningItemsParams(courseID, nodeID uuid.UUID, orderedItemIDs []uuid.UUID) error {
	if courseID == uuid.Nil {
		return ErrCourseNotFound
	}
	if nodeID == uuid.Nil {
		return ErrLearningItemNodeNotFound
	}
	seen := make(map[uuid.UUID]struct{}, len(orderedItemIDs))
	for _, itemID := range orderedItemIDs {
		if itemID == uuid.Nil {
			return ErrLearningItemNotFound
		}
		if _, exists := seen[itemID]; exists {
			return ErrLearningItemReorderDuplicate
		}
		seen[itemID] = struct{}{}
	}
	return nil
}

func lockCourseForLearningItemReorder(transaction *goqu.TxDatabase, courseID uuid.UUID) error {
	found, err := transaction.From(CoursesTable).
		Select("id").
		Where(goqu.Ex{"id": courseID}).
		ForUpdate(goqu.Wait).
		Prepared(true).
		ScanVal(new(uuid.UUID))
	if err != nil {
		return newLearningItemPersistenceError(err)
	}
	if !found {
		return ErrCourseNotFound
	}
	return nil
}

func lockLearningItemSiblings(transaction *goqu.TxDatabase, courseID, nodeID uuid.UUID) ([]LearningItem, error) {
	items := make([]LearningItem, 0)
	err := transaction.From(LearningItemsTable).
		Select(learningItemColumns...).
		Where(goqu.Ex{"course_id": courseID, "course_node_id": nodeID}).
		Order(goqu.I("position").Asc(), goqu.I("id").Asc()).
		ForUpdate(goqu.Wait).
		Prepared(true).
		ScanStructs(&items)
	if err != nil {
		return nil, newLearningItemPersistenceError(err)
	}
	return items, nil
}

// ensureLearningItemReorderSetMatches requires a bidirectional exact set match:
// every locked sibling exactly once, no missing IDs, no extra IDs, no duplicates
// (duplicates are rejected earlier by validateReorderLearningItemsParams).
func ensureLearningItemReorderSetMatches(lockedSiblings []LearningItem, orderedItemIDs []uuid.UUID) error {
	if len(lockedSiblings) != len(orderedItemIDs) {
		return ErrLearningItemReorderMismatch
	}
	lockedIDs := make(map[uuid.UUID]struct{}, len(lockedSiblings))
	for _, sibling := range lockedSiblings {
		lockedIDs[sibling.ID] = struct{}{}
	}
	for _, itemID := range orderedItemIDs {
		if _, ok := lockedIDs[itemID]; !ok {
			return ErrLearningItemReorderMismatch
		}
	}
	return nil
}

func updateLearningItemSiblingPosition(
	transaction *goqu.TxDatabase,
	courseID, nodeID, itemID uuid.UUID,
	position int,
	touchUpdatedAt bool,
) (uuid.UUID, error) {
	record := goqu.Record{"position": position}
	if touchUpdatedAt {
		record["updated_at"] = goqu.L("now()")
	}
	var updatedID uuid.UUID
	found, err := transaction.Update(LearningItemsTable).
		Set(record).
		Where(goqu.Ex{
			"course_id":      courseID,
			"course_node_id": nodeID,
			"id":             itemID,
		}).
		Returning(goqu.C("id")).
		Prepared(true).
		Executor().
		ScanVal(&updatedID)
	if err != nil {
		return uuid.Nil, mapLearningItemReorderWriteError(err)
	}
	if !found {
		return uuid.Nil, ErrLearningItemReorderConflict
	}
	return updatedID, nil
}

func verifyLearningItemReorderUpdatedIDs(expected map[uuid.UUID]struct{}, returned []uuid.UUID) error {
	if len(returned) != len(expected) {
		return ErrLearningItemReorderConflict
	}
	seen := make(map[uuid.UUID]struct{}, len(returned))
	for _, id := range returned {
		if _, ok := expected[id]; !ok {
			return ErrLearningItemReorderConflict
		}
		if _, duplicate := seen[id]; duplicate {
			return ErrLearningItemReorderConflict
		}
		seen[id] = struct{}{}
	}
	for id := range expected {
		if _, ok := seen[id]; !ok {
			return ErrLearningItemReorderConflict
		}
	}
	return nil
}

func mapLearningItemReorderWriteError(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		if pqErr.Constraint == learningItemsNodePositionConstraint {
			return newLearningItemSemanticError(ErrLearningItemReorderConflict, err)
		}
	}
	return newLearningItemPersistenceError(err)
}

func validateGetAdjacentLearningItemsParams(courseID, nodeID, currentItemID uuid.UUID) error {
	if courseID == uuid.Nil {
		return ErrCourseNotFound
	}
	if nodeID == uuid.Nil {
		return ErrLearningItemNodeNotFound
	}
	if currentItemID == uuid.Nil {
		return ErrLearningItemNotFound
	}
	return nil
}

// findAdjacentLearningItem returns the immediate previous (wantNext=false) or next
// (wantNext=true) sibling under the same CourseNode using (position, id) tuple order.
func (model *LearningItemModel) findAdjacentLearningItem(
	courseID, nodeID uuid.UUID,
	current LearningItem,
	wantNext bool,
	publishedOnly bool,
) (*LearningItem, error) {
	var item LearningItem
	where := goqu.Ex{
		"course_id":      courseID,
		"course_node_id": nodeID,
	}
	if publishedOnly {
		where["publish_state"] = LearningItemPublishStatePublished
	}

	scoped := model.db.From(LearningItemsTable).
		Select(learningItemColumns...).
		Where(where).
		Limit(1).
		Prepared(true)

	if wantNext {
		scoped = scoped.Where(goqu.Or(
			goqu.C("position").Gt(current.Position),
			goqu.And(
				goqu.C("position").Eq(current.Position),
				goqu.C("id").Gt(current.ID),
			),
		)).Order(goqu.I("position").Asc(), goqu.I("id").Asc())
	} else {
		scoped = scoped.Where(goqu.Or(
			goqu.C("position").Lt(current.Position),
			goqu.And(
				goqu.C("position").Eq(current.Position),
				goqu.C("id").Lt(current.ID),
			),
		)).Order(goqu.I("position").Desc(), goqu.I("id").Desc())
	}

	found, err := scoped.ScanStruct(&item)
	if err != nil {
		return nil, newLearningItemPersistenceError(err)
	}
	if !found {
		return nil, nil
	}
	return &item, nil
}

func validateMoveLearningItemsParams(
	courseID, sourceNodeID, destinationNodeID uuid.UUID,
	orderedItemIDs []uuid.UUID,
) error {
	if courseID == uuid.Nil {
		return ErrCourseNotFound
	}
	if sourceNodeID == uuid.Nil || destinationNodeID == uuid.Nil {
		return ErrLearningItemNodeNotFound
	}
	if sourceNodeID == destinationNodeID {
		return ErrLearningItemMoveSameNode
	}
	seen := make(map[uuid.UUID]struct{}, len(orderedItemIDs))
	for _, itemID := range orderedItemIDs {
		if itemID == uuid.Nil {
			return ErrLearningItemMoveMismatch
		}
		if _, exists := seen[itemID]; exists {
			return ErrLearningItemMoveDuplicate
		}
		seen[itemID] = struct{}{}
	}
	return nil
}

// ensureLearningItemMoveSubsetMatches requires every requested ID to belong to the
// source sibling set. Missing or foreign IDs map to ErrLearningItemMoveMismatch
// without revealing whether the ID exists on another node or course.
func ensureLearningItemMoveSubsetMatches(sourceSiblings []LearningItem, orderedItemIDs []uuid.UUID) error {
	sourceIDs := make(map[uuid.UUID]struct{}, len(sourceSiblings))
	for _, item := range sourceSiblings {
		sourceIDs[item.ID] = struct{}{}
	}
	for _, itemID := range orderedItemIDs {
		if _, ok := sourceIDs[itemID]; !ok {
			return ErrLearningItemMoveMismatch
		}
	}
	return nil
}

func updateLearningItemCourseNode(
	transaction *goqu.TxDatabase,
	courseID, currentNodeID, newNodeID, itemID uuid.UUID,
) error {
	var updatedID uuid.UUID
	found, err := transaction.Update(LearningItemsTable).
		Set(goqu.Record{
			"course_node_id": newNodeID,
			"updated_at":     goqu.L("now()"),
		}).
		Where(goqu.Ex{
			"course_id":      courseID,
			"course_node_id": currentNodeID,
			"id":             itemID,
		}).
		Returning(goqu.C("id")).
		Prepared(true).
		Executor().
		ScanVal(&updatedID)
	if err != nil {
		return mapLearningItemMoveWriteError(err)
	}
	if !found {
		return ErrLearningItemMoveConflict
	}
	return nil
}

func mapLearningItemMoveWriteError(err error) error {
	if errors.Is(err, ErrLearningItemReorderConflict) {
		return ErrLearningItemMoveConflict
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		if pqErr.Constraint == learningItemsNodePositionConstraint {
			return newLearningItemSemanticError(ErrLearningItemMoveConflict, err)
		}
	}
	if errors.Is(err, ErrLearningItemPersistence) {
		return err
	}
	var liErr *learningItemError
	if errors.As(err, &liErr) {
		return err
	}
	return newLearningItemPersistenceError(err)
}

func verifyLearningItemMoveResult(
	sourceAfter, destAfter []LearningItem,
	expectedRemaining, expectedDest []LearningItem,
	sourceNodeID, destinationNodeID uuid.UUID,
) error {
	if len(sourceAfter) != len(expectedRemaining) || len(destAfter) != len(expectedDest) {
		return ErrLearningItemMoveConflict
	}
	for index, item := range sourceAfter {
		if item.Position != index || item.ID != expectedRemaining[index].ID {
			return ErrLearningItemMoveConflict
		}
		if item.CourseNodeID != sourceNodeID {
			return ErrLearningItemMoveConflict
		}
	}
	for index, item := range destAfter {
		if item.Position != index || item.ID != expectedDest[index].ID {
			return ErrLearningItemMoveConflict
		}
		if item.CourseNodeID != destinationNodeID {
			return ErrLearningItemMoveConflict
		}
	}
	return nil
}

func ensureLearningItemNodeInCourse(db learningItemQuerier, courseID, nodeID uuid.UUID, forUpdate bool) error {
	var node struct {
		ID       uuid.UUID `db:"id"`
		CourseID uuid.UUID `db:"course_id"`
	}
	scoped := db.From(CourseNodesTable).
		Select("id", "course_id").
		Where(goqu.Ex{"id": nodeID, "course_id": courseID}).
		Prepared(true)
	if forUpdate {
		scoped = scoped.ForUpdate(goqu.Wait)
	}
	found, err := scoped.ScanStruct(&node)
	if err != nil {
		return newLearningItemPersistenceError(err)
	}
	if found {
		return nil
	}

	found, err = db.From(CourseNodesTable).
		Select("id", "course_id").
		Where(goqu.Ex{"id": nodeID}).
		Limit(1).
		Prepared(true).
		ScanStruct(&node)
	if err != nil {
		return newLearningItemPersistenceError(err)
	}
	if found {
		return ErrLearningItemCrossCourse
	}
	return ErrLearningItemNodeNotFound
}

func mapLearningItemWriteError(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		if pqErr.Constraint == learningItemsNodePositionConstraint {
			return newLearningItemSemanticError(ErrLearningItemConflict, err)
		}
	}
	return newLearningItemPersistenceError(err)
}

type learningItemError struct {
	kind  error
	cause error
}

func (err *learningItemError) Error() string {
	return err.kind.Error()
}

func (err *learningItemError) Unwrap() []error {
	return []error{err.kind, err.cause}
}

func newLearningItemPersistenceError(cause error) error {
	return &learningItemError{kind: ErrLearningItemPersistence, cause: cause}
}

func newLearningItemSemanticError(kind, cause error) error {
	return &learningItemError{kind: kind, cause: cause}
}

type LearningChainItem struct {
	ID    uuid.UUID `db:"id"`
	Title string    `db:"title"`
}

// ProjectPublishedLearningChain returns only the published ID and Title values for a node,
// ordered by position ASC, id ASC. Successful empty nodes return an empty non-nil slice.
func (model *LearningItemModel) ProjectPublishedLearningChain(
	courseID uuid.UUID,
	nodeID uuid.UUID,
) ([]LearningChainItem, error) {
	if courseID == uuid.Nil {
		return nil, ErrCourseNotFound
	}
	if nodeID == uuid.Nil {
		return nil, ErrLearningItemNodeNotFound
	}

	if err := ensureLearningItemNodeInCourse(model.db, courseID, nodeID, false); err != nil {
		return nil, err
	}

	chain := make([]LearningChainItem, 0)
	err := model.db.From(LearningItemsTable).
		Select("id", "title").
		Where(goqu.Ex{
			"course_id":      courseID,
			"course_node_id": nodeID,
			"publish_state":  LearningItemPublishStatePublished,
		}).
		Order(goqu.I("position").Asc(), goqu.I("id").Asc()).
		Prepared(true).
		ScanStructs(&chain)
	if err != nil {
		return nil, newLearningItemPersistenceError(err)
	}

	return chain, nil
}
