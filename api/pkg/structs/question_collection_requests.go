package structs

// CollectionDynamicFilter stores DYNAMIC collection criteria per ADR-024 §3.
type CollectionDynamicFilter struct {
	Subject    *string `json:"subject,omitempty"`
	Topic      *string `json:"topic,omitempty"`
	Year       *int    `json:"year,omitempty"`
	Difficulty *string `json:"difficulty,omitempty"`
	PYQStatus  *bool   `json:"pyq_status,omitempty"`
}

// ReqCreateQuestionCollection is the create-collection request body.
type ReqCreateQuestionCollection struct {
	Title    string                   `json:"title"`
	Kind     string                   `json:"kind"`
	Position int                      `json:"position"`
	Filter   *CollectionDynamicFilter `json:"filter,omitempty"`
}

// ReqUpdateQuestionCollection is the partial-update collection request body.
type ReqUpdateQuestionCollection struct {
	Title    *string                  `json:"title,omitempty"`
	Position *int                     `json:"position,omitempty"`
	Filter   *CollectionDynamicFilter `json:"filter,omitempty"`
}

// ReqReplaceQuestionCollectionMembers replaces STATIC collection membership.
type ReqReplaceQuestionCollectionMembers struct {
	QuestionIDs []string `json:"question_ids"`
}

// QuestionCollectionResponse is the stable collection API representation.
type QuestionCollectionResponse struct {
	ID        string                              `json:"id"`
	QuizID    string                              `json:"quiz_id"`
	Title     string                              `json:"title"`
	Kind      string                              `json:"kind"`
	Position  int                                 `json:"position"`
	Filter    *CollectionDynamicFilter            `json:"filter,omitempty"`
	CreatedBy string                              `json:"created_by"`
	CreatedAt string                              `json:"created_at"`
	UpdatedAt string                              `json:"updated_at"`
	Members   []QuestionCollectionMemberResponse  `json:"members,omitempty"`
}

type QuestionCollectionMemberResponse struct {
	ID         string `json:"id"`
	QuestionID string `json:"question_id"`
	Position   int    `json:"position"`
}

// CollectionResolutionResponse is the resolve preview for STATIC/DYNAMIC collections.
type CollectionResolutionResponse struct {
	CollectionID     string   `json:"collection_id"`
	Kind             string   `json:"kind"`
	QuestionIDs      []string `json:"question_ids"`
	ResolutionStatus string   `json:"resolution_status"`
	Message          string   `json:"message,omitempty"`
}
