-- +migrate Down

ALTER TABLE "learning_items"
  DROP CONSTRAINT IF EXISTS "learning_items_publish_state_check";

ALTER TABLE "learning_items"
  DROP COLUMN IF EXISTS "publish_state";
