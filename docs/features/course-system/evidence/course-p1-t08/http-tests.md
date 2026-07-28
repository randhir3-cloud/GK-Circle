# COURSE-P1-T08 HTTP Tests

## Route inventory

| Method | Path | AuthN | AuthZ |
|---|---|---|---|
| POST | `/api/v1/admin/courses` | KratosAuthenticated | allowlist |
| GET | `/api/v1/admin/courses` | KratosAuthenticated | allowlist |
| GET | `/api/v1/admin/courses/:course_id` | KratosAuthenticated | allowlist |
| PATCH | `/api/v1/admin/courses/:course_id` | KratosAuthenticated | allowlist |

No DELETE Course or CourseNode routes.

## Request/response DTO inventory

- `ReqCreateAdminCourse` — title + presence-aware optional fields; no `owner_id`
- `ReqUpdateAdminCourse` — presence-aware PATCH fields
- `OptionalString` — omit / null / empty / value
- `AdminCourseResponse` — snake_case public Course fields

## Error mapping matrix

| Condition | Status |
|---|---|
| Missing Kratos session | 401 |
| Missing identity locals | 401 |
| Authenticated non-allowlisted | 403 (Course-admin wording) |
| Malformed JSON / empty patch / invalid title / bad UUID | 400 |
| Course not found | 404 |
| Persistence failure | 500 |

## Coverage

- Unauthenticated 401 via real `KratosAuthenticated`
- Non-admin 403 with Course-admin message
- Admin create/list/get/patch success
- Body `owner_id` ignored; persisted owner is authenticated user
- Empty list returns non-nil JSON array
- Empty PATCH rejected
- Explicit JSON null clears optional field
- CourseNode paths not registered
- Course model + T03–T07 CourseNode regressions green

Admin Course APIs implemented: Yes  
CourseNode APIs implemented: No  
Hierarchy mutation APIs implemented: No
