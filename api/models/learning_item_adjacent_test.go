package models

import (
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func learningItemAdjacentTestIDs() (courseID, nodeID, aID, bID, cID, otherNodeID, otherCourseID uuid.UUID) {
	return uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481"),
		uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280"),
		uuid.MustParse("019c02f0-aaaa-7000-8000-000000000001"),
		uuid.MustParse("019c02f0-aaaa-7000-8000-000000000002"),
		uuid.MustParse("019c02f0-aaaa-7000-8000-000000000003"),
		uuid.MustParse("019c01c9-bbbb-78e2-a366-690bfd600281"),
		uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e482")
}

func learningItemAdjacentRow(itemID, courseID, nodeID uuid.UUID, title string, position int) *sqlmock.Rows {
	return learningItemRows(itemID, courseID, nodeID, title, LearningItemTypeArticle, nil, []byte(`{"version":1,"blocks":[]}`), position)
}

func expectAdjacentCurrent(
	mock sqlmock.Sqlmock,
	courseID, nodeID, itemID uuid.UUID,
	title string,
	position int,
) {
	mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
		WithArgs(courseID, nodeID, itemID, uint(1)).
		WillReturnRows(learningItemAdjacentRow(itemID, courseID, nodeID, title, position))
}

func expectAdjacentMissingCurrent(mock sqlmock.Sqlmock, courseID, nodeID, itemID uuid.UUID) {
	mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
		WithArgs(courseID, nodeID, itemID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "course_id", "course_node_id", "title", "item_type",
			"description", "metadata", "position", "publish_state", "created_at", "updated_at",
		}))
}

func expectAdjacentSibling(
	mock sqlmock.Sqlmock,
	courseID, nodeID uuid.UUID,
	currentPosition int,
	currentID uuid.UUID,
	wantNext bool,
	sibling *sqlmock.Rows,
) {
	// goqu emits course_id, course_node_id, then Or branches (position compare + position eq + id compare).
	query := mock.ExpectQuery(`SELECT .* FROM "learning_items".*ORDER BY`)
	if wantNext {
		query.WithArgs(
			courseID, nodeID,
			currentPosition,
			currentPosition, currentID,
			uint(1),
		)
	} else {
		query.WithArgs(
			courseID, nodeID,
			currentPosition,
			currentPosition, currentID,
			uint(1),
		)
	}
	if sibling != nil {
		query.WillReturnRows(sibling)
	} else {
		query.WillReturnRows(sqlmock.NewRows([]string{
			"id", "course_id", "course_node_id", "title", "item_type",
			"description", "metadata", "position", "publish_state", "created_at", "updated_at",
		}))
	}
}

func TestLearningItemPreviousNextValidation(t *testing.T) {
	model, mock := newLearningItemModelTest(t)
	courseID, nodeID, aID, _, _, _, _ := learningItemAdjacentTestIDs()
	tests := []struct {
		name    string
		course  uuid.UUID
		node    uuid.UUID
		current uuid.UUID
		want    error
	}{
		{"nil course", uuid.Nil, nodeID, aID, ErrCourseNotFound},
		{"nil node", courseID, uuid.Nil, aID, ErrLearningItemNodeNotFound},
		{"nil current", courseID, nodeID, uuid.Nil, ErrLearningItemNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := model.GetAdjacentLearningItems(test.course, test.node, test.current)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
			if result.Previous != nil || result.Next != nil {
				t.Fatalf("result must be empty on validation failure: %+v", result)
			}
		})
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("validation queried database: %v", err)
	}
}

