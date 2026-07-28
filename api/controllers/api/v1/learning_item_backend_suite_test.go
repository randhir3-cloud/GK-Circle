package v1

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
)

func assertLearningItemBackendSuiteKeys(t *testing.T, value map[string]any, keys ...string) {
	t.Helper()
	if len(value) != len(keys) {
		t.Fatalf("keys = %v, want exactly %v", value, keys)
	}
	for _, key := range keys {
		if _, ok := value[key]; !ok {
			t.Fatalf("missing key %q in %v", key, value)
		}
	}
}

func assertLearningItemBackendSuiteJSend(
	t *testing.T,
	payload map[string]any,
	status string,
) {
	t.Helper()
	if got, _ := payload["status"].(string); got != status {
		t.Fatalf("status = %q, want %q; payload=%v", got, status, payload)
	}
}

func assertLearningItemBackendSuiteAdminRepresentation(t *testing.T, data map[string]any) {
	t.Helper()
	assertLearningItemBackendSuiteKeys(
		t,
		data,
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
	)
}

func TestLearningItemBackendSuiteCreateDefaultsAndAdminRepresentation(t *testing.T) {
	env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000151")
	now := time.Date(2026, time.July, 27, 1, 2, 3, 456000000, time.UTC)
	metadata := []byte(`{"version":1,"blocks":[]}`)

	env.mock.ExpectBegin()
	expectLearningItemNodeLockHTTP(env.mock, courseID, nodeID, true)
	env.mock.ExpectQuery(`SELECT MAX\("position"\) FROM "learning_items"`).
		WithArgs(nodeID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(int64(4)))
	env.mock.ExpectQuery(`INSERT INTO "learning_items".*RETURNING`).
		WithArgs(
			courseID,
			nodeID,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			models.LearningItemTypeArticle,
			metadata,
			5,
			models.LearningItemPublishStateDraft,
			"Defaulted lesson",
		).
		WillReturnRows(learningItemRow(
			itemID,
			courseID,
			nodeID,
			"Defaulted lesson",
			models.LearningItemTypeArticle,
			nil,
			metadata,
			5,
			now,
		))
	env.mock.ExpectCommit()

	response, payload := doJSON(
		t,
		env.app,
		http.MethodPost,
		learningItemsBase(courseID, nodeID),
		map[string]any{
			"title":     "Defaulted lesson",
			"item_type": string(models.LearningItemTypeArticle),
		},
	)
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d payload=%v", response.StatusCode, payload)
	}
	assertLearningItemBackendSuiteJSend(t, payload, "success")
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("data = %#v", payload["data"])
	}
	assertLearningItemBackendSuiteAdminRepresentation(t, data)
	if data["id"] != itemID.String() ||
		data["course_id"] != courseID.String() ||
		data["course_node_id"] != nodeID.String() ||
		data["description"] != nil ||
		data["position"] != float64(5) ||
		data["publish_state"] != string(models.LearningItemPublishStateDraft) {
		t.Fatalf("defaulted representation = %v", data)
	}
	metadataValue, ok := data["metadata"].(map[string]any)
	if !ok || metadataValue["version"] != float64(1) {
		t.Fatalf("metadata = %#v", data["metadata"])
	}
	blocks, ok := metadataValue["blocks"].([]any)
	if !ok || len(blocks) != 0 {
		t.Fatalf("blocks = %#v", metadataValue["blocks"])
	}
	wantTimestamp := now.Format(time.RFC3339Nano)
	if data["created_at"] != wantTimestamp || data["updated_at"] != wantTimestamp {
		t.Fatalf("timestamps = %v / %v, want %s", data["created_at"], data["updated_at"], wantTimestamp)
	}
	if err := env.mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL: %v", err)
	}
}

