# Key Artifacts (audit references)

Audit date: 2026-07-27

## Governance

| Artifact | Path |
|---|---|
| ADR-024 (domain model) | `docs/architecture/ADR/ADR-024-exam-platform-domain-model.md` |
| ADR-023 (course hierarchy) | `docs/architecture/ADR/ADR-023-canonical-course-hierarchy-model.md` |
| Exam Product Roadmap | `docs/development/modules/exam-platform/PRODUCT_ROADMAP.md` |
| Exam Engineering Roadmap | `docs/development/modules/exam-platform/ENGINEERING_ROADMAP.md` |
| P1 phase ledger | `docs/development/modules/exam-platform/phases/exam-p01-course-builder.md` |
| Module decisions | `docs/development/modules/exam-platform/DECISIONS.md` |
| Course P4 assessments phase | `docs/development/modules/course-system/phases/phase-04-assessments.md` |

## Database migrations (exam-relevant)

| Purpose | Path |
|---|---|
| Courses | `api/database/migrations/20260725110500_create_courses_table.up.sql` |
| Course nodes | `api/database/migrations/20260725134707_create_course_nodes_table.up.sql` |
| Learning items | `api/database/migrations/20260726083000_create_learning_items_table.up.sql` |
| Learning item publish state | `api/database/migrations/20260726094500_alter_learning_items_add_publish_state.up.sql` |
| Course enrollments | `api/database/migrations/20260727100000_create_course_enrollments_table.up.sql` |
| Quiz self-paced columns | `api/database/migrations/20260723100000_alter_quizzes_add_self_paced_columns.up.sql` |
| Assessment attempts (schema only) | `api/database/migrations/20260723100100_create_table_assessment_attempts.up.sql` |
| Attempt answers (schema only) | `api/database/migrations/20260723100200_create_table_attempt_answers.up.sql` |
| Question revisions (foundation) | `api/database/migrations/20260727120000_question_revisions_and_answer_authority.up.sql` |
| Legacy questions | `api/database/migrations/20240215030020_create_table_questions.up.sql` |

## Backend — Course System

| Layer | Path |
|---|---|
| Course model | `api/models/course.go` |
| Course node model | `api/models/course_node.go` |
| Learning item model | `api/models/learning_item.go` |
| Enrollment model | `api/models/course_enrollment.go` |
| Course admin controller | `api/controllers/api/v1/course_controller.go` |
| Course node controller | `api/controllers/api/v1/course_node_controller.go` |
| Learning item controller | `api/controllers/api/v1/learning_item_controller.go` |
| Learner course controller | `api/controllers/api/v1/learner_course_controller.go` |
| Learner enrollment controller | `api/controllers/api/v1/learner_course_enrollment_controller.go` |
| Route registration | `api/routes/main.go` |

## Backend — Quiz / assessment (inherited + future)

| Layer | Path |
|---|---|
| Questions model | `api/models/questions.go` |
| Question revision model | `api/models/question_revision.go` |
| Question authority helpers | `api/models/question_authority.go` |
| Quiz model | `api/models/quiz.go` |
| User quiz responses | `api/models/user_quiz_responses.go` |
| Question controller | `api/controllers/api/v1/question_controller.go` |
| Live scoring utility | `api/utils/calculate_points_score.go` |
| Analytics (live) | `api/controllers/api/v1/analytics_board_user_controller.go` |
| Played quiz review | `api/controllers/api/v1/user_played_score_quiz.go` |
| Final scoreboard | `api/controllers/api/v1/final_score_board_controller.go` |

## Frontend

| Surface | Path |
|---|---|
| Admin Course Builder | `app/pages/admin/courses/index.vue` |
| Admin learning items | `app/pages/admin/courses/learning-items.vue` |
| Learner catalog | `app/pages/courses/index.vue` |
| Learner course outline | `app/pages/courses/[course_id]/index.vue` |
| Learner learning items | `app/pages/courses/[course_id]/nodes/[node_id]/learning-items/` |
| Admin quiz management | `app/pages/admin/quiz/` |
| Live play | `app/pages/join/play/[code].vue` |
| Question editor | `app/components/Quiz/EditQuestion.vue` |
| Course API composable | `app/composables/course_learning_items.js` |
| Learner API composable | `app/composables/learner_learning_items.js` |

## Tests (representative)

| Area | Path |
|---|---|
| Course controller | `api/controllers/api/v1/course_controller_test.go` |
| Course enrollment gate | `api/models/course_enrollment_test.go` |
| Learner enrollment | `api/controllers/api/v1/learner_course_enrollment_controller_test.go` |
| Learning item suite | `api/models/learning_item_*_test.go` |
| Question authority | `api/models/question_authority_test.go` |
