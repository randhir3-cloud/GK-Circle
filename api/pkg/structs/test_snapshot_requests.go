package structs

// ReqCreateTestSnapshot optionally selects collection IDs; empty uses all quiz collections.
type ReqCreateTestSnapshot struct {
	CollectionIDs []string `json:"collection_ids"`
}

// TestSnapshotResponse is the editor/admin inspection payload (includes answer keys).
type TestSnapshotResponse struct {
	ID                  string                     `json:"id"`
	QuizID              string                     `json:"quiz_id"`
	CreatedBy           string                     `json:"created_by"`
	Status              string                     `json:"status"`
	SourceCollectionIDs []string                   `json:"source_collection_ids"`
	QuestionCount       int                        `json:"question_count"`
	CreatedAt           string                     `json:"created_at"`
	Items               []TestSnapshotItemResponse `json:"items,omitempty"`
}

type TestSnapshotItemResponse struct {
	ID                  string            `json:"id"`
	Position            int               `json:"position"`
	CollectionID        *string           `json:"collection_id,omitempty"`
	QuestionID          string            `json:"question_id"`
	LineageID           string            `json:"lineage_id"`
	RevisionNumber      int               `json:"revision_number"`
	Question            string            `json:"question"`
	Type                int               `json:"type"`
	Options             map[string]string `json:"options"`
	Answers             []int             `json:"answers"`
	OfficialAnswer      []int             `json:"official_answer"`
	AuthoritativeAnswer []int             `json:"authoritative_answer"`
	AnswerReviewStatus  string            `json:"answer_review_status"`
	Points              *int16            `json:"points,omitempty"`
	DurationInSeconds   *int              `json:"duration_in_seconds,omitempty"`
	QuestionMedia       string            `json:"question_media,omitempty"`
	OptionsMedia        string            `json:"options_media,omitempty"`
	Resource            *string           `json:"resource,omitempty"`
}

// TestSnapshotLearnerResponse omits answer keys for future player consumption.
type TestSnapshotLearnerResponse struct {
	ID                  string                            `json:"id"`
	QuizID              string                            `json:"quiz_id"`
	Status              string                            `json:"status"`
	SourceCollectionIDs []string                          `json:"source_collection_ids"`
	QuestionCount       int                               `json:"question_count"`
	CreatedAt           string                            `json:"created_at"`
	Items               []TestSnapshotLearnerItemResponse `json:"items"`
}

type TestSnapshotLearnerItemResponse struct {
	Position          int               `json:"position"`
	QuestionID        string            `json:"question_id"`
	LineageID         string            `json:"lineage_id"`
	RevisionNumber    int               `json:"revision_number"`
	Question          string            `json:"question"`
	Type              int               `json:"type"`
	Options           map[string]string `json:"options"`
	Points            *int16            `json:"points,omitempty"`
	DurationInSeconds *int              `json:"duration_in_seconds,omitempty"`
	QuestionMedia     string            `json:"question_media,omitempty"`
	OptionsMedia      string            `json:"options_media,omitempty"`
	Resource          *string           `json:"resource,omitempty"`
}