func TestLearningItemBackendSuiteListAndGetContracts(t *testing.T) {
	env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	firstID := uuid.MustParse("019c02a0-1111-7000-8000-000000000161")
	secondID := uuid.MustParse("019c02a0-1111-7000-8000-000000000162")
	now := time.Date(2026, time.July, 27, 2, 0, 0, 0, time.UTC)
	metadata := []byte(`{"version":1,"blocks":[]}`)
	base := learningItemsBase(courseID, nodeID)

	expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*ORDER BY "position" ASC, "id" ASC`).
		WithArgs(courseID, nodeID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "course_id", "course_node_id", "title", "item_type",
			"description", "metadata", "position", "publish_state", "created_at", "updated_at",
		}))

	response, payload := doJSON(t, env.app, http.MethodGet, base, nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("empty status = %d payload=%v", response.StatusCode, payload)
	}
	assertLearningItemBackendSuiteJSend(t, payload, "success")
	empty, ok := payload["data"].([]any)
	if !ok || len(empty) != 0 {
		t.Fatalf("empty data = %#v", payload["data"])
	}

	expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)
	rows := learningItemRow(
		firstID, courseID, nodeID, "ID tie-break first", models.LearningItemTypeArticle,
		nil, metadata, 7, now,
	)
	rows.AddRow(
		secondID,
		courseID,
		nodeID,
		"ID tie-break second",
		models.LearningItemTypeVideo,
		"description",
		metadata,
		7,
		models.LearningItemPublishStatePublished,
		now,
		now,
	)
	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*ORDER BY "position" ASC, "id" ASC`).
		WithArgs(courseID, nodeID).
		WillReturnRows(rows)

	response, payload = doJSON(t, env.app, http.MethodGet, base, nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d payload=%v", response.StatusCode, payload)
	}
	assertLearningItemBackendSuiteJSend(t, payload, "success")
	list, ok := payload["data"].([]any)
	if !ok || len(list) != 2 {
		t.Fatalf("list = %#v", payload["data"])
	}
	first, _ := list[0].(map[string]any)
	second, _ := list[1].(map[string]any)
	if first["id"] != firstID.String() || second["id"] != secondID.String() {
		t.Fatalf("order = %v", list)
	}

	env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*LIMIT`).
		WithArgs(courseID, nodeID, secondID, uint(1)).
		WillReturnRows(learningItemRowWithPublishState(
			secondID,
			courseID,
			nodeID,
			"ID tie-break second",
			models.LearningItemTypeVideo,
			"description",
			metadata,
			7,
			now,
			models.LearningItemPublishStatePublished,
		))

	response, payload = doJSON(t, env.app, http.MethodGet, base+"/"+secondID.String(), nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d payload=%v", response.StatusCode, payload)
	}
	assertLearningItemBackendSuiteJSend(t, payload, "success")
	item, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("item = %#v", payload["data"])
	}
	assertLearningItemBackendSuiteAdminRepresentation(t, item)
	if item["id"] != secondID.String() ||
		item["description"] != "description" ||
		item["publish_state"] != string(models.LearningItemPublishStatePublished) {
		t.Fatalf("item = %v", item)
	}
	if err := env.mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL: %v", err)
	}
}

func TestLearningItemBackendSuitePatchPresenceContract(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000171")
	now := time.Date(2026, time.July, 27, 3, 0, 0, 0, time.UTC)
	defaultMetadata := []byte(`{"version":1,"blocks":[]}`)
	metadataInput := map[string]any{
		"version": 1,
		"blocks": []any{
			map[string]any{
				"id":   "text-1",
				"type": "TEXT",
				"data": map[string]any{"text": "Updated"},
			},
		},
	}
	metadataStored := []byte(`{"version":1,"blocks":[{"id":"text-1","type":"TEXT","data":{"text":"Updated"},"visibility":{"mode":"ALL"}}]}`)

	tests := []struct {
		name             string
		body             map[string]any
		sqlArg           any
		returnTitle      string
		returnType       models.LearningItemType
		returnDesc       any
		returnMetadata   []byte
		returnPublish    models.LearningItemPublishState
		wantField        string
		wantResponseData any
	}{
		{
			name: "title only", body: map[string]any{"title": "Renamed"}, sqlArg: "Renamed",
			returnTitle: "Renamed", returnType: models.LearningItemTypeArticle,
			returnMetadata: defaultMetadata, returnPublish: models.LearningItemPublishStateDraft,
			wantField: "title", wantResponseData: "Renamed",
		},
		{
			name: "item type only", body: map[string]any{"item_type": "VIDEO"}, sqlArg: models.LearningItemTypeVideo,
			returnTitle: "Original", returnType: models.LearningItemTypeVideo,
			returnMetadata: defaultMetadata, returnPublish: models.LearningItemPublishStateDraft,
			wantField: "item_type", wantResponseData: "VIDEO",
		},
		{
			name: "description only", body: map[string]any{"description": "Updated description"}, sqlArg: "Updated description",
			returnTitle: "Original", returnType: models.LearningItemTypeArticle, returnDesc: "Updated description",
			returnMetadata: defaultMetadata, returnPublish: models.LearningItemPublishStateDraft,
			wantField: "description", wantResponseData: "Updated description",
		},
		{
			name: "description null clears", body: map[string]any{"description": nil}, sqlArg: nil,
			returnTitle: "Original", returnType: models.LearningItemTypeArticle,
			returnMetadata: defaultMetadata, returnPublish: models.LearningItemPublishStateDraft,
			wantField: "description", wantResponseData: nil,
		},
		{
			name: "metadata replacement", body: map[string]any{"metadata": metadataInput}, sqlArg: metadataStored,
			returnTitle: "Original", returnType: models.LearningItemTypeArticle,
			returnMetadata: metadataStored, returnPublish: models.LearningItemPublishStateDraft,
			wantField: "metadata",
		},
		{
			name: "publish state only", body: map[string]any{"publish_state": "PUBLISHED"}, sqlArg: models.LearningItemPublishStatePublished,
			returnTitle: "Original", returnType: models.LearningItemTypeArticle,
			returnMetadata: defaultMetadata, returnPublish: models.LearningItemPublishStatePublished,
			wantField: "publish_state", wantResponseData: "PUBLISHED",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
			path := learningItemsBase(courseID, nodeID) + "/" + itemID.String()
			env.mock.ExpectQuery(`UPDATE "learning_items".*RETURNING`).
				WithArgs(test.sqlArg, courseID, nodeID, itemID).
				WillReturnRows(learningItemRowWithPublishState(
					itemID,
					courseID,
					nodeID,
					test.returnTitle,
					test.returnType,
					test.returnDesc,
					test.returnMetadata,
					9,
					now,
					test.returnPublish,
				))

			response, payload := doJSON(t, env.app, http.MethodPatch, path, test.body)
			if response.StatusCode != http.StatusOK {
				t.Fatalf("status = %d payload=%v", response.StatusCode, payload)
			}
			assertLearningItemBackendSuiteJSend(t, payload, "success")
			data, ok := payload["data"].(map[string]any)
			if !ok {
				t.Fatalf("data = %#v", payload["data"])
			}
			assertLearningItemBackendSuiteAdminRepresentation(t, data)
			if test.wantField == "metadata" {
				value, ok := data["metadata"].(map[string]any)
				if !ok || value["version"] != float64(1) {
					t.Fatalf("metadata = %#v", data["metadata"])
				}
			} else if data[test.wantField] != test.wantResponseData {
				t.Fatalf("%s = %#v, want %#v", test.wantField, data[test.wantField], test.wantResponseData)
			}
			if test.name != "title only" && data["title"] != "Original" {
				t.Fatalf("omitted title changed: %v", data)
			}
			if err := env.mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet SQL: %v", err)
			}
		})
	}

	rejections := []struct {
		name string
		body map[string]any
		data string
	}{
		{name: "title null", body: map[string]any{"title": nil}, data: constants.ErrLearningItemTitleInvalid},
		{name: "item type null", body: map[string]any{"item_type": nil}, data: constants.ErrLearningItemTypeInvalid},
		{name: "metadata null", body: map[string]any{"metadata": nil}, data: constants.ErrLearningItemMetadataInvalid},
		{name: "publish state null", body: map[string]any{"publish_state": nil}, data: constants.ErrLearningItemPublishState},
		{name: "empty patch", body: map[string]any{}, data: constants.ErrLearningItemEmptyPatch},
	}
	for _, test := range rejections {
		t.Run(test.name, func(t *testing.T) {
			env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
			path := learningItemsBase(courseID, nodeID) + "/" + itemID.String()
			response, payload := doJSON(t, env.app, http.MethodPatch, path, test.body)
			if response.StatusCode != http.StatusBadRequest {
				t.Fatalf("status = %d payload=%v", response.StatusCode, payload)
			}
			assertLearningItemBackendSuiteJSend(t, payload, "fail")
			if payload["data"] != test.data {
				t.Fatalf("data = %#v, want %q", payload["data"], test.data)
			}
			if err := env.mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("rejected PATCH queried database: %v", err)
			}
		})
	}
}

func TestLearningItemBackendSuiteDeleteAndJSendErrors(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	itemID := uuid.MustParse("019c02a0-1111-7000-8000-000000000181")
	path := learningItemsBase(courseID, nodeID) + "/" + itemID.String()

	t.Run("delete success and missing", func(t *testing.T) {
		env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
		env.mock.ExpectExec(`DELETE FROM "learning_items"`).
			WithArgs(courseID, nodeID, itemID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		response, payload := doJSON(t, env.app, http.MethodDelete, path, nil)
		if response.StatusCode != http.StatusOK {
			t.Fatalf("status = %d payload=%v", response.StatusCode, payload)
		}
		assertLearningItemBackendSuiteJSend(t, payload, "success")
		if payload["data"] != "success" {
			t.Fatalf("data = %#v", payload["data"])
		}

		env.mock.ExpectExec(`DELETE FROM "learning_items"`).
			WithArgs(courseID, nodeID, itemID).
			WillReturnResult(sqlmock.NewResult(0, 0))
		response, payload = doJSON(t, env.app, http.MethodDelete, path, nil)
		if response.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d payload=%v", response.StatusCode, payload)
		}
		assertLearningItemBackendSuiteJSend(t, payload, "fail")
		if payload["data"] != constants.ErrLearningItemNotFound {
			t.Fatalf("data = %#v", payload["data"])
		}
	})

	t.Run("unauthenticated error envelope", func(t *testing.T) {
		env := newLearningItemHTTPEnv(t, kratosAuthMiddleware(t))
		response, payload := doJSON(t, env.app, http.MethodGet, learningItemsBase(courseID, nodeID), nil)
		if response.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d payload=%v", response.StatusCode, payload)
		}
		assertLearningItemBackendSuiteJSend(t, payload, "error")
		if payload["message"] != constants.ErrKratosIDEmpty || payload["code"] != float64(http.StatusUnauthorized) {
			t.Fatalf("payload = %v", payload)
		}
	})

	t.Run("forbidden fail envelope", func(t *testing.T) {
		env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
		response, payload := doJSON(t, env.app, http.MethodGet, learningItemsBase(courseID, nodeID), nil)
		if response.StatusCode != http.StatusForbidden {
			t.Fatalf("status = %d payload=%v", response.StatusCode, payload)
		}
		assertLearningItemBackendSuiteJSend(t, payload, "fail")
		if payload["data"] != constants.ErrCourseAdminForbidden {
			t.Fatalf("payload = %v", payload)
		}
	})

	t.Run("create conflict fail envelope", func(t *testing.T) {
		env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
		metadata := []byte(`{"version":1,"blocks":[]}`)
		env.mock.ExpectBegin()
		expectLearningItemNodeLockHTTP(env.mock, courseID, nodeID, true)
		env.mock.ExpectQuery(`SELECT MAX\("position"\) FROM "learning_items"`).
			WithArgs(nodeID, uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(int64(0)))
		env.mock.ExpectQuery(`INSERT INTO "learning_items".*RETURNING`).
			WithArgs(
				courseID,
				nodeID,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				models.LearningItemTypeArticle,
				metadata,
				1,
				models.LearningItemPublishStateDraft,
				"Conflict",
			).
			WillReturnError(&pq.Error{
				Code:       "23505",
				Constraint: "learning_items_node_position_unique",
			})
		env.mock.ExpectRollback()

		response, payload := doJSON(
			t,
			env.app,
			http.MethodPost,
			learningItemsBase(courseID, nodeID),
			map[string]any{"title": "Conflict", "item_type": "ARTICLE"},
		)
		if response.StatusCode != http.StatusConflict {
			t.Fatalf("status = %d payload=%v", response.StatusCode, payload)
		}
		assertLearningItemBackendSuiteJSend(t, payload, "fail")
		if payload["data"] != constants.ErrLearningItemConflict {
			t.Fatalf("payload = %v", payload)
		}
	})

	t.Run("repository failure error envelope", func(t *testing.T) {
		env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
		expectLearningItemNodeLookupHTTP(env.mock, courseID, nodeID, true)
		env.mock.ExpectQuery(`SELECT .* FROM "learning_items".*ORDER BY`).
			WithArgs(courseID, nodeID).
			WillReturnError(errors.New("database unavailable"))

		response, payload := doJSON(t, env.app, http.MethodGet, learningItemsBase(courseID, nodeID), nil)
		if response.StatusCode != http.StatusInternalServerError {
			t.Fatalf("status = %d payload=%v", response.StatusCode, payload)
		}
		assertLearningItemBackendSuiteJSend(t, payload, "error")
		if payload["message"] != constants.ErrListLearningItems ||
			payload["code"] != float64(http.StatusInternalServerError) {
			t.Fatalf("payload = %v", payload)
		}
	})
}

func TestLearningItemBackendSuiteCreateRequiredFields(t *testing.T) {
	tests := []struct {
		name string
		body map[string]any
		data string
	}{
		{
			name: "title required",
			body: map[string]any{"item_type": "ARTICLE"},
			data: constants.ErrLearningItemTitleInvalid,
		},
		{
			name: "item type required",
			body: map[string]any{"title": "Missing type"},
			data: constants.ErrLearningItemTypeInvalid,
		},
		{
			name: "null publish state rejected",
			body: map[string]any{"title": "Lesson", "item_type": "ARTICLE", "publish_state": nil},
			data: constants.ErrLearningItemPublishState,
		},
		{
			name: "invalid publish state rejected",
			body: map[string]any{"title": "Lesson", "item_type": "ARTICLE", "publish_state": "published"},
			data: constants.ErrLearningItemPublishState,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := newLearningItemHTTPEnv(t, injectAuthenticatedUser(adminUser()))
			response, payload := doJSON(
				t,
				env.app,
				http.MethodPost,
				learningItemsBase(uuid.New(), uuid.New()),
				test.body,
			)
			if response.StatusCode != http.StatusBadRequest {
				t.Fatalf("status = %d payload=%v", response.StatusCode, payload)
			}
			assertLearningItemBackendSuiteJSend(t, payload, "fail")
			if payload["data"] != test.data {
				t.Fatalf("data = %#v, want %q", payload["data"], test.data)
			}
			if err := env.mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("invalid create queried database: %v", err)
			}
		})
	}
}
