package models

import (
	"bytes"
	"errors"
	"math"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func learningItemMoveTestIDs() (courseID, sourceNodeID, destNodeID, aID, bID, cID, dID, foreignID uuid.UUID) {
	return uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481"),
		uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280"), // source (higher UUID)
		uuid.MustParse("019c01c7-1111-7222-8333-944455556666"), // dest (lower UUID — locks first)
		uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6"),
		uuid.MustParse("019c01ca-7c0e-79ac-aab3-f320f4161b7a"),
		uuid.MustParse("019c01cb-7f39-7f26-900f-6947e75e7284"),
		uuid.MustParse("019c01cd-2222-7333-8444-955566667777"),
		uuid.MustParse("019c01cc-1111-7222-8333-944455556666")
}

func expectLearningItemMoveCourseNodeUpdate(
	mock sqlmock.Sqlmock,
	courseID, currentNodeID, newNodeID, itemID uuid.UUID,
) {
	mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
		WithArgs(newNodeID, courseID, currentNodeID, itemID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(itemID))
}

func expectLearningItemMoveNodeLocksUUIDOrder(
	mock sqlmock.Sqlmock,
	courseID, sourceNodeID, destNodeID uuid.UUID,
	sourceExists, destExists bool,
) {
	firstID, secondID := sourceNodeID, destNodeID
	firstExists, secondExists := sourceExists, destExists
	if bytes.Compare(destNodeID[:], sourceNodeID[:]) < 0 {
		firstID, secondID = destNodeID, sourceNodeID
		firstExists, secondExists = destExists, sourceExists
	}
	expectLearningItemNodeLock(mock, courseID, firstID, firstExists)
	if !firstExists {
		return
	}
	expectLearningItemNodeLock(mock, courseID, secondID, secondExists)
}

func expectLearningItemMoveSiblingLocksUUIDOrder(
	mock sqlmock.Sqlmock,
	courseID, sourceNodeID, destNodeID uuid.UUID,
	sourceRows, destRows *sqlmock.Rows,
) {
	firstID, secondID := sourceNodeID, destNodeID
	firstRows, secondRows := sourceRows, destRows
	if bytes.Compare(destNodeID[:], sourceNodeID[:]) < 0 {
		firstID, secondID = destNodeID, sourceNodeID
		firstRows, secondRows = destRows, sourceRows
	}
	expectLearningItemReorderSiblingLock(mock, courseID, firstID, firstRows)
	expectLearningItemReorderSiblingLock(mock, courseID, secondID, secondRows)
}

func siblingItem(id uuid.UUID, position int) struct {
	id       uuid.UUID
	position int
} {
	return struct {
		id       uuid.UUID
		position int
	}{id, position}
}

func TestVerifyLearningItemMoveResult(t *testing.T) {
	sourceNodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	destNodeID := uuid.MustParse("019c01c7-1111-7222-8333-944455556666")
	a := uuid.MustParse("019c01c9-b06d-7d4f-8c70-d35f7344bdd6")
	b := uuid.MustParse("019c01ca-7c0e-79ac-aab3-f320f4161b7a")

	remaining := []LearningItem{{ID: a, CourseNodeID: sourceNodeID, Position: 0}}
	finalDest := []LearningItem{
		{ID: b, CourseNodeID: destNodeID, Position: 0},
	}
	sourceAfter := []LearningItem{{ID: a, CourseNodeID: sourceNodeID, Position: 0}}
	destAfter := []LearningItem{{ID: b, CourseNodeID: destNodeID, Position: 0}}

	if err := verifyLearningItemMoveResult(
		sourceAfter, destAfter, remaining, finalDest, sourceNodeID, destNodeID,
	); err != nil {
		t.Fatalf("exact match: %v", err)
	}
	if err := verifyLearningItemMoveResult(
		sourceAfter, destAfter, remaining, finalDest, sourceNodeID, sourceNodeID,
	); !errors.Is(err, ErrLearningItemMoveConflict) {
		t.Fatalf("wrong dest ownership: got %v", err)
	}
	badSource := []LearningItem{{ID: a, CourseNodeID: destNodeID, Position: 0}}
	if err := verifyLearningItemMoveResult(
		badSource, destAfter, remaining, finalDest, sourceNodeID, destNodeID,
	); !errors.Is(err, ErrLearningItemMoveConflict) {
		t.Fatalf("wrong source ownership: got %v", err)
	}
	if err := verifyLearningItemMoveResult(
		[]LearningItem{}, destAfter, remaining, finalDest, sourceNodeID, destNodeID,
	); !errors.Is(err, ErrLearningItemMoveConflict) {
		t.Fatalf("count mismatch: got %v", err)
	}
}

func TestMoveLearningItemsValidation(t *testing.T) {
	model, mock := newLearningItemModelTest(t)
	_, sourceNodeID, destNodeID, aID, bID, _, _, _ := learningItemMoveTestIDs()
	tests := []struct {
		name    string
		course  uuid.UUID
		source  uuid.UUID
		dest    uuid.UUID
		ordered []uuid.UUID
		want    error
	}{
		{"course required", uuid.Nil, sourceNodeID, destNodeID, nil, ErrCourseNotFound},
		{"source required", uuid.New(), uuid.Nil, destNodeID, []uuid.UUID{aID}, ErrLearningItemNodeNotFound},
		{"dest required", uuid.New(), sourceNodeID, uuid.Nil, []uuid.UUID{aID}, ErrLearningItemNodeNotFound},
		{"same node", uuid.New(), sourceNodeID, sourceNodeID, []uuid.UUID{aID}, ErrLearningItemMoveSameNode},
		{"nil ordered id", uuid.New(), sourceNodeID, destNodeID, []uuid.UUID{uuid.Nil}, ErrLearningItemMoveMismatch},
		{"duplicate ordered id", uuid.New(), sourceNodeID, destNodeID, []uuid.UUID{aID, aID}, ErrLearningItemMoveDuplicate},
		{"duplicate among three", uuid.New(), sourceNodeID, destNodeID, []uuid.UUID{aID, bID, aID}, ErrLearningItemMoveDuplicate},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := model.MoveLearningItems(test.course, test.source, test.dest, test.ordered)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("validation unexpectedly queried the database: %v", err)
	}
}

func TestMoveLearningItemsSubsetMismatch(t *testing.T) {
	courseID, sourceNodeID, destNodeID, aID, bID, _, _, foreignID := learningItemMoveTestIDs()

	t.Run("missing from source", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectBegin()
		expectLearningItemReorderCourseLock(mock, courseID)
		expectLearningItemMoveNodeLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID, true, true)
		expectLearningItemMoveSiblingLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID,
			learningItemSiblingRows(courseID, sourceNodeID, siblingItem(aID, 0)),
			learningItemSiblingRows(courseID, destNodeID),
		)
		mock.ExpectRollback()
		_, err := model.MoveLearningItems(courseID, sourceNodeID, destNodeID, []uuid.UUID{aID, bID})
		if !errors.Is(err, ErrLearningItemMoveMismatch) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemMoveMismatch)
		}
	})

	t.Run("foreign node id", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectBegin()
		expectLearningItemReorderCourseLock(mock, courseID)
		expectLearningItemMoveNodeLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID, true, true)
		expectLearningItemMoveSiblingLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID,
			learningItemSiblingRows(courseID, sourceNodeID, siblingItem(aID, 0), siblingItem(bID, 1)),
			learningItemSiblingRows(courseID, destNodeID),
		)
		mock.ExpectRollback()
		_, err := model.MoveLearningItems(courseID, sourceNodeID, destNodeID, []uuid.UUID{foreignID})
		if !errors.Is(err, ErrLearningItemMoveMismatch) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemMoveMismatch)
		}
	})

	t.Run("foreign course id treated as mismatch", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectBegin()
		expectLearningItemReorderCourseLock(mock, courseID)
		expectLearningItemMoveNodeLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID, true, true)
		expectLearningItemMoveSiblingLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID,
			learningItemSiblingRows(courseID, sourceNodeID, siblingItem(aID, 0)),
			learningItemSiblingRows(courseID, destNodeID),
		)
		mock.ExpectRollback()
		// ID not present on source siblings — no existence probe of other courses/nodes.
		_, err := model.MoveLearningItems(courseID, sourceNodeID, destNodeID, []uuid.UUID{bID})
		if !errors.Is(err, ErrLearningItemMoveMismatch) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemMoveMismatch)
		}
	})
}

