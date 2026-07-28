package structs

import "encoding/json"

// ReqCreateAdminLearningItem is the admin create-LearningItem request body.
// course_id and course_node_id come from the path; owner is not client-supplied.
type ReqCreateAdminLearningItem struct {
	Title        string         `json:"title"`
	ItemType     string         `json:"item_type"`
	Description  OptionalString `json:"description"`
	Metadata     OptionalJSON   `json:"metadata"`
	QuizID       OptionalString `json:"quiz_id"`
	PublishState OptionalString `json:"publish_state"`
}

// ReqUpdateAdminLearningItem is the admin partial-update LearningItem request body.
// Identifiers come from the path; only mutable fields are accepted.
type ReqUpdateAdminLearningItem struct {
	Title        OptionalString `json:"title"`
	ItemType     OptionalString `json:"item_type"`
	Description  OptionalString `json:"description"`
	Metadata     OptionalJSON   `json:"metadata"`
	QuizID       OptionalString `json:"quiz_id"`
	PublishState OptionalString `json:"publish_state"`
}

// ReqReorderAdminLearningItems is the admin sibling-reorder request body.
// course_id and course_node_id come from the path.
type ReqReorderAdminLearningItems struct {
	OrderedItemIDs []string `json:"ordered_item_ids"`
}

// ResReorderAdminLearningItems is the frozen admin reorder response payload.
type ResReorderAdminLearningItems struct {
	CourseNodeID      string `json:"course_node_id"`
	LearningItemCount int    `json:"learning_item_count"`
	PositionsUpdated  int    `json:"positions_updated"`
	Noop              bool   `json:"noop"`
}

// ReqMoveAdminLearningItems is the admin cross-node move request body.
// course_id and source node_id come from the path.
type ReqMoveAdminLearningItems struct {
	DestinationNodeID string   `json:"destination_node_id"`
	OrderedItemIDs    []string `json:"ordered_item_ids"`
}

// ResMoveLearningItems is the frozen admin move response payload.
type ResMoveLearningItems struct {
	SourceNodeID         string `json:"source_node_id"`
	DestinationNodeID    string `json:"destination_node_id"`
	ItemsMoved           int    `json:"items_moved"`
	SourceItemCount      int    `json:"source_item_count"`
	DestinationItemCount int    `json:"destination_item_count"`
	Noop                 bool   `json:"noop"`
}

// AdminLearningItemResponse is the stable admin LearningItem representation.
type AdminLearningItemResponse struct {
	ID           string          `json:"id"`
	CourseID     string          `json:"course_id"`
	CourseNodeID string          `json:"course_node_id"`
	Title        string          `json:"title"`
	ItemType     string          `json:"item_type"`
	Description  *string         `json:"description"`
	Metadata     json.RawMessage `json:"metadata"`
	Position     int             `json:"position"`
	PublishState string          `json:"publish_state"`
	CreatedAt    string          `json:"created_at"`
	UpdatedAt    string          `json:"updated_at"`
}

// LearnerLearningItemResponse is the learner-facing LearningItem representation.
// Admin-only ordering/timing fields are intentionally omitted.
type LearnerLearningItemResponse struct {
	ID           string          `json:"id"`
	Title        string          `json:"title"`
	ItemType     string          `json:"item_type"`
	Description  *string         `json:"description"`
	Metadata     json.RawMessage `json:"metadata"`
	PublishState string          `json:"publish_state"`
}

// LearnerLearningItemNavigation represents the minimal DTO for learner navigation.
type LearnerLearningItemNavigation struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// LearnerLearningItemDetailResponse wraps the item and its navigation links.
type LearnerLearningItemDetailResponse struct {
	LearningItem LearnerLearningItemResponse    `json:"learning_item"`
	Previous     *LearnerLearningItemNavigation `json:"previous"`
	Next         *LearnerLearningItemNavigation `json:"next"`
}

// LearnerCourseEnrollmentResponse is the learner Course enrollment status payload.
type LearnerCourseEnrollmentResponse struct {
	CourseID   string  `json:"course_id"`
	Enrolled   bool    `json:"enrolled"`
	EnrolledAt *string `json:"enrolled_at,omitempty"`
}
