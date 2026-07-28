# Capability Status Matrix

Audit date: 2026-07-27

Status legend: **COMPLETE** · **PARTIAL** · **UNSAFE** · **MISSING** · **OUT OF SCOPE**

| Domain | Status | Notes | Key references |
|---|---|---|---|
| Admin Course CRUD + status transitions | COMPLETE | `DRAFT` / `PUBLISHED` / `ARCHIVED` via API | `api/models/course.go`, `api/controllers/api/v1/course_controller.go` |
| CourseNode SUBJECT/TOPIC/SECTION hierarchy | COMPLETE | Tree, move, reorder, delete guards | `api/models/course_node.go`, `api/controllers/api/v1/course_node_controller.go` |
| LearningItem admin CRUD + publish_state | COMPLETE | Course System P2 VERIFIED | `api/models/learning_item.go`, `api/controllers/api/v1/learning_item_controller.go` |
| Learner published LearningItem reads | COMPLETE | Draft filtering, enrollment gate | `api/controllers/api/v1/learner_learning_item_controller.go` |
| Admin LearningItem UI | PARTIAL | Dedicated page; not full Course Builder integration | `app/pages/admin/courses/learning-items.vue` |
| Admin Course Builder UI | COMPLETE | Create course, nodes, publish controls | `app/pages/admin/courses/index.vue` |
| Learner course catalog + outline | COMPLETE | Published courses only | `app/pages/courses/index.vue`, `app/pages/courses/[course_id]/index.vue`, `api/controllers/api/v1/learner_course_controller.go` |
| Enrollment publish gate | COMPLETE | Rejects non-`PUBLISHED` courses | `api/models/course_enrollment.go` (`ErrCourseNotPublished`) |
| Live MCQ authoring (quiz/questions) | COMPLETE | Inherited jovVix flows | `api/models/questions.go`, `app/pages/admin/quiz/` |
| Question versioning + answer authority | PARTIAL | Foundation migration/model/API in tree; full P2 scope not claimed | `api/database/migrations/20260727120000_question_revisions_and_answer_authority.up.sql`, `api/models/question_revision.go` |
| PYQ metadata / bank taxonomy | MISSING | EXAM-P2+ / EXAM-P10 | — |
| Question Collections (STATIC/DYNAMIC) | MISSING | EXAM-P4 | ADR-024 §3 |
| CSV bank import (PCS-scale) | PARTIAL | Per-quiz CSV only | `api/controllers/api/v1/question_controller.go` (`ImportQuestionsByCsv`) |
| Visual Test Builder | MISSING | EXAM-P4 | — |
| Self-paced `assessment_attempts` runtime | MISSING | Schema only; zero Go models | `api/database/migrations/20260723100100_create_table_assessment_attempts.up.sql` |
| PCS mark-based / negative scoring | MISSING | Live game scorer only | `api/utils/calculate_points_score.go` |
| Student self-paced test player | MISSING | EXAM-P6 | — |
| Results + answer review (PCS policy) | MISSING | Live-only analytics | EXAM-P7 |
| Answer-key HTTP protection | UNSAFE | Unauthenticated review endpoints | `api/routes/main.go` (`analytics_board`, `user_played_quizes`, `final_score`) |
| Learning analytics (weak topics, revise-today) | MISSING | EXAM-P8 | — |
| Revision queue / incorrect notebook | MISSING | EXAM-P9 | — |
| Content / learner / mastery coverage | MISSING | EXAM-P10 | ADR-024 §9 |
| `QUIZ_REFERENCE` validated quiz FK | PARTIAL | Type exists; no enforced FK contract | `api/models/learning_item.go`, `api/database/migrations/20260726083000_create_learning_items_table.up.sql` |
| Social / monetisation / native mobile | OUT OF SCOPE | Product roadmap | `PRODUCT_ROADMAP.md` |
| Next.js frontend rewrite | OUT OF SCOPE | Forbidden | `AGENTS.md`, ADR-024 §1 |

## Programme completion estimate (audit snapshot)

Rough evidence-backed progress against the 10 product phases: **~25%** (P1 largely addressable in tree; P2 foundation started; P3–P10 not started).
