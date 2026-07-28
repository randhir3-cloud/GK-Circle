-- +migrate Up

CREATE TABLE "courses" (
  "id" uuid PRIMARY KEY,
  "owner_id" bpchar(20) NOT NULL REFERENCES "users" ("id") ON DELETE RESTRICT,
  "title" text NOT NULL,
  "short_description" text,
  "language" text,
  "difficulty" text,
  "visibility" text,
  "status" varchar(20) NOT NULL DEFAULT 'DRAFT',
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  CONSTRAINT "courses_title_not_blank" CHECK (btrim("title") <> ''),
  CONSTRAINT "courses_status_check" CHECK ("status" IN ('DRAFT', 'PUBLISHED', 'ARCHIVED'))
);

CREATE INDEX "courses_owner_id_idx" ON "courses" ("owner_id");