func TestMoveLearningItemsMissingNodes(t *testing.T) {
	courseID, sourceNodeID, destNodeID, aID, _, _, _, _ := learningItemMoveTestIDs()

	t.Run("source missing", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectBegin()
		expectLearningItemReorderCourseLock(mock, courseID)
		// dest UUID < source UUID ⇒ dest locked first and succeeds; source missing next.
		expectLearningItemNodeLock(mock, courseID, destNodeID, true)
		expectLearningItemNodeLock(mock, courseID, sourceNodeID, false)
		// Second unscoped probe distinguishes not-found from cross-course.
		mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
			WithArgs(sourceNodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}))
		mock.ExpectRollback()
		_, err := model.MoveLearningItems(courseID, sourceNodeID, destNodeID, []uuid.UUID{aID})
		if !errors.Is(err, ErrLearningItemNodeNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNodeNotFound)
		}
	})

	t.Run("dest missing", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectBegin()
		expectLearningItemReorderCourseLock(mock, courseID)
		expectLearningItemNodeLock(mock, courseID, destNodeID, false)
		mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
			WithArgs(destNodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}))
		mock.ExpectRollback()
		_, err := model.MoveLearningItems(courseID, sourceNodeID, destNodeID, []uuid.UUID{aID})
		if !errors.Is(err, ErrLearningItemNodeNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNodeNotFound)
		}
	})
}

