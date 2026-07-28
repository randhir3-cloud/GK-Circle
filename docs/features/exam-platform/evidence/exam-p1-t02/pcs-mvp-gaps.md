# PCS MVP Study Loop Gaps

Audit date: 2026-07-27

## Target loop (Product Roadmap north star)

```text
Create PCS Course → Subjects/Topics → MCQs → Topic Test → Attempt (scored)
  → Results → Learning Analytics → Daily Revision Queue
```

## Step-by-step gap analysis

| Step | Status at audit | Gap / note | Path forward | Evidence |
|---|---|---|---|---|
| Create PCS course + subjects/topics | COMPLETE in tree | APIs + Course Builder UI present | Maintain via EXAM-P1-T03 | `app/pages/admin/courses/index.vue`, `api/controllers/api/v1/course_node_controller.go` |
| Publish course | COMPLETE in tree | Status transitions via API/UI | EXAM-P1-T03 | `api/models/course.go` (`CourseStatusPublished`), `course_controller.go` |
| Learner opens course outline | COMPLETE in tree | Catalog + outline routes | EXAM-P1-T03 | `learner_course_controller.go`, `app/pages/courses/` |
| Learner enrollment | COMPLETE in tree | Publish gate enforced | EXAM-P1-T03 | `api/models/course_enrollment.go`, `learner_course_enrollment_controller.go` |
| Attach assessments to topics | PARTIAL | `QUIZ_REFERENCE` type without validated quiz link | COURSE-P4 + ADR-024 §6 | `api/models/learning_item.go` |
| Create/import versioned MCQs | PARTIAL | Live quiz CRUD/CSV; revision foundation in tree | EXAM-P2 / EXAM-P3 | `api/models/questions.go`, `question_revision.go` |
| Build topic/mock tests | MISSING | No collections or visual builder | EXAM-P4 | — |
| Start attempt + score (PCS) | MISSING | `assessment_*` schema unused | EXAM-P5 | `20260723100100_create_table_assessment_attempts.up.sql` |
| Student test player | MISSING | No self-paced player UI | EXAM-P6 | — |
| Results + gated review | MISSING | Live review only; unsafe HTTP | EXAM-P7 + EXAM-P2-T03 | `analytics_board_user_controller.go` |
| Learning analytics | MISSING | Live session analytics only | EXAM-P8 | `api/models/analytics_board_user.go` |
| Revision queue | MISSING | Not implemented | EXAM-P9 | — |
| PYQ filters + coverage metrics | MISSING | Not implemented | EXAM-P10 | ADR-024 §9 |

## Security findings (open)

| # | Finding | Route / file | Risk | Planned fix |
|---|---|---|---|---|
| 1 | Unauthenticated analytics with correct answers | `GET /api/v1/analytics_board/user` | Answer-key leak | EXAM-P2-T03 / EXAM-P7 |
| 2 | Unauthenticated played-quiz review | `GET /api/v1/user_played_quizes/:user_played_quiz_id` | Answer-key leak | EXAM-P2-T03 / EXAM-P7 |
| 3 | Unauthenticated final score by session UUID | `GET /api/v1/final_score/user` | Session enumeration / score leak | EXAM-P2-T03 |
| 4 | ~~Enrollment without publish check~~ | ~~`course_enrollment`~~ | **Resolved in tree** | EXAM-P1-T03 |

Evidence for open items:

```410:443:api/routes/main.go
	finalScore := v1.Group("/final_score")
	finalScore.Get("/user", finalScoreBoardController.GetScore)
	// ...
	analyticsBoard := v1.Group("/analytics_board")
	analyticsBoard.Get("/user", analyticsBoardUserController.GetAnalyticsForUser)
	// ...
	userRouter := v1.Group("/user_played_quizes")
	userRouter.Get(fmt.Sprintf("/:%s", constants.UserPlayedQuizId), userPlayedQuizeController.ListUserPlayedQuizesWithQuestionById)
```

## Data-integrity findings (open)

| # | Finding | Evidence | Planned fix |
|---|---|---|---|
| 1 | Self-paced schema without Go runtime | Migrations `20260723100100`, `20260723100200`; no `AssessmentAttempt` model | EXAM-P5 |
| 2 | `attempt_answers` ON DELETE CASCADE | `20260723100200_create_table_attempt_answers.up.sql` L25 | EXAM-P5 per ADR-024 §7 |
| 3 | `QUIZ_REFERENCE` without validated quiz FK | `learning_items.item_type` check only | COURSE-P4 / ADR-024 §6 |
| 4 | Scorer uses live bank answers for live quiz | By design for LIVE; PCS needs snapshot scorer | EXAM-P4 snapshots + EXAM-P5 |

## ADR-024 alignment notes

| ADR decision | Audit result |
|---|---|
| Nuxt retained | ✅ No Next.js app in repo |
| Single engine | ✅ Quiz/question tables reused |
| Collections + snapshots | ❌ Not implemented |
| Answer authority | ⚠️ Foundation in tree; not end-to-end |
| Versioning in P2 | ⚠️ Migration/model started |
| P5 before P6 | ✅ Not violated (neither shipped) |
