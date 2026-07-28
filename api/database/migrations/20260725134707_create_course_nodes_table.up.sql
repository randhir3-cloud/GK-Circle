-- +migrate Up

CREATE TABLE "course_nodes" (
  "id" uuid PRIMARY KEY,
  "course_id" uuid NOT NULL,
  "parent_id" uuid,
  "node_type" varchar(20) NOT NULL,
  "title" text NOT NULL,
  "position" integer NOT NULL,
  "path" text NOT NULL,
  "status" varchar(20) NOT NULL DEFAULT 'DRAFT',
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  CONSTRAINT "course_nodes_course_id_fkey"
    FOREIGN KEY ("course_id") REFERENCES "courses" ("id") ON DELETE RESTRICT,
  CONSTRAINT "course_nodes_course_id_id_unique"
    UNIQUE ("course_id", "id"),
  CONSTRAINT "course_nodes_course_parent_fkey"
    FOREIGN KEY ("course_id", "parent_id")
    REFERENCES "course_nodes" ("course_id", "id") ON DELETE RESTRICT,
  CONSTRAINT "course_nodes_not_self_parent"
    CHECK ("parent_id" IS NULL OR "parent_id" <> "id"),
  CONSTRAINT "course_nodes_node_type_check"
    CHECK ("node_type" IN ('SECTION', 'SUBJECT', 'TOPIC')),
  CONSTRAINT "course_nodes_title_not_blank"
    CHECK (btrim("title") <> ''),
  CONSTRAINT "course_nodes_position_nonnegative"
    CHECK ("position" >= 0),
  CONSTRAINT "course_nodes_status_check"
    CHECK ("status" IN ('DRAFT', 'PUBLISHED', 'ARCHIVED')),
  CONSTRAINT "course_nodes_course_path_unique"
    UNIQUE ("course_id", "path")
);

CREATE UNIQUE INDEX "course_nodes_top_level_position_unique"
  ON "course_nodes" ("course_id", "position")
  WHERE "parent_id" IS NULL;

CREATE UNIQUE INDEX "course_nodes_child_position_unique"
  ON "course_nodes" ("course_id", "parent_id", "position")
  WHERE "parent_id" IS NOT NULL;

CREATE INDEX "course_nodes_course_path_prefix_idx"
  ON "course_nodes" ("course_id", "path" text_pattern_ops);
