-- +migrate Up

ALTER TABLE "learning_items"
  ADD COLUMN "quiz_id" uuid;

ALTER TABLE "learning_items"
  ADD CONSTRAINT "learning_items_quiz_fkey"
    FOREIGN KEY ("quiz_id")
    REFERENCES "quizzes" ("id")
    ON DELETE RESTRICT;

ALTER TABLE "learning_items"
  ADD CONSTRAINT "learning_items_quiz_reference_required"
    CHECK (
      ("item_type" = 'QUIZ_REFERENCE' AND "quiz_id" IS NOT NULL)
      OR
      ("item_type" <> 'QUIZ_REFERENCE' AND "quiz_id" IS NULL)
    ) NOT VALID;

CREATE INDEX "learning_items_quiz_id_idx"
  ON "learning_items" ("quiz_id")
  WHERE "quiz_id" IS NOT NULL;
