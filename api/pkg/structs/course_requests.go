package structs

// ReqCreateAdminCourse is the admin create-Course request body.
// owner_id is intentionally absent; ownership comes from authenticated context.
type ReqCreateAdminCourse struct {
	Title            string         `json:"title"`
	ShortDescription OptionalString `json:"short_description"`
	Language         OptionalString `json:"language"`
	Difficulty       OptionalString `json:"difficulty"`
	Visibility       OptionalString `json:"visibility"`
}

// ReqUpdateAdminCourse is the admin partial-update Course request body.
type ReqUpdateAdminCourse struct {
	Title            OptionalString `json:"title"`
	ShortDescription OptionalString `json:"short_description"`
	Language         OptionalString `json:"language"`
	Difficulty       OptionalString `json:"difficulty"`
	Visibility       OptionalString `json:"visibility"`
	Status           OptionalString `json:"status"`
}

// AdminCourseResponse is the stable admin Course representation.
type AdminCourseResponse struct {
	ID               string  `json:"id"`
	OwnerID          string  `json:"owner_id"`
	Title            string  `json:"title"`
	ShortDescription *string `json:"short_description,omitempty"`
	Language         *string `json:"language,omitempty"`
	Difficulty       *string `json:"difficulty,omitempty"`
	Visibility       *string `json:"visibility,omitempty"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// LearnerCourseResponse is the learner-facing published Course summary.
type LearnerCourseResponse struct {
	ID               string  `json:"id"`
	Title            string  `json:"title"`
	ShortDescription *string `json:"short_description,omitempty"`
	Language         *string `json:"language,omitempty"`
	Difficulty       *string `json:"difficulty,omitempty"`
	Status           string  `json:"status"`
}
