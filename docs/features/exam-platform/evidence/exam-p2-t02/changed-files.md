# EXAM-P2-T02 Changed Files

## Added

- `app/utils/question_authority.js`
- `app/composables/quiz_questions.js`
- `app/components/quiz-manage/McqQuestionEditor.vue`
- `app/components/quiz-manage/QuestionRevisionHistory.vue`
- `app/test/utils/question_authority.test.js`
- `app/test/composables/quiz_questions.test.js`
- `app/test/components/quiz-manage/McqQuestionEditor.test.js`
- `docs/features/exam-platform/evidence/exam-p2-t02/`

## Modified

- `app/components/quiz-manage/QuestionFormCard.vue` — wraps `McqQuestionEditor`; links to full editor
- `app/components/Quiz/EditQuestion.vue` — delegates to shared editor + composable
- `app/pages/admin/quiz/list-quiz/[quiz_id]/index.vue` — passes quiz/question IDs; shows review/revision badges
- `app/pages/admin/quiz/list-quiz/[quiz_id]/[question_id].vue` — back link to quiz
- `app/test/components/Quiz/EditQuestion.test.js` — updated for shared editor
- `docs/development/modules/exam-platform/phases/exam-p02-question-bank.md`
- `docs/development/modules/exam-platform/CURRENT_STATUS.md`
