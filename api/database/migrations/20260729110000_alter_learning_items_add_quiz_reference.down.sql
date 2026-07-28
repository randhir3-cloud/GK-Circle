-- +migrate Down

DROP INDEX IF EXISTS "learning_items_quiz_id_idx";

ALTER TABLE "learning_items"
  DROP CONSTRAINT IF EXISTS "learning_items_quiz_reference_required",
  DROP CONSTRAINT IF EXISTS "learning_items_quiz_fkey",
  DROP COLUMN IF EXISTS "quiz_id";
