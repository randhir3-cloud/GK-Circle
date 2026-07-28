-- +migrate Up

CREATE TABLE "learning_items" (
  "id" uuid PRIMARY KEY,
  "course_id" uuid NOT NULL,
  "course_node_id" uuid NOT NULL,
  "title" text NOT NULL,
  "item_type" varchar(32) NOT NULL,
  "description" text,
  "metadata" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "position" integer NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  CONSTRAINT "learning_items_course_node_fkey"
    FOREIGN KEY ("course_id", "course_node_id")
    REFERENCES "course_nodes" ("course_id", "id") ON DELETE RESTRICT,
  CONSTRAINT "learning_items_title_not_blank"
    CHECK (btrim("title") <> ''),
  CONSTRAINT "learning_items_item_type_check"
    CHECK ("item_type" IN ('ARTICLE', 'VIDEO', 'PDF', 'LINK', 'QUIZ_REFERENCE')),
  CONSTRAINT "learning_items_position_nonnegative"
    CHECK ("position" >= 0),
  CONSTRAINT "learning_items_node_position_unique"
    UNIQUE ("course_node_id", "position")
);

CREATE INDEX "learning_items_node_position_idx"
  ON "learning_items" ("course_node_id", "position");
