# COURSE-P1-T09 Integration Verification

Environment: Fiber `app.Test` with sqlmock CourseNodeModel and route middleware
chain (real `KratosAuthenticated` for 401; injected authenticated identities for
403/admin success).

| Case | Result |
|---|---|
| Unauthenticated list roots | PASSED (401) |
| Non-admin list | PASSED (403) |
| Admin create root → list roots → get by id | PASSED |
| Required position / parent_id rules | PASSED |
| Body `course_id` / `owner_id` ignored | PASSED |
| Cross-Course get / children | PASSED (404) |
| Missing Course list | PASSED (404) |
| Empty tree for existing Course | PASSED (`roots: []`) |
| Mutation routes absent | PASSED (404 or 405) |
| T03–T08 Course/CourseNode regression | PASSED |

Docker Compose `api-verify` Go 1.23 vet/test/race/build: PASSED  
`docker compose config --quiet`: PASSED  
Frontend: NOT_APPLICABLE  
Database Migration: NONE  
Production / NUC: NOT_TOUCHED
