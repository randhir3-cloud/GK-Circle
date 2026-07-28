-- +migrate Up

-- Supports portfolio quiz queries and per-quiz attempt ordering
CREATE INDEX IF NOT EXISTS idx_attempts_quiz_submitted
  ON assessment_attempts(quiz_id, submitted_at DESC, id DESC);

-- Supports learner-level multi-quiz aggregations
CREATE INDEX IF NOT EXISTS idx_attempts_quiz_user_status
  ON assessment_attempts(quiz_id, user_id, status);

-- Supports question quality bulk queries joining attempt_answers to attempts
CREATE INDEX IF NOT EXISTS idx_attempt_answers_attempt_question
  ON attempt_answers(attempt_id, question_id, is_correct);

-- Supports chronological timeline queries by occurred_at for owned quizzes
CREATE INDEX IF NOT EXISTS idx_analytics_quiz_timeline
  ON assessment_analytics_events(quiz_id, occurred_at DESC, created_at DESC, id DESC);