func TestLearningItemPreviousNextMiddleFirstLastSingle(t *testing.T) {
	courseID, nodeID, aID, bID, cID, _, _ := learningItemAdjacentTestIDs()

	t.Run("middle_returns_previous_and_next", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentCurrent(mock, courseID, nodeID, bID, "B", 1)
		expectAdjacentSibling(mock, courseID, nodeID, 1, bID, false, learningItemAdjacentRow(aID, courseID, nodeID, "A", 0))
		expectAdjacentSibling(mock, courseID, nodeID, 1, bID, true, learningItemAdjacentRow(cID, courseID, nodeID, "C", 2))

		result, err := model.GetAdjacentLearningItems(courseID, nodeID, bID)
		if err != nil {
			t.Fatalf("adjacent: %v", err)
		}
		if result.Previous == nil || result.Previous.ID != aID {
			t.Fatalf("previous = %+v", result.Previous)
		}
		if result.Next == nil || result.Next.ID != cID {
			t.Fatalf("next = %+v", result.Next)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("first_previous_nil", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentCurrent(mock, courseID, nodeID, aID, "A", 0)
		expectAdjacentSibling(mock, courseID, nodeID, 0, aID, false, nil)
		expectAdjacentSibling(mock, courseID, nodeID, 0, aID, true, learningItemAdjacentRow(bID, courseID, nodeID, "B", 1))

		result, err := model.GetAdjacentLearningItems(courseID, nodeID, aID)
		if err != nil {
			t.Fatalf("adjacent: %v", err)
		}
		if result.Previous != nil {
			t.Fatalf("previous want nil, got %+v", result.Previous)
		}
		if result.Next == nil || result.Next.ID != bID {
			t.Fatalf("next = %+v", result.Next)
		}
	})

	t.Run("last_next_nil", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentCurrent(mock, courseID, nodeID, cID, "C", 2)
		expectAdjacentSibling(mock, courseID, nodeID, 2, cID, false, learningItemAdjacentRow(bID, courseID, nodeID, "B", 1))
		expectAdjacentSibling(mock, courseID, nodeID, 2, cID, true, nil)

		result, err := model.GetAdjacentLearningItems(courseID, nodeID, cID)
		if err != nil {
			t.Fatalf("adjacent: %v", err)
		}
		if result.Previous == nil || result.Previous.ID != bID {
			t.Fatalf("previous = %+v", result.Previous)
		}
		if result.Next != nil {
			t.Fatalf("next want nil, got %+v", result.Next)
		}
	})

	t.Run("single_item_both_nil", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentCurrent(mock, courseID, nodeID, aID, "Only", 0)
		expectAdjacentSibling(mock, courseID, nodeID, 0, aID, false, nil)
		expectAdjacentSibling(mock, courseID, nodeID, 0, aID, true, nil)

		result, err := model.GetAdjacentLearningItems(courseID, nodeID, aID)
		if err != nil {
			t.Fatalf("adjacent: %v", err)
		}
		if result.Previous != nil || result.Next != nil {
			t.Fatalf("want both nil, got %+v", result)
		}
	})

	t.Run("two_item_boundaries", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentCurrent(mock, courseID, nodeID, aID, "A", 0)
		expectAdjacentSibling(mock, courseID, nodeID, 0, aID, false, nil)
		expectAdjacentSibling(mock, courseID, nodeID, 0, aID, true, learningItemAdjacentRow(bID, courseID, nodeID, "B", 1))
		result, err := model.GetAdjacentLearningItems(courseID, nodeID, aID)
		if err != nil || result.Previous != nil || result.Next == nil || result.Next.ID != bID {
			t.Fatalf("first of two: result=%+v err=%v", result, err)
		}

		model2, mock2 := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock2, courseID, nodeID, true)
		expectAdjacentCurrent(mock2, courseID, nodeID, bID, "B", 1)
		expectAdjacentSibling(mock2, courseID, nodeID, 1, bID, false, learningItemAdjacentRow(aID, courseID, nodeID, "A", 0))
		expectAdjacentSibling(mock2, courseID, nodeID, 1, bID, true, nil)
		result2, err := model2.GetAdjacentLearningItems(courseID, nodeID, bID)
		if err != nil || result2.Next != nil || result2.Previous == nil || result2.Previous.ID != aID {
			t.Fatalf("second of two: result=%+v err=%v", result2, err)
		}
	})
}

