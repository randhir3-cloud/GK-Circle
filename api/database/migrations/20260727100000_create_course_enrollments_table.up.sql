-- +migrate Up

CREATE TABLE "course_enrollments" (
  "id" uuid PRIMARY KEY,
  "course_id" uuid NOT NULL REFERENCES "courses" ("id") ON DELETE CASCADE,
  "user_id" bpchar(20) NOT NULL REFERENCES "users" ("id") ON DELETE CASCADE,
  "enrolled_at" timestamp NOT NULL DEFAULT (now()),
  CONSTRAINT "course_enrollments_course_user_unique"
    UNIQUE ("course_id", "user_id")
);

CREATE INDEX "course_enrollments_user_id_idx"
  ON "course_enrollments" ("user_id");

CREATE INDEX "course_enrollments_course_id_idx"
  ON "course_enrollments" ("course_id");
