# Source Mapping

| Tested contract | Production source | T16 test source |
|---|---|---|
| Course and node selectors | `app/components/course/CourseSelector.vue`, `CourseNodeSelector.vue` | `LearningItemCompositionComponents.test.js` |
| Ordered rows and actions | `LearningItemTable.vue`, `LearningItemRow.vue` | `LearningItemCompositionComponents.test.js` |
| Scalar editor | `LearningItemEditorDialog.vue` | `LearningItemEditorDialog.test.js` |
| Delete confirmation | `LearningItemDeleteDialog.vue` | `LearningItemCompositionComponents.test.js` |
| Admin transport | `app/composables/course_learning_items.js` | `CourseLearningItemsApi.test.js` |
| Admin composition orchestration | Admin learning-items page | `LearningItemsPage.test.js` |
| Learner transport and pages | Learner composable and list/detail pages | Existing composable/page tests |
| Renderer and URL safety | LearningItem renderer, block components, content URL utility | Existing renderer/URL tests |

No production source was added or modified by T16.