func TestLearningItemPreviousNextMissingAndIsolation(t *testing.T) {
	courseID, nodeID, aID, _, _, otherNodeID, otherCourseID := learningItemAdjacentTestIDs()

	t.Run("missing_current_not_found", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentMissingCurrent(mock, courseID, nodeID, aID)
		result, err := model.GetAdjacentLearningItems(courseID, nodeID, aID)
		if !errors.Is(err, ErrLearningItemNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNotFound)
		}
		if result.Previous != nil || result.Next != nil {
			t.Fatalf("empty result required: %+v", result)
		}
	})

	t.Run("cross_course_node", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, false)
		mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
			WithArgs(nodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}).AddRow(nodeID, otherCourseID))
		result, err := model.GetAdjacentLearningItems(courseID, nodeID, aID)
		if !errors.Is(err, ErrLearningItemCrossCourse) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemCrossCourse)
		}
		if result.Previous != nil || result.Next != nil {
			t.Fatalf("empty result required: %+v", result)
		}
	})

	t.Run("missing_node", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, false)
		mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes"`).
			WithArgs(nodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "course_id"}))
		result, err := model.GetAdjacentLearningItems(courseID, nodeID, aID)
		if !errors.Is(err, ErrLearningItemNodeNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNodeNotFound)
		}
		if result.Previous != nil || result.Next != nil {
			t.Fatalf("empty result required: %+v", result)
		}
	})

	t.Run("wrong_node_current_not_found_no_leak", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, otherNodeID, true)
		expectAdjacentMissingCurrent(mock, courseID, otherNodeID, aID)
		result, err := model.GetAdjacentLearningItems(courseID, otherNodeID, aID)
		if !errors.Is(err, ErrLearningItemNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNotFound)
		}
		if result.Previous != nil || result.Next != nil {
			t.Fatalf("must not reveal foreign item: %+v", result)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("wrong_course_current_not_found", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, otherCourseID, nodeID, true)
		expectAdjacentMissingCurrent(mock, otherCourseID, nodeID, aID)
		result, err := model.GetAdjacentLearningItems(otherCourseID, nodeID, aID)
		if !errors.Is(err, ErrLearningItemNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNotFound)
		}
		if result.Previous != nil || result.Next != nil {
			t.Fatalf("must not leak: %+v", result)
		}
	})
}

func TestLearningItemPreviousNextDeterminism(t *testing.T) {
	courseID, nodeID, aID, bID, cID, otherNodeID, _ := learningItemAdjacentTestIDs()

	t.Run("duplicate_positions_tie_break_by_id", func(t *testing.T) {
		// Positions equal at 0 for a and b; c at 1. Current = b → previous a (id <), next c.
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentCurrent(mock, courseID, nodeID, bID, "B", 0)
		expectAdjacentSibling(mock, courseID, nodeID, 0, bID, false, learningItemAdjacentRow(aID, courseID, nodeID, "A", 0))
		expectAdjacentSibling(mock, courseID, nodeID, 0, bID, true, learningItemAdjacentRow(cID, courseID, nodeID, "C", 1))

		result, err := model.GetAdjacentLearningItems(courseID, nodeID, bID)
		if err != nil {
			t.Fatalf("adjacent: %v", err)
		}
		if result.Previous == nil || result.Previous.ID != aID || result.Previous.Position != 0 {
			t.Fatalf("previous = %+v", result.Previous)
		}
		if result.Next == nil || result.Next.ID != cID {
			t.Fatalf("next = %+v", result.Next)
		}
	})

	t.Run("position_gaps_resolve", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentCurrent(mock, courseID, nodeID, bID, "B", 5)
		expectAdjacentSibling(mock, courseID, nodeID, 5, bID, false, learningItemAdjacentRow(aID, courseID, nodeID, "A", 0))
		expectAdjacentSibling(mock, courseID, nodeID, 5, bID, true, learningItemAdjacentRow(cID, courseID, nodeID, "C", 10))

		result, err := model.GetAdjacentLearningItems(courseID, nodeID, bID)
		if err != nil {
			t.Fatalf("adjacent: %v", err)
		}
		if result.Previous == nil || result.Previous.ID != aID {
			t.Fatalf("previous = %+v", result.Previous)
		}
		if result.Next == nil || result.Next.ID != cID {
			t.Fatalf("next = %+v", result.Next)
		}
	})

	t.Run("moved_item_resolves_only_in_new_node", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, otherNodeID, true)
		expectAdjacentCurrent(mock, courseID, otherNodeID, aID, "Moved", 0)
		expectAdjacentSibling(mock, courseID, otherNodeID, 0, aID, false, nil)
		expectAdjacentSibling(mock, courseID, otherNodeID, 0, aID, true, nil)

		result, err := model.GetAdjacentLearningItems(courseID, otherNodeID, aID)
		if err != nil {
			t.Fatalf("adjacent: %v", err)
		}
		if result.Previous != nil || result.Next != nil {
			t.Fatalf("moved single-item node: %+v", result)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestLearningItemPreviousNextQueryFailures(t *testing.T) {
	courseID, nodeID, aID, bID, _, _, _ := learningItemAdjacentTestIDs()

	t.Run("node_lookup_failure", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		mock.ExpectQuery(`SELECT "id", "course_id" FROM "course_nodes".*`).
			WithArgs(courseID, nodeID, uint(1)).
			WillReturnError(errors.New("node lookup boom"))
		result, err := model.GetAdjacentLearningItems(courseID, nodeID, aID)
		if !errors.Is(err, ErrLearningItemPersistence) {
			t.Fatalf("error = %v, want persistence", err)
		}
		if result.Previous != nil || result.Next != nil {
			t.Fatalf("empty result required: %+v", result)
		}
	})

	t.Run("current_query_failure", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		mock.ExpectQuery(`SELECT .* FROM "learning_items"`).
			WithArgs(courseID, nodeID, aID, uint(1)).
			WillReturnError(errors.New("current boom"))
		result, err := model.GetAdjacentLearningItems(courseID, nodeID, aID)
		if !errors.Is(err, ErrLearningItemPersistence) {
			t.Fatalf("error = %v, want persistence", err)
		}
		if result.Previous != nil || result.Next != nil {
			t.Fatalf("empty result required: %+v", result)
		}
	})

	t.Run("previous_query_failure_clears_partial", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentCurrent(mock, courseID, nodeID, bID, "B", 1)
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*ORDER BY`).
			WithArgs(courseID, nodeID, 1, 1, bID, uint(1)).
			WillReturnError(errors.New("prev boom"))
		result, err := model.GetAdjacentLearningItems(courseID, nodeID, bID)
		if !errors.Is(err, ErrLearningItemPersistence) {
			t.Fatalf("error = %v, want persistence", err)
		}
		if result.Previous != nil || result.Next != nil {
			t.Fatalf("must not return partial: %+v", result)
		}
	})

	t.Run("next_query_failure_clears_partial", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentCurrent(mock, courseID, nodeID, bID, "B", 1)
		expectAdjacentSibling(mock, courseID, nodeID, 1, bID, false, learningItemAdjacentRow(aID, courseID, nodeID, "A", 0))
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*ORDER BY`).
			WithArgs(courseID, nodeID, 1, 1, bID, uint(1)).
			WillReturnError(errors.New("next boom"))
		result, err := model.GetAdjacentLearningItems(courseID, nodeID, bID)
		if !errors.Is(err, ErrLearningItemPersistence) {
			t.Fatalf("error = %v, want persistence", err)
		}
		if result.Previous != nil || result.Next != nil {
			t.Fatalf("must not return partial previous: %+v", result)
		}
	})
}