func TestMoveLearningItemsEmptyNoopWithRealCounts(t *testing.T) {
	courseID, sourceNodeID, destNodeID, aID, bID, cID, _, _ := learningItemMoveTestIDs()
	model, mock := newLearningItemModelTest(t)
	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemMoveNodeLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID, true, true)
	expectLearningItemMoveSiblingLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID,
		learningItemSiblingRows(courseID, sourceNodeID, siblingItem(aID, 0), siblingItem(bID, 1)),
		learningItemSiblingRows(courseID, destNodeID, siblingItem(cID, 0)),
	)
	mock.ExpectCommit()

	result, err := model.MoveLearningItems(courseID, sourceNodeID, destNodeID, nil)
	if err != nil {
		t.Fatalf("empty noop: %v", err)
	}
	if !result.Noop || result.ItemsMoved != 0 {
		t.Fatalf("result = %+v", result)
	}
	if result.SourceItemCount != 2 || result.DestinationItemCount != 1 {
		t.Fatalf("counts must reflect locked siblings, got source=%d dest=%d",
			result.SourceItemCount, result.DestinationItemCount)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("empty noop must not UPDATE: %v", err)
	}
}

func TestMoveLearningItemsEmptyDestinationSuccess(t *testing.T) {
	courseID, sourceNodeID, destNodeID, aID, bID, _, _, _ := learningItemMoveTestIDs()
	model, mock := newLearningItemModelTest(t)

	// maxPosition=1, sourceCount=2, destCount=0 → sourceTempBase=4, destTempBase=6
	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemMoveNodeLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID, true, true)
	expectLearningItemMoveSiblingLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID,
		learningItemSiblingRows(courseID, sourceNodeID, siblingItem(aID, 0), siblingItem(bID, 1)),
		learningItemSiblingRows(courseID, destNodeID),
	)
	// 1) stage source temps
	expectLearningItemReorderPositionUpdate(mock, courseID, sourceNodeID, aID, 4)
	expectLearningItemReorderPositionUpdate(mock, courseID, sourceNodeID, bID, 5)
	// 2) move b
	expectLearningItemMoveCourseNodeUpdate(mock, courseID, sourceNodeID, destNodeID, bID)
	// 3) final source
	expectLearningItemReorderPositionUpdate(mock, courseID, sourceNodeID, aID, 0)
	// 4) final dest
	expectLearningItemReorderPositionUpdate(mock, courseID, destNodeID, bID, 0)
	// 5) verify re-lock (source then dest)
	expectLearningItemReorderSiblingLock(mock, courseID, sourceNodeID,
		learningItemSiblingRows(courseID, sourceNodeID, siblingItem(aID, 0)))
	expectLearningItemReorderSiblingLock(mock, courseID, destNodeID,
		learningItemSiblingRows(courseID, destNodeID, siblingItem(bID, 0)))
	mock.ExpectCommit()

	result, err := model.MoveLearningItems(courseID, sourceNodeID, destNodeID, []uuid.UUID{bID})
	if err != nil {
		t.Fatalf("empty dest move: %v", err)
	}
	if result.Noop || result.ItemsMoved != 1 {
		t.Fatalf("result = %+v", result)
	}
	if result.SourceItemCount != 1 || result.DestinationItemCount != 1 {
		t.Fatalf("post-move counts = %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMoveLearningItemsSingleItem(t *testing.T) {
	courseID, sourceNodeID, destNodeID, aID, bID, cID, _, _ := learningItemMoveTestIDs()
	model, mock := newLearningItemModelTest(t)

	// source [A@0,B@1], dest [C@0]; move [B]
	// maxPosition=1, sourceCount=2, destCount=1 → sourceTempBase=5, destTempBase=7
	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemMoveNodeLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID, true, true)
	expectLearningItemMoveSiblingLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID,
		learningItemSiblingRows(courseID, sourceNodeID, siblingItem(aID, 0), siblingItem(bID, 1)),
		learningItemSiblingRows(courseID, destNodeID, siblingItem(cID, 0)),
	)
	expectLearningItemReorderPositionUpdate(mock, courseID, sourceNodeID, aID, 5)
	expectLearningItemReorderPositionUpdate(mock, courseID, sourceNodeID, bID, 6)
	expectLearningItemReorderPositionUpdate(mock, courseID, destNodeID, cID, 7)
	expectLearningItemMoveCourseNodeUpdate(mock, courseID, sourceNodeID, destNodeID, bID)
	expectLearningItemReorderPositionUpdate(mock, courseID, sourceNodeID, aID, 0)
	expectLearningItemReorderPositionUpdate(mock, courseID, destNodeID, cID, 0)
	expectLearningItemReorderPositionUpdate(mock, courseID, destNodeID, bID, 1)
	expectLearningItemReorderSiblingLock(mock, courseID, sourceNodeID,
		learningItemSiblingRows(courseID, sourceNodeID, siblingItem(aID, 0)))
	expectLearningItemReorderSiblingLock(mock, courseID, destNodeID,
		learningItemSiblingRows(courseID, destNodeID, siblingItem(cID, 0), siblingItem(bID, 1)))
	mock.ExpectCommit()

	result, err := model.MoveLearningItems(courseID, sourceNodeID, destNodeID, []uuid.UUID{bID})
	if err != nil {
		t.Fatalf("single move: %v", err)
	}
	if result.Noop || result.ItemsMoved != 1 || result.SourceItemCount != 1 || result.DestinationItemCount != 2 {
		t.Fatalf("result = %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMoveLearningItemsMultiItemCanonicalRenumber(t *testing.T) {
	courseID, sourceNodeID, destNodeID, aID, bID, cID, dID, _ := learningItemMoveTestIDs()
	model, mock := newLearningItemModelTest(t)

	// source [A@0,B@1,C@2], dest [D@0]; move [C,A] → remaining [B], dest [D,C,A]
	// maxPosition=2, sourceCount=3, destCount=1 → sourceTempBase=7, destTempBase=10
	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemMoveNodeLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID, true, true)
	expectLearningItemMoveSiblingLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID,
		learningItemSiblingRows(courseID, sourceNodeID,
			siblingItem(aID, 0), siblingItem(bID, 1), siblingItem(cID, 2)),
		learningItemSiblingRows(courseID, destNodeID, siblingItem(dID, 0)),
	)
	// stage source then dest (disjoint temps — no direct collide)
	expectLearningItemReorderPositionUpdate(mock, courseID, sourceNodeID, aID, 7)
	expectLearningItemReorderPositionUpdate(mock, courseID, sourceNodeID, bID, 8)
	expectLearningItemReorderPositionUpdate(mock, courseID, sourceNodeID, cID, 9)
	expectLearningItemReorderPositionUpdate(mock, courseID, destNodeID, dID, 10)
	// ownership change while still at source temps
	expectLearningItemMoveCourseNodeUpdate(mock, courseID, sourceNodeID, destNodeID, cID)
	expectLearningItemMoveCourseNodeUpdate(mock, courseID, sourceNodeID, destNodeID, aID)
	// final source 0..n-1
	expectLearningItemReorderPositionUpdate(mock, courseID, sourceNodeID, bID, 0)
	// final dest 0..m-1
	expectLearningItemReorderPositionUpdate(mock, courseID, destNodeID, dID, 0)
	expectLearningItemReorderPositionUpdate(mock, courseID, destNodeID, cID, 1)
	expectLearningItemReorderPositionUpdate(mock, courseID, destNodeID, aID, 2)
	expectLearningItemReorderSiblingLock(mock, courseID, sourceNodeID,
		learningItemSiblingRows(courseID, sourceNodeID, siblingItem(bID, 0)))
	expectLearningItemReorderSiblingLock(mock, courseID, destNodeID,
		learningItemSiblingRows(courseID, destNodeID,
			siblingItem(dID, 0), siblingItem(cID, 1), siblingItem(aID, 2)))
	mock.ExpectCommit()

	result, err := model.MoveLearningItems(courseID, sourceNodeID, destNodeID, []uuid.UUID{cID, aID})
	if err != nil {
		t.Fatalf("multi move: %v", err)
	}
	if result.Noop || result.ItemsMoved != 2 {
		t.Fatalf("result = %+v", result)
	}
	if result.SourceItemCount != 1 || result.DestinationItemCount != 3 {
		t.Fatalf("post-move counts = %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMoveLearningItemsUUIDLockOrderDestLower(t *testing.T) {
	courseID, sourceNodeID, destNodeID, aID, _, _, _, _ := learningItemMoveTestIDs()
	if bytes.Compare(destNodeID[:], sourceNodeID[:]) >= 0 {
		t.Fatal("test fixture requires dest UUID < source UUID")
	}
	model, mock := newLearningItemModelTest(t)
	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	// Explicit order: dest node, then source node, then dest siblings, then source siblings.
	expectLearningItemNodeLock(mock, courseID, destNodeID, true)
	expectLearningItemNodeLock(mock, courseID, sourceNodeID, true)
	expectLearningItemReorderSiblingLock(mock, courseID, destNodeID, learningItemSiblingRows(courseID, destNodeID))
	expectLearningItemReorderSiblingLock(mock, courseID, sourceNodeID,
		learningItemSiblingRows(courseID, sourceNodeID, siblingItem(aID, 0)))
	mock.ExpectCommit()

	result, err := model.MoveLearningItems(courseID, sourceNodeID, destNodeID, nil)
	if err != nil {
		t.Fatalf("uuid lock order noop: %v", err)
	}
	if !result.Noop || result.SourceItemCount != 1 || result.DestinationItemCount != 0 {
		t.Fatalf("result = %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMoveLearningItemsConcurrentConflict(t *testing.T) {
	courseID, sourceNodeID, destNodeID, aID, bID, _, _, _ := learningItemMoveTestIDs()
	model, mock := newLearningItemModelTest(t)
	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemMoveNodeLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID, true, true)
	expectLearningItemMoveSiblingLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID,
		learningItemSiblingRows(courseID, sourceNodeID, siblingItem(aID, 0), siblingItem(bID, 1)),
		learningItemSiblingRows(courseID, destNodeID),
	)
	mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
		WithArgs(4, courseID, sourceNodeID, aID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectRollback()

	_, err := model.MoveLearningItems(courseID, sourceNodeID, destNodeID, []uuid.UUID{bID})
	if !errors.Is(err, ErrLearningItemMoveConflict) {
		t.Fatalf("error = %v, want %v", err, ErrLearningItemMoveConflict)
	}
}

func TestMoveLearningItemsUniqueViolationConflict(t *testing.T) {
	courseID, sourceNodeID, destNodeID, aID, bID, _, _, _ := learningItemMoveTestIDs()
	model, mock := newLearningItemModelTest(t)
	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemMoveNodeLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID, true, true)
	expectLearningItemMoveSiblingLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID,
		learningItemSiblingRows(courseID, sourceNodeID, siblingItem(aID, 0), siblingItem(bID, 1)),
		learningItemSiblingRows(courseID, destNodeID),
	)
	mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
		WithArgs(4, courseID, sourceNodeID, aID).
		WillReturnError(&pq.Error{Code: "23505", Constraint: learningItemsNodePositionConstraint})
	mock.ExpectRollback()

	_, err := model.MoveLearningItems(courseID, sourceNodeID, destNodeID, []uuid.UUID{bID})
	if !errors.Is(err, ErrLearningItemMoveConflict) {
		t.Fatalf("error = %v, want %v", err, ErrLearningItemMoveConflict)
	}
}

