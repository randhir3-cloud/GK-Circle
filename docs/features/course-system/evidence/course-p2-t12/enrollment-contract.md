# COURSE-P2-T12 Enrollment Contract

## Authority

- Frozen acceptance: `phases/phase-02-learning-items.md` COURSE-P2-T12
- Decision: `DECISIONS.md` **D-006**

## Persistence

Table `course_enrollments`:

| Column | Type | Notes |
|---|---|---|
| id | uuid PK | |
| course_id | uuid FK → courses(id) ON DELETE CASCADE | |
| user_id | bpchar(20) FK → users(id) ON DELETE CASCADE | |
| enrolled_at | timestamp | default now() |
| UNIQUE(course_id, user_id) | | |

## Learner LearningItem GET gate

Paths:

- `GET /api/v1/learner/courses/:course_id/nodes/:node_id/learning-items`
- `GET /api/v1/learner/courses/:course_id/nodes/:node_id/learning-items/:item_id`

Order:

1. Kratos authentication
2. Course enrollment check (`RequireUserEnrolled`)
3. Repository published read + visibility projection

### Documented denial (unenrolled)

- HTTP **404**
- JSend fail data: `course enrollment required`
- No LearningItem list/detail payload (cannot discover draft or published content)

Draft items for enrolled users remain 404 `learning item not found` (unchanged publish filter).

## Self-enrollment

- `GET /api/v1/learner/courses/:course_id/enrollment` → `{ course_id, enrolled }`
- `POST /api/v1/learner/courses/:course_id/enrollment` → idempotent enroll
- `DELETE /api/v1/learner/courses/:course_id/enrollment` → unenroll (noop if absent)

Missing Course on enroll → 404 `course not found`.

## Frontend

Learner list/detail pages show an accessible **Enroll in course** control when
the API returns enrollment-required denial, then reload content after enroll.