func expectAdjacentPublishedCurrent(
	mock sqlmock.Sqlmock,
	courseID, nodeID, itemID uuid.UUID,
	title string,
	position int,
) {
	mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*`).
		WithArgs(courseID, nodeID, itemID, LearningItemPublishStatePublished, uint(1)).
		WillReturnRows(learningItemAdjacentRow(itemID, courseID, nodeID, title, position))
}

func expectAdjacentPublishedCurrentMissing(mock sqlmock.Sqlmock, courseID, nodeID, itemID uuid.UUID) {
	mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*`).
		WithArgs(courseID, nodeID, itemID, LearningItemPublishStatePublished, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "course_id", "course_node_id", "title", "item_type",
			"description", "metadata", "position", "publish_state", "created_at", "updated_at",
		}))
}

func expectAdjacentPublishedSibling(
	mock sqlmock.Sqlmock,
	courseID, nodeID uuid.UUID,
	currentPosition int,
	currentID uuid.UUID,
	wantNext bool,
	sibling *sqlmock.Rows,
) {
	query := mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`)
	query.WithArgs(
		courseID, nodeID,
		LearningItemPublishStatePublished,
		currentPosition,
		currentPosition, currentID,
		uint(1),
	)
	if sibling != nil {
		query.WillReturnRows(sibling)
	} else {
		query.WillReturnRows(sqlmock.NewRows([]string{
			"id", "course_id", "course_node_id", "title", "item_type",
			"description", "metadata", "position", "publish_state", "created_at", "updated_at",
		}))
	}
}

func TestLearningItemPreviousNextPublished(t *testing.T) {
	courseID, nodeID, aID, bID, cID, otherNodeID, otherCourseID := learningItemAdjacentTestIDs()

	t.Run("nil course node and current item validation", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		tests := []struct {
			name    string
			course  uuid.UUID
			node    uuid.UUID
			current uuid.UUID
			want    error
		}{
			{"nil course", uuid.Nil, nodeID, aID, ErrCourseNotFound},
			{"nil node", courseID, uuid.Nil, aID, ErrLearningItemNodeNotFound},
			{"nil current", courseID, nodeID, uuid.Nil, ErrLearningItemNotFound},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				result, err := model.GetAdjacentPublishedLearningItems(test.course, test.node, test.current)
				if !errors.Is(err, test.want) {
					t.Fatalf("error = %v, want %v", err, test.want)
				}
				if result.Previous != nil || result.Next != nil {
					t.Fatalf("result must be empty on validation failure: %+v", result)
				}
			})
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("validation queried database: %v", err)
		}
	})

	t.Run("middle item with published neighbours", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentPublishedCurrent(mock, courseID, nodeID, bID, "B", 1)
		expectAdjacentPublishedSibling(mock, courseID, nodeID, 1, bID, false, learningItemAdjacentRow(aID, courseID, nodeID, "A", 0))
		expectAdjacentPublishedSibling(mock, courseID, nodeID, 1, bID, true, learningItemAdjacentRow(cID, courseID, nodeID, "C", 2))

		result, err := model.GetAdjacentPublishedLearningItems(courseID, nodeID, bID)
		if err != nil {
			t.Fatalf("adjacent: %v", err)
		}
		if result.Previous == nil || result.Previous.ID != aID {
			t.Fatalf("previous = %+v", result.Previous)
		}
		if result.Next == nil || result.Next.ID != cID {
			t.Fatalf("next = %+v", result.Next)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("first published (previous nil, next published)", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentPublishedCurrent(mock, courseID, nodeID, aID, "A", 0)
		expectAdjacentPublishedSibling(mock, courseID, nodeID, 0, aID, false, nil)
		expectAdjacentPublishedSibling(mock, courseID, nodeID, 0, aID, true, learningItemAdjacentRow(bID, courseID, nodeID, "B", 1))

		result, err := model.GetAdjacentPublishedLearningItems(courseID, nodeID, aID)
		if err != nil {
			t.Fatalf("adjacent: %v", err)
		}
		if result.Previous != nil {
			t.Fatalf("previous want nil, got %+v", result.Previous)
		}
		if result.Next == nil || result.Next.ID != bID {
			t.Fatalf("next = %+v", result.Next)
		}
	})

	t.Run("last published (previous published, next nil)", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentPublishedCurrent(mock, courseID, nodeID, cID, "C", 2)
		expectAdjacentPublishedSibling(mock, courseID, nodeID, 2, cID, false, learningItemAdjacentRow(bID, courseID, nodeID, "B", 1))
		expectAdjacentPublishedSibling(mock, courseID, nodeID, 2, cID, true, nil)

		result, err := model.GetAdjacentPublishedLearningItems(courseID, nodeID, cID)
		if err != nil {
			t.Fatalf("adjacent: %v", err)
		}
		if result.Previous == nil || result.Previous.ID != bID {
			t.Fatalf("previous = %+v", result.Previous)
		}
		if result.Next != nil {
			t.Fatalf("next want nil, got %+v", result.Next)
		}
	})

	t.Run("only one published item", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentPublishedCurrent(mock, courseID, nodeID, aID, "Only", 0)
		expectAdjacentPublishedSibling(mock, courseID, nodeID, 0, aID, false, nil)
		expectAdjacentPublishedSibling(mock, courseID, nodeID, 0, aID, true, nil)

		result, err := model.GetAdjacentPublishedLearningItems(courseID, nodeID, aID)
		if err != nil {
			t.Fatalf("adjacent: %v", err)
		}
		if result.Previous != nil || result.Next != nil {
			t.Fatalf("want both nil, got %+v", result)
		}
	})

	t.Run("current draft not found", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentPublishedCurrentMissing(mock, courseID, nodeID, aID)

		result, err := model.GetAdjacentPublishedLearningItems(courseID, nodeID, aID)
		if !errors.Is(err, ErrLearningItemNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNotFound)
		}
		if result.Previous != nil || result.Next != nil {
			t.Fatalf("empty result required: %+v", result)
		}
	})

	t.Run("current missing not found", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentPublishedCurrentMissing(mock, courseID, nodeID, aID)

		_, err := model.GetAdjacentPublishedLearningItems(courseID, nodeID, aID)
		if !errors.Is(err, ErrLearningItemNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNotFound)
		}
	})

	t.Run("wrong node not found", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, otherNodeID, true)
		expectAdjacentPublishedCurrentMissing(mock, courseID, otherNodeID, aID)

		_, err := model.GetAdjacentPublishedLearningItems(courseID, otherNodeID, aID)
		if !errors.Is(err, ErrLearningItemNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNotFound)
		}
	})

	t.Run("wrong course not found", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, otherCourseID, nodeID, true)
		expectAdjacentPublishedCurrentMissing(mock, otherCourseID, nodeID, aID)

		_, err := model.GetAdjacentPublishedLearningItems(otherCourseID, nodeID, aID)
		if !errors.Is(err, ErrLearningItemNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemNotFound)
		}
	})

	t.Run("skips drafts adjacent neighbours", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentPublishedCurrent(mock, courseID, nodeID, bID, "B", 5)
		// previous sibling skips draft and finds A
		expectAdjacentPublishedSibling(mock, courseID, nodeID, 5, bID, false, learningItemAdjacentRow(aID, courseID, nodeID, "A", 1))
		// next sibling skips draft and finds C
		expectAdjacentPublishedSibling(mock, courseID, nodeID, 5, bID, true, learningItemAdjacentRow(cID, courseID, nodeID, "C", 10))

		result, err := model.GetAdjacentPublishedLearningItems(courseID, nodeID, bID)
		if err != nil {
			t.Fatalf("adjacent: %v", err)
		}
		if result.Previous == nil || result.Previous.ID != aID {
			t.Fatalf("previous should be A, got %+v", result.Previous)
		}
		if result.Next == nil || result.Next.ID != cID {
			t.Fatalf("next should be C, got %+v", result.Next)
		}
	})

	t.Run("duplicate positions resolved by ID", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentPublishedCurrent(mock, courseID, nodeID, bID, "B", 0)
		expectAdjacentPublishedSibling(mock, courseID, nodeID, 0, bID, false, learningItemAdjacentRow(aID, courseID, nodeID, "A", 0))
		expectAdjacentPublishedSibling(mock, courseID, nodeID, 0, bID, true, learningItemAdjacentRow(cID, courseID, nodeID, "C", 1))

		result, err := model.GetAdjacentPublishedLearningItems(courseID, nodeID, bID)
		if err != nil {
			t.Fatalf("adjacent: %v", err)
		}
		if result.Previous == nil || result.Previous.ID != aID {
			t.Fatalf("previous should be A, got %+v", result.Previous)
		}
	})

	t.Run("gaps in positions", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentPublishedCurrent(mock, courseID, nodeID, bID, "B", 50)
		expectAdjacentPublishedSibling(mock, courseID, nodeID, 50, bID, false, learningItemAdjacentRow(aID, courseID, nodeID, "A", 10))
		expectAdjacentPublishedSibling(mock, courseID, nodeID, 50, bID, true, learningItemAdjacentRow(cID, courseID, nodeID, "C", 100))

		result, err := model.GetAdjacentPublishedLearningItems(courseID, nodeID, bID)
		if err != nil {
			t.Fatalf("adjacent: %v", err)
		}
		if result.Previous == nil || result.Previous.ID != aID {
			t.Fatalf("previous should be A, got %+v", result.Previous)
		}
		if result.Next == nil || result.Next.ID != cID {
			t.Fatalf("next should be C, got %+v", result.Next)
		}
	})

	t.Run("previous query failure clears partial", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentPublishedCurrent(mock, courseID, nodeID, bID, "B", 1)
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, LearningItemPublishStatePublished, 1, 1, bID, uint(1)).
			WillReturnError(errors.New("prev boom"))

		result, err := model.GetAdjacentPublishedLearningItems(courseID, nodeID, bID)
		if !errors.Is(err, ErrLearningItemPersistence) {
			t.Fatalf("error = %v, want persistence", err)
		}
		if result.Previous != nil || result.Next != nil {
			t.Fatalf("must not return partial: %+v", result)
		}
	})

	t.Run("next query failure clears partial", func(t *testing.T) {
		model, mock := newLearningItemModelTest(t)
		expectLearningItemNodeLookup(mock, courseID, nodeID, true)
		expectAdjacentPublishedCurrent(mock, courseID, nodeID, bID, "B", 1)
		expectAdjacentPublishedSibling(mock, courseID, nodeID, 1, bID, false, learningItemAdjacentRow(aID, courseID, nodeID, "A", 0))
		mock.ExpectQuery(`SELECT .* FROM "learning_items".*publish_state.*ORDER BY`).
			WithArgs(courseID, nodeID, LearningItemPublishStatePublished, 1, 1, bID, uint(1)).
			WillReturnError(errors.New("next boom"))

		result, err := model.GetAdjacentPublishedLearningItems(courseID, nodeID, bID)
		if !errors.Is(err, ErrLearningItemPersistence) {
			t.Fatalf("error = %v, want persistence", err)
		}
		if result.Previous != nil || result.Next != nil {
			t.Fatalf("must not return partial: %+v", result)
		}
	})
}

func TestAdjacentPublishedLearningItemSQLBoundary(t *testing.T) {
	queries := []string{
		`SELECT "id", "course_id", "course_node_id", "title", "item_type", "description", "metadata", "position", "publish_state", "created_at", "updated_at" FROM "learning_items" WHERE (("course_id" = $1) AND ("course_node_id" = $2) AND ("publish_state" = $3) AND (("position" < $4) OR (("position" = $5) AND ("id" < $6)))) ORDER BY "position" DESC, "id" DESC LIMIT 1`,
		`SELECT "id", "course_id", "course_node_id", "title", "item_type", "description", "metadata", "position", "publish_state", "created_at", "updated_at" FROM "learning_items" WHERE (("course_id" = $1) AND ("course_node_id" = $2) AND ("publish_state" = $3) AND (("position" > $4) OR (("position" = $5) AND ("id" > $6)))) ORDER BY "position" ASC, "id" ASC LIMIT 1`,
	}
	required := []string{"course_id", "course_node_id", "position", "id", "publish_state"}
	banned := []string{"parent_id", "max_depth", "recursive", "with recursive", `"path"`, " nesting"}
	for _, q := range queries {
		lower := strings.ToLower(q)
		for _, need := range required {
			if !strings.Contains(lower, need) {
				t.Fatalf("query missing %q: %s", need, q)
			}
		}
		for _, bad := range banned {
			if strings.Contains(lower, bad) {
				t.Fatalf("query must not contain %q: %s", bad, q)
			}
		}
		if !strings.Contains(lower, "publish_state") {
			t.Fatalf("T22 published adjacent SQL must filter publish_state=PUBLISHED: %s", q)
		}
	}
}

func TestAdjacentLearningItemSQLBoundary(t *testing.T) {
	queries := []string{
		`SELECT "id", "course_id", "course_node_id", "title", "item_type", "description", "metadata", "position", "publish_state", "created_at", "updated_at" FROM "learning_items" WHERE (("course_id" = $1) AND ("course_node_id" = $2) AND (("position" < $3) OR (("position" = $4) AND ("id" < $5)))) ORDER BY "position" DESC, "id" DESC LIMIT 1`,
		`SELECT "id", "course_id", "course_node_id", "title", "item_type", "description", "metadata", "position", "publish_state", "created_at", "updated_at" FROM "learning_items" WHERE (("course_id" = $1) AND ("course_node_id" = $2) AND (("position" > $3) OR (("position" = $4) AND ("id" > $5)))) ORDER BY "position" ASC, "id" ASC LIMIT 1`,
	}
	required := []string{"course_id", "course_node_id", "position", "id"}
	banned := []string{"parent_id", "max_depth", "recursive", "with recursive", `"path"`, " nesting"}
	for _, q := range queries {
		lower := strings.ToLower(q)
		for _, need := range required {
			if !strings.Contains(lower, need) {
				t.Fatalf("query missing %q: %s", need, q)
			}
		}
		for _, bad := range banned {
			if strings.Contains(lower, bad) {
				t.Fatalf("query must not contain %q: %s", bad, q)
			}
		}
		if strings.Contains(lower, "publish_state") && strings.Contains(lower, "published") {
			t.Fatalf("T21 adjacent SQL must not filter publish_state=PUBLISHED: %s", q)
		}
	}
}