func TestMoveLearningItemsRollbackOnInjectedFailure(t *testing.T) {
	courseID, sourceNodeID, destNodeID, aID, bID, _, _, _ := learningItemMoveTestIDs()
	model, mock := newLearningItemModelTest(t)
	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemMoveNodeLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID, true, true)
	expectLearningItemMoveSiblingLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID,
		learningItemSiblingRows(courseID, sourceNodeID, siblingItem(aID, 0), siblingItem(bID, 1)),
		learningItemSiblingRows(courseID, destNodeID),
	)
	mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
		WithArgs(4, courseID, sourceNodeID, aID).
		WillReturnError(errors.New("update failure"))
	mock.ExpectRollback()

	_, err := model.MoveLearningItems(courseID, sourceNodeID, destNodeID, []uuid.UUID{bID})
	if !errors.Is(err, ErrLearningItemPersistence) {
		t.Fatalf("error = %v, want persistence wrapper", err)
	}
}

func TestMoveLearningItemsPositionOverflow(t *testing.T) {
	courseID, sourceNodeID, destNodeID, aID, _, _, _, _ := learningItemMoveTestIDs()
	model, mock := newLearningItemModelTest(t)
	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemMoveNodeLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID, true, true)
	expectLearningItemMoveSiblingLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID,
		learningItemSiblingRows(courseID, sourceNodeID, siblingItem(aID, math.MaxInt32)),
		learningItemSiblingRows(courseID, destNodeID),
	)
	mock.ExpectRollback()

	_, err := model.MoveLearningItems(courseID, sourceNodeID, destNodeID, []uuid.UUID{aID})
	if !errors.Is(err, ErrLearningItemPositionInvalid) {
		t.Fatalf("error = %v, want %v", err, ErrLearningItemPositionInvalid)
	}
}

