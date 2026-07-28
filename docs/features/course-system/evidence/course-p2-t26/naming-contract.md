# COURSE-P2-T26 Naming Contract

All API JSON examples use the implementation's snake_case field names,
including:

- `course_id`
- `course_node_id`
- `item_type`
- `publish_state`
- `created_at`
- `updated_at`
- `learning_item`

## State distinction

`CourseNode.status` represents CourseNode lifecycle state.

`LearningItem.publish_state` represents LearningItem publication and accepts
`DRAFT` or `PUBLISHED`.

The canonical document does not rename `publish_state` to `status` or claim that
the two fields are interchangeable.

