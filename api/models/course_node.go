package models

import (
	"database/sql"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

const CourseNodesTable = "course_nodes"

const (
	courseNodesTopLevelPositionConstraint = "course_nodes_top_level_position_unique"
	courseNodesChildPositionConstraint    = "course_nodes_child_position_unique"
)

type CourseNodeType string

const (
	SECTION CourseNodeType = "SECTION"
	SUBJECT CourseNodeType = "SUBJECT"
	TOPIC   CourseNodeType = "TOPIC"

	CourseNodeTypeSection = SECTION
	CourseNodeTypeSubject = SUBJECT
	CourseNodeTypeTopic   = TOPIC
)

var (
	ErrCourseNodeCourseRequired      = errors.New("course node course is required")
	ErrCourseNodeCourseNotFound      = errors.New("course node course not found")
	ErrCourseNodeTitleRequired       = errors.New("course node title is required")
	ErrCourseNodeTypeInvalid         = errors.New("course node type is invalid")
	ErrCourseNodePositionInvalid     = errors.New("course node position is invalid")
	ErrCourseNodeParentNotFound      = errors.New("course node parent not found")
	ErrCourseNodeCrossCourseParent   = errors.New("course node parent belongs to another course")
	ErrCourseNodeNotFound            = errors.New("course node not found")
	ErrCourseNodePositionConflict    = errors.New("course node sibling position conflicts")
	ErrCourseNodeHierarchyIntegrity  = errors.New("course node hierarchy integrity failure")
	ErrCourseNodeCycle               = errors.New("course node move would create a cycle")
	ErrCourseNodeInvalidMove         = errors.New("course node move is invalid")
	ErrCourseNodeSubtreeConflict     = errors.New("course node subtree changed during move")
	ErrCourseNodeReorderDuplicate    = errors.New("course node reorder duplicate ID")
	ErrCourseNodeReorderMismatch     = errors.New("course node reorder mismatched ID set")
	ErrCourseNodeReorderConflict     = errors.New("course node reorder concurrency conflict")
	ErrCourseNodeDeleteConflict      = errors.New("course node delete concurrency conflict")
	ErrCourseNodeDeleteRootProtected = errors.New("course node delete root protected")
	ErrCourseNodePersistence         = errors.New("course node persistence failure")
)

var courseNodeColumns = []interface{}{
	"id",
	"course_id",
	"parent_id",
	"node_type",
	"title",
	"position",
	"path",
	"status",
	"created_at",
	"updated_at",
}

type CourseNode struct {
	ID        uuid.UUID      `json:"id" db:"id"`
	CourseID  uuid.UUID      `json:"course_id" db:"course_id"`
	ParentID  uuid.NullUUID  `json:"parent_id" db:"parent_id"`
	NodeType  CourseNodeType `json:"node_type" db:"node_type"`
	Title     string         `json:"title" db:"title"`
	Position  int            `json:"position" db:"position"`
	Path      string         `json:"-" db:"path"`
	Status    CourseStatus   `json:"status" db:"status"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}

// CourseNodeHierarchy is a repository read model. It deliberately has no
// parent pointer so callers receive a serializable, acyclic tree.
type CourseNodeHierarchy struct {
	Node     CourseNode            `json:"node"`
	Children []CourseNodeHierarchy `json:"children"`
}

// CourseHierarchy identifies the Course represented by a nested hierarchy.
// Roots is non-nil, including for an existing Course with no nodes.
type CourseHierarchy struct {
	CourseID uuid.UUID             `json:"courseId"`
	Roots    []CourseNodeHierarchy `json:"roots"`
}

type CreateCourseNodeParams struct {
	CourseID uuid.UUID
	ParentID uuid.NullUUID
	NodeType CourseNodeType
	Title    string
	Position int
}

type CourseNodeModel struct {
	db      *goqu.Database
	newUUID func() (uuid.UUID, error)
}

func InitCourseNodeModel(goquDB *goqu.Database) *CourseNodeModel {
	return &CourseNodeModel{
		db:      goquDB,
		newUUID: uuid.NewUUID,
	}
}

func (model *CourseNodeModel) CreateCourseNode(params CreateCourseNodeParams) (CourseNode, error) {
	var node CourseNode

	title, err := validateCreateCourseNodeParams(params)
	if err != nil {
		return node, err
	}

	transaction, err := model.db.Begin()
	if err != nil {
		return node, newCourseNodePersistenceError(err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = transaction.Rollback()
		}
	}()

	found, err := transaction.From(CoursesTable).
		Select("id").
		Where(goqu.Ex{"id": params.CourseID}).
		ForUpdate(goqu.Wait).
		Prepared(true).
		ScanVal(new(uuid.UUID))
	if err != nil {
		return node, newCourseNodePersistenceError(err)
	}
	if !found {
		return node, ErrCourseNodeCourseNotFound
	}

	parentPath := ""
	if params.ParentID.Valid {
		var parent struct {
			ID       uuid.UUID `db:"id"`
			CourseID uuid.UUID `db:"course_id"`
			Path     string    `db:"path"`
		}
		found, err = transaction.From(CourseNodesTable).
			Select("id", "course_id", "path").
			Where(goqu.Ex{"id": params.ParentID.UUID}).
			ForUpdate(goqu.Wait).
			Prepared(true).
			ScanStruct(&parent)
		if err != nil {
			return node, newCourseNodePersistenceError(err)
		}
		if !found {
			return node, ErrCourseNodeParentNotFound
		}
		if parent.CourseID != params.CourseID {
			return node, ErrCourseNodeCrossCourseParent
		}
		parentPath = parent.Path
	}

	nodeID, err := model.newUUID()
	if err != nil {
		return node, newCourseNodePersistenceError(err)
	}
	path := encodeCourseNodePath(parentPath, nodeID)

	found, err = transaction.Insert(CourseNodesTable).
		Rows(goqu.Record{
			"id":        nodeID,
			"course_id": params.CourseID,
			"parent_id": params.ParentID,
			"node_type": params.NodeType,
			"title":     title,
			"position":  params.Position,
			"path":      path,
			"status":    CourseStatusDraft,
		}).
		Returning(courseNodeColumns...).
		Prepared(true).
		Executor().
		ScanStruct(&node)
	if err != nil {
		return node, mapCourseNodeWriteError(err)
	}
	if !found {
		return node, newCourseNodePersistenceError(sql.ErrNoRows)
	}
	if err := transaction.Commit(); err != nil {
		return CourseNode{}, newCourseNodePersistenceError(err)
	}
	committed = true

	return node, nil
}

func (model *CourseNodeModel) GetCourseNodeByID(courseID, nodeID uuid.UUID) (CourseNode, error) {
	var node CourseNode

	found, err := model.db.From(CourseNodesTable).
		Select(courseNodeColumns...).
		Where(goqu.Ex{
			"course_id": courseID,
			"id":        nodeID,
		}).
		Limit(1).
		Prepared(true).
		ScanStruct(&node)
	if err != nil {
		return node, newCourseNodePersistenceError(err)
	}
	if !found {
		return node, ErrCourseNodeNotFound
	}
	return node, nil
}

// DeleteSubtree deletes a CourseNode and every boundary-safe descendant in one
// transaction. Unrelated branches, sibling positions, and paths are untouched.
func (model *CourseNodeModel) DeleteSubtree(courseID, nodeID uuid.UUID) error {
	if courseID == uuid.Nil {
		return ErrCourseNodeCourseRequired
	}
	if nodeID == uuid.Nil {
		return ErrCourseNodeNotFound
	}

	transaction, err := model.db.Begin()
	if err != nil {
		return newCourseNodePersistenceError(err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = transaction.Rollback()
		}
	}()

	if err := lockCourseForCourseNodeMove(transaction, courseID); err != nil {
		return err
	}
	target, err := lockCourseNodeForMove(transaction, courseID, nodeID)
	if err != nil {
		return err
	}

	subtree, err := lockCourseNodeSubtreeForDelete(transaction, courseID, target.Path)
	if err != nil {
		return err
	}
	if len(subtree) == 0 {
		return ErrCourseNodeDeleteConflict
	}

	expectedIDs := make(map[uuid.UUID]struct{}, len(subtree))
	rootPresent := false
	for _, subtreeNode := range subtree {
		if !isCourseNodePathInSubtree(subtreeNode.Path, target.Path) {
			return ErrCourseNodeDeleteConflict
		}
		if subtreeNode.ID == target.ID {
			rootPresent = true
		}
		expectedIDs[subtreeNode.ID] = struct{}{}
	}
	if !rootPresent {
		return ErrCourseNodeDeleteConflict
	}
	if len(expectedIDs) != len(subtree) {
		return ErrCourseNodeDeleteConflict
	}

	deletedIDs, err := deleteCourseNodeSubtree(transaction, courseID, target.Path)
	if err != nil {
		return err
	}
	if err := verifyDeleteIDs(expectedIDs, deletedIDs); err != nil {
		return err
	}

	if err := transaction.Commit(); err != nil {
		return newCourseNodePersistenceError(err)
	}
	committed = true
	return nil
}

// ReorderChildren replaces the sibling order under a parent (or course roots)
// with the exact ordered ID list. Positions are normalized to 0..n-1 using a
// two-phase temporary staging update to avoid unique-index collisions.
func (model *CourseNodeModel) ReorderChildren(
	courseID uuid.UUID,
	parentID uuid.NullUUID,
	orderedNodeIDs []uuid.UUID,
) error {
	if err := validateReorderChildrenParams(courseID, parentID, orderedNodeIDs); err != nil {
		return err
	}

	transaction, err := model.db.Begin()
	if err != nil {
		return newCourseNodePersistenceError(err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = transaction.Rollback()
		}
	}()

	if err := lockCourseForCourseNodeMove(transaction, courseID); err != nil {
		return err
	}
	if parentID.Valid {
		if _, err := lockMoveDestinationParent(transaction, courseID, parentID.UUID); err != nil {
			return err
		}
	}

	lockedSiblings, err := lockCourseNodeSiblings(transaction, courseID, parentID)
	if err != nil {
		return err
	}
	if len(lockedSiblings) == 0 {
		if len(orderedNodeIDs) != 0 {
			return ErrCourseNodeReorderMismatch
		}
		if err := transaction.Commit(); err != nil {
			return newCourseNodePersistenceError(err)
		}
		committed = true
		return nil
	}
	if err := ensureReorderSiblingSetMatches(lockedSiblings, orderedNodeIDs); err != nil {
		return err
	}

	finalPositionByID := make(map[uuid.UUID]int, len(orderedNodeIDs))
	for index, nodeID := range orderedNodeIDs {
		finalPositionByID[nodeID] = index
	}
	alreadyCanonical := true
	for _, sibling := range lockedSiblings {
		if sibling.Position != finalPositionByID[sibling.ID] {
			alreadyCanonical = false
			break
		}
	}
	if alreadyCanonical {
		if err := transaction.Commit(); err != nil {
			return newCourseNodePersistenceError(err)
		}
		committed = true
		return nil
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
		return ErrCourseNodePositionInvalid
	}
	temporaryBase := int(temporaryBase64)

	expectedIDs := make(map[uuid.UUID]struct{}, len(lockedSiblings))
	for _, sibling := range lockedSiblings {
		expectedIDs[sibling.ID] = struct{}{}
	}

	phase1IDs := make([]uuid.UUID, 0, len(lockedSiblings))
	for index, sibling := range lockedSiblings {
		updatedID, err := updateCourseNodeSiblingPosition(
			transaction,
			courseID,
			sibling.ID,
			temporaryBase+index,
			false,
		)
		if err != nil {
			return err
		}
		phase1IDs = append(phase1IDs, updatedID)
	}
	if err := verifyReorderUpdatedIDs(expectedIDs, phase1IDs); err != nil {
		return err
	}

	phase2IDs := make([]uuid.UUID, 0, len(lockedSiblings))
	for _, sibling := range lockedSiblings {
		updatedID, err := updateCourseNodeSiblingPosition(
			transaction,
			courseID,
			sibling.ID,
			finalPositionByID[sibling.ID],
			true,
		)
		if err != nil {
			return err
		}
		phase2IDs = append(phase2IDs, updatedID)
	}
	if err := verifyReorderUpdatedIDs(expectedIDs, phase2IDs); err != nil {
		return err
	}

	if err := transaction.Commit(); err != nil {
		return newCourseNodePersistenceError(err)
	}
	committed = true
	return nil
}

// MoveNode reparents a CourseNode and rewrites its complete subtree path in
// one transaction. Same-parent position changes must use ReorderChildren.
func (model *CourseNodeModel) MoveNode(courseID, nodeID uuid.UUID, newParentID uuid.NullUUID, newPosition int) error {
	if err := validateMoveCourseNodeParams(courseID, nodeID, newParentID, newPosition); err != nil {
		return err
	}

	transaction, err := model.db.Begin()
	if err != nil {
		return newCourseNodePersistenceError(err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = transaction.Rollback()
		}
	}()

	if err := lockCourseForCourseNodeMove(transaction, courseID); err != nil {
		return err
	}
	movedNode, err := lockCourseNodeForMove(transaction, courseID, nodeID)
	if err != nil {
		return err
	}
	if sameCourseNodeParent(movedNode.ParentID, newParentID) {
		if movedNode.Position == newPosition {
			if err := transaction.Commit(); err != nil {
				return newCourseNodePersistenceError(err)
			}
			committed = true
			return nil
		}
		return ErrCourseNodeInvalidMove
	}

	newParentPath := ""
	if newParentID.Valid {
		parent, err := lockMoveDestinationParent(transaction, courseID, newParentID.UUID)
		if err != nil {
			return err
		}
		newParentPath = parent.Path
	}

	subtree, err := lockCourseNodeSubtree(transaction, courseID, movedNode.Path)
	if err != nil {
		return err
	}
	if len(subtree) == 0 {
		return ErrCourseNodeSubtreeConflict
	}
	subtreeIDs := make(map[uuid.UUID]struct{}, len(subtree))
	rootPresent := false
	for _, subtreeNode := range subtree {
		if !isCourseNodePathInSubtree(subtreeNode.Path, movedNode.Path) {
			return ErrCourseNodeSubtreeConflict
		}
		if subtreeNode.ID == movedNode.ID {
			rootPresent = true
		}
		subtreeIDs[subtreeNode.ID] = struct{}{}
	}
	if !rootPresent {
		return ErrCourseNodeSubtreeConflict
	}
	if newParentID.Valid {
		if _, isDescendant := subtreeIDs[newParentID.UUID]; isDescendant {
			return ErrCourseNodeCycle
		}
	}
	if err := ensureMovePositionAvailable(transaction, courseID, newParentID, newPosition, nodeID); err != nil {
		return err
	}

	newRootPath := encodeCourseNodePath(newParentPath, movedNode.ID)
	updatedPathIDs, err := rewriteCourseNodeSubtreePaths(transaction, courseID, movedNode.Path, newRootPath)
	if err != nil {
		return err
	}
	if len(updatedPathIDs) != len(subtree) {
		return ErrCourseNodeSubtreeConflict
	}
	for _, subtreeNode := range subtree {
		if _, updated := updatedPathIDs[subtreeNode.ID]; !updated {
			return ErrCourseNodeSubtreeConflict
		}
	}

	result, err := transaction.Update(CourseNodesTable).
		Set(goqu.Record{
			"parent_id":  nullableCourseNodeParent(newParentID),
			"position":   newPosition,
			"updated_at": goqu.L("now()"),
		}).
		Where(goqu.Ex{"course_id": courseID, "id": nodeID}).
		Executor().
		Exec()
	if err != nil {
		return mapCourseNodeWriteError(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return newCourseNodePersistenceError(err)
	}
	if affected != 1 {
		return ErrCourseNodeSubtreeConflict
	}
	if err := transaction.Commit(); err != nil {
		return newCourseNodePersistenceError(err)
	}
	committed = true
	return nil
}

func validateMoveCourseNodeParams(courseID, nodeID uuid.UUID, newParentID uuid.NullUUID, newPosition int) error {
	if courseID == uuid.Nil {
		return ErrCourseNodeCourseRequired
	}
	if nodeID == uuid.Nil {
		return ErrCourseNodeNotFound
	}
	if newParentID.Valid && newParentID.UUID == uuid.Nil {
		return ErrCourseNodeParentNotFound
	}
	if newPosition < 0 {
		return ErrCourseNodePositionInvalid
	}
	return nil
}

func validateReorderChildrenParams(courseID uuid.UUID, parentID uuid.NullUUID, orderedNodeIDs []uuid.UUID) error {
	if courseID == uuid.Nil {
		return ErrCourseNodeCourseRequired
	}
	if parentID.Valid && parentID.UUID == uuid.Nil {
		return ErrCourseNodeParentNotFound
	}
	seen := make(map[uuid.UUID]struct{}, len(orderedNodeIDs))
	for _, nodeID := range orderedNodeIDs {
		if nodeID == uuid.Nil {
			return ErrCourseNodeNotFound
		}
		if _, exists := seen[nodeID]; exists {
			return ErrCourseNodeReorderDuplicate
		}
		seen[nodeID] = struct{}{}
	}
	return nil
}

func lockCourseNodeSiblings(transaction *goqu.TxDatabase, courseID uuid.UUID, parentID uuid.NullUUID) ([]CourseNode, error) {
	query := transaction.From(CourseNodesTable).
		Select(courseNodeColumns...).
		Where(goqu.Ex{"course_id": courseID}).
		Order(goqu.I("id").Asc()).
		ForUpdate(goqu.Wait).
		Prepared(true)
	if parentID.Valid {
		query = query.Where(goqu.Ex{"parent_id": parentID.UUID})
	} else {
		query = query.Where(goqu.Ex{"parent_id": nil})
	}
	nodes := make([]CourseNode, 0)
	if err := query.ScanStructs(&nodes); err != nil {
		return nil, newCourseNodePersistenceError(err)
	}
	return nodes, nil
}

func ensureReorderSiblingSetMatches(lockedSiblings []CourseNode, orderedNodeIDs []uuid.UUID) error {
	if len(lockedSiblings) != len(orderedNodeIDs) {
		return ErrCourseNodeReorderMismatch
	}
	lockedIDs := make(map[uuid.UUID]struct{}, len(lockedSiblings))
	for _, sibling := range lockedSiblings {
		lockedIDs[sibling.ID] = struct{}{}
	}
	for _, nodeID := range orderedNodeIDs {
		if _, ok := lockedIDs[nodeID]; !ok {
			return ErrCourseNodeReorderMismatch
		}
	}
	return nil
}

func updateCourseNodeSiblingPosition(
	transaction *goqu.TxDatabase,
	courseID, nodeID uuid.UUID,
	position int,
	touchUpdatedAt bool,
) (uuid.UUID, error) {
	record := goqu.Record{"position": position}
	if touchUpdatedAt {
		record["updated_at"] = goqu.L("now()")
	}
	var updatedID uuid.UUID
	found, err := transaction.Update(CourseNodesTable).
		Set(record).
		Where(goqu.Ex{"course_id": courseID, "id": nodeID}).
		Returning(goqu.C("id")).
		Prepared(true).
		Executor().
		ScanVal(&updatedID)
	if err != nil {
		return uuid.Nil, mapCourseNodeWriteError(err)
	}
	if !found {
		return uuid.Nil, ErrCourseNodeReorderConflict
	}
	return updatedID, nil
}

func verifyReorderUpdatedIDs(expected map[uuid.UUID]struct{}, returned []uuid.UUID) error {
	if len(returned) != len(expected) {
		return ErrCourseNodeReorderConflict
	}
	seen := make(map[uuid.UUID]struct{}, len(returned))
	for _, id := range returned {
		if _, ok := expected[id]; !ok {
			return ErrCourseNodeReorderConflict
		}
		if _, duplicate := seen[id]; duplicate {
			return ErrCourseNodeReorderConflict
		}
		seen[id] = struct{}{}
	}
	for id := range expected {
		if _, ok := seen[id]; !ok {
			return ErrCourseNodeReorderConflict
		}
	}
	return nil
}

func lockCourseForCourseNodeMove(transaction *goqu.TxDatabase, courseID uuid.UUID) error {
	found, err := transaction.From(CoursesTable).
		Select("id").
		Where(goqu.Ex{"id": courseID}).
		ForUpdate(goqu.Wait).
		Prepared(true).
		ScanVal(new(uuid.UUID))
	if err != nil {
		return newCourseNodePersistenceError(err)
	}
	if !found {
		return ErrCourseNodeCourseNotFound
	}
	return nil
}

func lockCourseNodeForMove(transaction *goqu.TxDatabase, courseID, nodeID uuid.UUID) (CourseNode, error) {
	var node CourseNode
	found, err := transaction.From(CourseNodesTable).
		Select(courseNodeColumns...).
		Where(goqu.Ex{"course_id": courseID, "id": nodeID}).
		ForUpdate(goqu.Wait).
		Prepared(true).
		ScanStruct(&node)
	if err != nil {
		return node, newCourseNodePersistenceError(err)
	}
	if !found {
		return node, ErrCourseNodeNotFound
	}
	return node, nil
}

func lockMoveDestinationParent(transaction *goqu.TxDatabase, courseID, parentID uuid.UUID) (CourseNode, error) {
	var parent CourseNode
	found, err := transaction.From(CourseNodesTable).
		Select(courseNodeColumns...).
		Where(goqu.Ex{"id": parentID}).
		ForUpdate(goqu.Wait).
		Prepared(true).
		ScanStruct(&parent)
	if err != nil {
		return parent, newCourseNodePersistenceError(err)
	}
	if !found {
		return parent, ErrCourseNodeParentNotFound
	}
	if parent.CourseID != courseID {
		return parent, ErrCourseNodeCrossCourseParent
	}
	return parent, nil
}

func lockCourseNodeSubtree(transaction *goqu.TxDatabase, courseID uuid.UUID, rootPath string) ([]CourseNode, error) {
	nodes := make([]CourseNode, 0)
	err := transaction.ScanStructs(&nodes, courseNodeSubtreeLockQuery, courseID, rootPath, courseNodeDescendantPathPattern(rootPath))
	if err != nil {
		return nil, newCourseNodePersistenceError(err)
	}
	return nodes, nil
}

func lockCourseNodeSubtreeForDelete(transaction *goqu.TxDatabase, courseID uuid.UUID, rootPath string) ([]CourseNode, error) {
	nodes := make([]CourseNode, 0)
	err := transaction.ScanStructs(&nodes, courseNodeSubtreeDeleteLockQuery, courseID, rootPath, courseNodeDescendantPathPattern(rootPath))
	if err != nil {
		return nil, newCourseNodePersistenceError(err)
	}
	return nodes, nil
}

func deleteCourseNodeSubtree(transaction *goqu.TxDatabase, courseID uuid.UUID, rootPath string) ([]uuid.UUID, error) {
	deletedIDs := make([]uuid.UUID, 0)
	err := transaction.ScanVals(
		&deletedIDs,
		courseNodeSubtreeDeleteQuery,
		courseID,
		rootPath,
		courseNodeDescendantPathPattern(rootPath),
	)
	if err != nil {
		return nil, mapCourseNodeWriteError(err)
	}
	return deletedIDs, nil
}

func verifyDeleteIDs(expected map[uuid.UUID]struct{}, returned []uuid.UUID) error {
	if len(returned) != len(expected) {
		return ErrCourseNodeDeleteConflict
	}
	seen := make(map[uuid.UUID]struct{}, len(returned))
	for _, id := range returned {
		if _, ok := expected[id]; !ok {
			return ErrCourseNodeDeleteConflict
		}
		if _, duplicate := seen[id]; duplicate {
			return ErrCourseNodeDeleteConflict
		}
		seen[id] = struct{}{}
	}
	for id := range expected {
		if _, ok := seen[id]; !ok {
			return ErrCourseNodeDeleteConflict
		}
	}
	return nil
}

func ensureMovePositionAvailable(transaction *goqu.TxDatabase, courseID uuid.UUID, parentID uuid.NullUUID, position int, movedNodeID uuid.UUID) error {
	query := transaction.From(CourseNodesTable).
		Select("id").
		Where(goqu.Ex{"course_id": courseID, "position": position}).
		Where(goqu.I("id").Neq(movedNodeID)).
		ForUpdate(goqu.Wait).
		Prepared(true)
	if parentID.Valid {
		query = query.Where(goqu.Ex{"parent_id": parentID.UUID})
	} else {
		query = query.Where(goqu.Ex{"parent_id": nil})
	}
	found, err := query.Limit(1).ScanVal(new(uuid.UUID))
	if err != nil {
		return newCourseNodePersistenceError(err)
	}
	if found {
		return ErrCourseNodePositionConflict
	}
	return nil
}

func rewriteCourseNodeSubtreePaths(transaction *goqu.TxDatabase, courseID uuid.UUID, oldRootPath, newRootPath string) (map[uuid.UUID]struct{}, error) {
	updatedIDs := make([]uuid.UUID, 0)
	err := transaction.ScanVals(
		&updatedIDs,
		courseNodeSubtreePathRewriteQuery,
		courseID,
		oldRootPath,
		newRootPath,
		courseNodeDescendantPathPattern(oldRootPath),
	)
	if err != nil {
		return nil, mapCourseNodeWriteError(err)
	}
	updated := make(map[uuid.UUID]struct{}, len(updatedIDs))
	for _, id := range updatedIDs {
		updated[id] = struct{}{}
	}
	return updated, nil
}

func sameCourseNodeParent(left, right uuid.NullUUID) bool {
	return left.Valid == right.Valid && (!left.Valid || left.UUID == right.UUID)
}

func nullableCourseNodeParent(parentID uuid.NullUUID) interface{} {
	if !parentID.Valid {
		return nil
	}
	return parentID.UUID
}

func courseNodeDescendantPathPattern(rootPath string) string {
	return rootPath + "/%"
}

func isCourseNodePathInSubtree(path, rootPath string) bool {
	return path == rootPath || strings.HasPrefix(path, rootPath+"/")
}

const courseNodeSubtreeLockQuery = `
SELECT id, course_id, parent_id, node_type, title, position, path, status, created_at, updated_at
FROM course_nodes
WHERE course_id = $1 AND (path = $2 OR path LIKE $3)
ORDER BY id
FOR UPDATE`

const courseNodeSubtreeDeleteLockQuery = `
SELECT id, course_id, parent_id, node_type, title, position, path, status, created_at, updated_at
FROM course_nodes
WHERE course_id = $1 AND (path = $2 OR path LIKE $3)
ORDER BY path ASC
FOR UPDATE`

const courseNodeSubtreeDeleteQuery = `
DELETE FROM course_nodes
WHERE course_id = $1 AND (path = $2 OR path LIKE $3)
RETURNING id`

const courseNodeSubtreePathRewriteQuery = `
UPDATE course_nodes
SET path = CASE
		WHEN path = $2 THEN $3
		ELSE $3 || substring(path FROM char_length($2) + 1)
	END,
	updated_at = now()
WHERE course_id = $1 AND (path = $2 OR path LIKE $4)
RETURNING id`

// ListRootNodes returns top-level nodes in deterministic sibling order.
func (model *CourseNodeModel) ListRootNodes(courseID uuid.UUID) ([]CourseNode, error) {
	if err := model.ensureCourseExists(courseID); err != nil {
		return nil, err
	}

	nodes := make([]CourseNode, 0)
	err := model.db.From(CourseNodesTable).
		Select(courseNodeColumns...).
		Where(goqu.Ex{"course_id": courseID, "parent_id": nil}).
		Order(goqu.I("position").Asc(), goqu.I("id").Asc()).
		Prepared(true).
		ScanStructs(&nodes)
	if err != nil {
		return nil, newCourseNodePersistenceError(err)
	}
	return nodes, nil
}

// ListChildren returns a parent's direct children in deterministic sibling order.
func (model *CourseNodeModel) ListChildren(courseID, parentID uuid.UUID) ([]CourseNode, error) {
	if err := model.ensureCourseExists(courseID); err != nil {
		return nil, err
	}
	if parentID == uuid.Nil {
		return nil, ErrCourseNodeParentNotFound
	}
	if err := model.ensureParentInCourse(courseID, parentID); err != nil {
		return nil, err
	}

	nodes := make([]CourseNode, 0)
	err := model.db.From(CourseNodesTable).
		Select(courseNodeColumns...).
		Where(goqu.Ex{"course_id": courseID, "parent_id": parentID}).
		Order(goqu.I("position").Asc(), goqu.I("id").Asc()).
		Prepared(true).
		ScanStructs(&nodes)
	if err != nil {
		return nil, newCourseNodePersistenceError(err)
	}
	return nodes, nil
}

// GetHierarchy returns the complete Course tree in deterministic preorder.
// The CTE includes the total Course node count so malformed disconnected data
// cannot be silently returned as a partial hierarchy.
func (model *CourseNodeModel) GetHierarchy(courseID uuid.UUID) (CourseHierarchy, error) {
	hierarchy := CourseHierarchy{CourseID: courseID, Roots: make([]CourseNodeHierarchy, 0)}
	if err := model.ensureCourseExists(courseID); err != nil {
		return hierarchy, err
	}

	rows := make([]courseNodeHierarchyRow, 0)
	err := model.db.ScanStructs(&rows, courseHierarchyQuery, courseID)
	if err != nil {
		return hierarchy, newCourseNodePersistenceError(err)
	}

	buildNodes := make(map[uuid.UUID]*courseNodeHierarchyBuild)
	roots := make([]*courseNodeHierarchyBuild, 0)
	var totalNodes int64
	for _, row := range rows {
		totalNodes = row.TotalNodes
		if !row.ID.Valid {
			continue
		}
		node := row.courseNode()
		current := &courseNodeHierarchyBuild{node: node, children: make([]*courseNodeHierarchyBuild, 0)}
		if _, exists := buildNodes[node.ID]; exists {
			return hierarchy, ErrCourseNodeHierarchyIntegrity
		}
		buildNodes[node.ID] = current
		if !node.ParentID.Valid {
			roots = append(roots, current)
			continue
		}
		parent, exists := buildNodes[node.ParentID.UUID]
		if !exists {
			return hierarchy, ErrCourseNodeHierarchyIntegrity
		}
		parent.children = append(parent.children, current)
	}
	if int64(len(buildNodes)) != totalNodes {
		return hierarchy, ErrCourseNodeHierarchyIntegrity
	}
	for _, root := range roots {
		hierarchy.Roots = append(hierarchy.Roots, root.public())
	}
	return hierarchy, nil
}

func (model *CourseNodeModel) ensureCourseExists(courseID uuid.UUID) error {
	if courseID == uuid.Nil {
		return ErrCourseNodeCourseRequired
	}
	found, err := model.db.From(CoursesTable).
		Select("id").
		Where(goqu.Ex{"id": courseID}).
		Limit(1).
		Prepared(true).
		ScanVal(new(uuid.UUID))
	if err != nil {
		return newCourseNodePersistenceError(err)
	}
	if !found {
		return ErrCourseNodeCourseNotFound
	}
	return nil
}

func (model *CourseNodeModel) ensureParentInCourse(courseID, parentID uuid.UUID) error {
	var parent struct {
		ID       uuid.UUID `db:"id"`
		CourseID uuid.UUID `db:"course_id"`
	}
	found, err := model.db.From(CourseNodesTable).
		Select("id", "course_id").
		Where(goqu.Ex{"id": parentID}).
		Limit(1).
		Prepared(true).
		ScanStruct(&parent)
	if err != nil {
		return newCourseNodePersistenceError(err)
	}
	if !found {
		return ErrCourseNodeParentNotFound
	}
	if parent.CourseID != courseID {
		return ErrCourseNodeCrossCourseParent
	}
	return nil
}

const courseHierarchyQuery = `
WITH RECURSIVE course_rows AS (
	SELECT id, course_id, parent_id, node_type, title, position, path, status, created_at, updated_at
	FROM course_nodes
	WHERE course_id = $1
), tree AS (
	SELECT id, course_id, parent_id, node_type, title, position, path, status, created_at, updated_at,
		ARRAY[id] AS ancestry,
		ARRAY[lpad(position::text, 10, '0') || ':' || id::text] AS sort_key
	FROM course_rows
	WHERE parent_id IS NULL
	UNION ALL
	SELECT child.id, child.course_id, child.parent_id, child.node_type, child.title, child.position, child.path, child.status, child.created_at, child.updated_at,
		tree.ancestry || child.id,
		tree.sort_key || (lpad(child.position::text, 10, '0') || ':' || child.id::text)
	FROM course_rows child
	JOIN tree ON child.parent_id = tree.id
	WHERE NOT child.id = ANY(tree.ancestry)
), totals AS (
	SELECT count(*)::bigint AS total_nodes FROM course_rows
)
SELECT tree.id, tree.course_id, tree.parent_id, tree.node_type, tree.title, tree.position, tree.path, tree.status, tree.created_at, tree.updated_at, totals.total_nodes
FROM totals
LEFT JOIN tree ON TRUE
ORDER BY tree.sort_key`

type courseNodeHierarchyRow struct {
	ID         uuid.NullUUID  `db:"id"`
	CourseID   uuid.NullUUID  `db:"course_id"`
	ParentID   uuid.NullUUID  `db:"parent_id"`
	NodeType   sql.NullString `db:"node_type"`
	Title      sql.NullString `db:"title"`
	Position   sql.NullInt32  `db:"position"`
	Path       sql.NullString `db:"path"`
	Status     sql.NullString `db:"status"`
	CreatedAt  sql.NullTime   `db:"created_at"`
	UpdatedAt  sql.NullTime   `db:"updated_at"`
	TotalNodes int64          `db:"total_nodes"`
}

func (row courseNodeHierarchyRow) courseNode() CourseNode {
	return CourseNode{
		ID:        row.ID.UUID,
		CourseID:  row.CourseID.UUID,
		ParentID:  row.ParentID,
		NodeType:  CourseNodeType(row.NodeType.String),
		Title:     row.Title.String,
		Position:  int(row.Position.Int32),
		Path:      row.Path.String,
		Status:    CourseStatus(row.Status.String),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}

type courseNodeHierarchyBuild struct {
	node     CourseNode
	children []*courseNodeHierarchyBuild
}

func (node *courseNodeHierarchyBuild) public() CourseNodeHierarchy {
	result := CourseNodeHierarchy{Node: node.node, Children: make([]CourseNodeHierarchy, 0, len(node.children))}
	for _, child := range node.children {
		result.Children = append(result.Children, child.public())
	}
	return result
}

func validateCreateCourseNodeParams(params CreateCourseNodeParams) (string, error) {
	if params.CourseID == uuid.Nil {
		return "", ErrCourseNodeCourseRequired
	}
	title := strings.TrimSpace(params.Title)
	if title == "" {
		return "", ErrCourseNodeTitleRequired
	}
	switch params.NodeType {
	case CourseNodeTypeSection, CourseNodeTypeSubject, CourseNodeTypeTopic:
	default:
		return "", ErrCourseNodeTypeInvalid
	}
	if params.Position < 0 {
		return "", ErrCourseNodePositionInvalid
	}
	if params.ParentID.Valid && params.ParentID.UUID == uuid.Nil {
		return "", ErrCourseNodeParentNotFound
	}
	return title, nil
}

func encodeCourseNodePath(parentPath string, nodeID uuid.UUID) string {
	nodePart := nodeID.String()
	if parentPath == "" {
		return nodePart
	}
	return parentPath + "/" + nodePart
}

func mapCourseNodeWriteError(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		switch pqErr.Constraint {
		case courseNodesTopLevelPositionConstraint, courseNodesChildPositionConstraint:
			return newCourseNodeSemanticError(ErrCourseNodePositionConflict, err)
		}
	}
	return newCourseNodePersistenceError(err)
}

type courseNodeError struct {
	kind  error
	cause error
}

func (err *courseNodeError) Error() string {
	return err.kind.Error()
}

func (err *courseNodeError) Unwrap() []error {
	return []error{err.kind, err.cause}
}

func newCourseNodePersistenceError(cause error) error {
	return &courseNodeError{kind: ErrCourseNodePersistence, cause: cause}
}

func newCourseNodeSemanticError(kind, cause error) error {
	return &courseNodeError{kind: kind, cause: cause}
}
