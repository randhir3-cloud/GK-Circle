# Migration Ledger — Before Repair

Date: 2026-07-28  
Captured at: ~2026-07-28T02:23 UTC

## Ledger state (gorp_migrations)

Last applied migration at capture time: `20260727100000_create_course_enrollments_table.up.sql`  
Applied at: `2026-07-27 01:10:35.223226+00`

Total rows in ledger: **89**

### All 89 entries (chronological)

| id | applied_at |
|---|---|
| 000001_create_user_table.down.sql | 2026-07-23 02:44:52 |
| 000001_create_user_table.up.sql | 2026-07-23 02:44:52 |
| 20240213163802_create_username_column.down.sql | 2026-07-23 02:44:52 |
| 20240213163802_create_username_column.up.sql | 2026-07-23 02:44:52 |
| 20240213191111_remove_unique_users_email.down.sql | 2026-07-23 02:44:52 |
| 20240213191111_remove_unique_users_email.up.sql | 2026-07-23 02:44:52 |
| 20240214191648_remove_not_null_users_email.down.sql | 2026-07-23 02:44:52 |
| 20240214191648_remove_not_null_users_email.up.sql | 2026-07-23 02:44:52 |
| 20240215025156_create_table_quiz.down.sql | 2026-07-23 02:44:52 |
| 20240215025156_create_table_quiz.up.sql | 2026-07-23 02:44:52 |
| 20240215030020_create_table_questions.down.sql | 2026-07-23 02:44:52 |
| 20240215030020_create_table_questions.up.sql | 2026-07-23 02:44:52 |
| 20240215031712_create_table_quiz_session.down.sql | 2026-07-23 02:44:52 |
| 20240215031712_create_table_quiz_session.up.sql | 2026-07-23 02:44:52 |
| 20240215032821_create_table_session_user.down.sql | 2026-07-23 02:44:52 |
| 20240215032821_create_table_session_user.up.sql | 2026-07-23 02:44:52 |
| 20240215034326_create_table_session_user_questions.down.sql | 2026-07-23 02:44:52 |
| 20240215034326_create_table_session_user_questions.up.sql | 2026-07-23 02:44:52 |
| 20240216144246_create_column_next_question_quiz_questions.down.sql | 2026-07-23 02:44:52 |
| 20240216144246_create_column_next_question_quiz_questions.up.sql | 2026-07-23 02:44:52 |
| 20240216144343_create_row_current_question_quiz_sessions.down.sql | 2026-07-23 02:44:52 |
| 20240216144343_create_row_current_question_quiz_sessions.up.sql | 2026-07-23 02:44:52 |
| 20240219012502_remove_column_max_attemt.down.sql | 2026-07-23 02:44:52 |
| 20240219012502_remove_column_max_attemt.up.sql | 2026-07-23 02:44:52 |
| 20240219122637_remove_not_null_quiz_sessions_code.down.sql | 2026-07-23 02:44:52 |
| 20240219122637_remove_not_null_quiz_sessions_code.up.sql | 2026-07-23 02:44:52 |
| 20240301103243_rename_column_code_quiz_session.down.sql | 2026-07-23 02:44:52 |
| 20240301103243_rename_column_code_quiz_session.up.sql | 2026-07-23 02:44:52 |
| 20240301103639_rename_table_active_quiz_quiz_session.down.sql | 2026-07-23 02:44:52 |
| 20240301103639_rename_table_active_quiz_quiz_session.up.sql | 2026-07-23 02:44:52 |
| 20240301105318_create_table_session_questions.down.sql | 2026-07-23 02:44:52 |
| 20240301105318_create_table_session_questions.up.sql | 2026-07-23 02:44:52 |
| 20240306210504_rename_table_played_quizzes_user_sessions.down.sql | 2026-07-23 02:44:52 |
| 20240306210504_rename_table_played_quizzes_user_sessions.up.sql | 2026-07-23 02:44:52 |
| 20240306213956_rename_table_played_quiz_questions_user_session_questions.down.sql | 2026-07-23 02:44:52 |
| 20240306213956_rename_table_played_quiz_questions_user_session_questions.up.sql | 2026-07-23 02:44:52 |
| 20240308161219_replace_column_user_played_quiz_id_user_id.down.sql | 2026-07-23 02:44:52 |
| 20240308161219_replace_column_user_played_quiz_id_user_id.up.sql | 2026-07-23 02:44:52 |
| 20240310002318_rename_table_active_quiz_questions_session_questions.down.sql | 2026-07-23 02:44:52 |
| 20240310002318_rename_table_active_quiz_questions_session_questions.up.sql | 2026-07-23 02:44:52 |
| 20240311102624_rename_column_is_attend_is_count.down.sql | 2026-07-23 02:44:52 |
| 20240311102624_rename_column_is_attend_is_count.up.sql | 2026-07-23 02:44:52 |
| 20240311162707_create_column_duration_questions.down.sql | 2026-07-23 02:44:52 |
| 20240311162707_create_column_duration_questions.up.sql | 2026-07-23 02:44:52 |
| 20240311173416_create_column_question_delivery_time_qctive_quizzes.down.sql | 2026-07-23 02:44:52 |
| 20240311173416_create_column_question_delivery_time_qctive_quizzes.up.sql | 2026-07-23 02:44:52 |
| 20240314154416_set_not_null_duration_in_second_questions.down.sql | 2026-07-23 02:44:52 |
| 20240314154416_set_not_null_duration_in_second_questions.up.sql | 2026-07-23 02:44:52 |
| 20240513190411_rename_column_score_to_points.down.sql | 2026-07-23 02:44:52 |
| 20240513190411_rename_column_score_to_points.up.sql | 2026-07-23 02:44:52 |
| 20240513191132_create_column_calculated_points_user_quiz_responses.down.sql | 2026-07-23 02:44:52 |
| 20240513191132_create_column_calculated_points_user_quiz_responses.up.sql | 2026-07-23 02:44:52 |
| 20240710130922_create_column_type_questions.down.sql | 2026-07-23 02:44:52 |
| 20240710130922_create_column_type_questions.up.sql | 2026-07-23 02:44:52 |
| 20240715132722_fill_type_column.down.sql | 2026-07-23 02:44:52 |
| 20240715132722_fill_type_column.up.sql | 2026-07-23 02:44:52 |
| 20240715134225_add_constraints_type_column.down.sql | 2026-07-23 02:44:52 |
| 20240715134225_add_constraints_type_column.up.sql | 2026-07-23 02:44:52 |
| 20240717104743_alter_points_where_one.down.sql | 2026-07-23 02:44:52 |
| 20240717104743_alter_points_where_one.up.sql | 2026-07-23 02:44:52 |
| 20240729125715_create_kratos_schema.down.sql | 2026-07-23 02:44:52 |
| 20240729125715_create_kratos_schema.up.sql | 2026-07-23 02:44:52 |
| 20240916111819_alter_question_add_column.up.sql | 2026-07-23 02:44:52 |
| 20240926151621_alter_users_add_img_key_column.up.sql | 2026-07-23 02:44:52 |
| 20241111151621_create_shared_quizzes_table.down.sql | 2026-07-23 02:44:52 |
| 20241111151621_create_shared_quizzes_table.up.sql | 2026-07-23 02:44:52 |
| 20241114111627_alter_user_quiz_responses_add_streak_count_column.up.sql | 2026-07-23 02:44:52 |
| 20260526120000_alter_quizzes_add_is_public_column.down.sql | 2026-07-23 02:44:52 |
| 20260526120000_alter_quizzes_add_is_public_column.up.sql | 2026-07-23 02:44:52 |
| 20260708100000_create_table_quiz_categories.down.sql | 2026-07-23 02:44:52 |
| 20260708100000_create_table_quiz_categories.up.sql | 2026-07-23 02:44:52 |
| 20260708100100_alter_quizzes_add_category_and_cover_image.down.sql | 2026-07-23 02:44:52 |
| 20260708100100_alter_quizzes_add_category_and_cover_image.up.sql | 2026-07-23 02:44:52 |
| 20260723100000_alter_quizzes_add_self_paced_columns.down.sql | 2026-07-23 10:16:11 |
| 20260723100000_alter_quizzes_add_self_paced_columns.up.sql | 2026-07-23 10:16:11 |
| 20260723100100_create_table_assessment_attempts.down.sql | 2026-07-23 10:16:11 |
| 20260723100100_create_table_assessment_attempts.up.sql | 2026-07-23 10:16:11 |
| 20260723100200_create_table_attempt_answers.down.sql | 2026-07-23 10:16:11 |
| 20260723100200_create_table_attempt_answers.up.sql | 2026-07-23 10:16:11 |
| 20260725110500_create_courses_table.down.sql | 2026-07-27 00:26:17 |
| 20260725110500_create_courses_table.up.sql | 2026-07-27 00:26:17 |
| 20260725134707_create_course_nodes_table.down.sql | 2026-07-27 00:26:17 |
| 20260725134707_create_course_nodes_table.up.sql | 2026-07-27 00:26:18 |
| 20260726083000_create_learning_items_table.down.sql | 2026-07-27 00:26:18 |
| 20260726083000_create_learning_items_table.up.sql | 2026-07-27 00:26:18 |
| 20260726094500_alter_learning_items_add_publish_state.down.sql | 2026-07-27 00:26:18 |
| 20260726094500_alter_learning_items_add_publish_state.up.sql | 2026-07-27 00:26:18 |
| **20260727100000_create_course_enrollments_table.down.sql** | **2026-07-27 01:10:34** |
| **20260727100000_create_course_enrollments_table.up.sql** | **2026-07-27 01:10:35** ← LAST APPLIED |

## Pending migrations (not yet in ledger — these caused errors)

| Migration | Status |
|---|---|
| 20260727120000_question_revisions_and_answer_authority.{down,up}.sql | PENDING |
| 20260727140000_create_question_import_jobs.{down,up}.sql | PENDING |
| 20260727150000_alter_question_import_jobs_commit.{down,up}.sql | PENDING |
| 20260727160000_create_question_collections.{down,up}.sql | PENDING |
| 20260727170000_create_test_snapshots.{down,up}.sql | PENDING |
| 20260728100000_alter_assessment_attempts_add_snapshot.{down,up}.sql | PENDING |
| 20260728100100_alter_attempt_answers_restrict.{down,up}.sql | PENDING |
| 20260728120000_create_assessment_attempt_snapshot_items.{down,up}.sql | PENDING |
