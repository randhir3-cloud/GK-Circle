package structs

import "github.com/google/uuid"

// ReqCreateAdminCourseNode is the admin create-CourseNode request body.
// course_id is taken from the route; unknown body fields are ignored.
type ReqCreateAdminCourseNode struct {
	Title    string          `json:"title"`
	NodeType string          `json:"node_type"`
	Position OptionalInteger `json:"position"`
	ParentID OptionalString  `json:"parent_id"`
}

// ReqMoveAdminCourseNode is the admin move-CourseNode request body.
type ReqMoveAdminCourseNode struct {
	NewParentID OptionalString  `json:"new_parent_id"`
	Position    OptionalInteger `json:"position"`
}

// ReqReorderAdminCourseNodes is the admin sibling-reorder request body.
type ReqReorderAdminCourseNodes struct {
	ParentID       OptionalString `json:"parent_id"`
	OrderedNodeIDs []string       `json:"ordered_node_ids"`
}

// AdminCourseNodeResponse is the stable admin CourseNode representation.
// Path is never exposed.
type AdminCourseNodeResponse struct {
	ID        uuid.UUID  `json:"id"`
	CourseID  uuid.UUID  `json:"course_id"`
	ParentID  *uuid.UUID `json:"parent_id"`
	Title     string     `json:"title"`
	NodeType  string     `json:"node_type"`
	Status    string     `json:"status"`
	Position  int        `json:"position"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}

// AdminCourseNodeHierarchyNode is one nested hierarchy entry.
type AdminCourseNodeHierarchyNode struct {
	Node     AdminCourseNodeResponse        `json:"node"`
	Children []AdminCourseNodeHierarchyNode `json:"children"`
}

// AdminCourseHierarchyResponse is the Course-scoped tree payload.
type AdminCourseHierarchyResponse struct {
	CourseID uuid.UUID                      `json:"course_id"`
	Roots    []AdminCourseNodeHierarchyNode `json:"roots"`
}
