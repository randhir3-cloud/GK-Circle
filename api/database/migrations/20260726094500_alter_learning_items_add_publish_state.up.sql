-- +migrate Up

ALTER TABLE "learning_items"
  ADD COLUMN "publish_state" text NOT NULL DEFAULT 'DRAFT',
  ADD CONSTRAINT "learning_items_publish_state_check"
    CHECK ("publish_state" IN ('DRAFT', 'PUBLISHED'));