func TestMoveLearningItemsStagingOrderNoDirectCollide(t *testing.T) {
	// Confirms both nodes stage to disjoint temps before any course_node_id change.
	courseID, sourceNodeID, destNodeID, aID, bID, cID, _, _ := learningItemMoveTestIDs()
	model, mock := newLearningItemModelTest(t)

	mock.ExpectBegin()
	expectLearningItemReorderCourseLock(mock, courseID)
	expectLearningItemMoveNodeLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID, true, true)
	expectLearningItemMoveSiblingLocksUUIDOrder(mock, courseID, sourceNodeID, destNodeID,
		learningItemSiblingRows(courseID, sourceNodeID, siblingItem(aID, 0), siblingItem(bID, 1)),
		learningItemSiblingRows(courseID, destNodeID, siblingItem(cID, 0)),
	)
	// All staging UPDATEs precede ownership UPDATE (sourceTempBase=5, destTempBase=7).
	expectLearningItemReorderPositionUpdate(mock, courseID, sourceNodeID, aID, 5)
	expectLearningItemReorderPositionUpdate(mock, courseID, sourceNodeID, bID, 6)
	expectLearningItemReorderPositionUpdate(mock, courseID, destNodeID, cID, 7)
	expectLearningItemMoveCourseNodeUpdate(mock, courseID, sourceNodeID, destNodeID, aID)
	expectLearningItemReorderPositionUpdate(mock, courseID, sourceNodeID, bID, 0)
	expectLearningItemReorderPositionUpdate(mock, courseID, destNodeID, cID, 0)
	expectLearningItemReorderPositionUpdate(mock, courseID, destNodeID, aID, 1)
	expectLearningItemReorderSiblingLock(mock, courseID, sourceNodeID,
		learningItemSiblingRows(courseID, sourceNodeID, siblingItem(bID, 0)))
	expectLearningItemReorderSiblingLock(mock, courseID, destNodeID,
		learningItemSiblingRows(courseID, destNodeID, siblingItem(cID, 0), siblingItem(aID, 1)))
	mock.ExpectCommit()

	_, err := model.MoveLearningItems(courseID, sourceNodeID, destNodeID, []uuid.UUID{aID})
	if err != nil {
		t.Fatalf("staging order move: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
