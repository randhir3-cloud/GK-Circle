# GK Circle Course Rules

Version: 2.0

Status: Mandatory Standard

Owner: Engineering

Last Updated: 2026-07-13

Supersedes:

- The retired educational Course terminology used before ADR-019.

Depends On:

- ADR-019 Course-to-Course Domain Rename.
- ADR-020 Course-Owned Curriculum Builder.

---

# Purpose

Define the active educational Course domain for GK Circle.

Course is the only valid active educational-domain noun. Course terminology is
reserved for a future marketplace domain and must not be used for educational
runtime code, APIs, routes, permissions, telemetry, tests, seeds, fixtures, or
current standards.

---

# Course First Rule

Everything educational is a Course.

All learner-facing curriculum, enrollment, progress, analytics, creator tooling,
test-series attachment, and discovery behavior must resolve through a Course
ownership boundary.

The canonical hierarchy is:

```text
Course
  -> CourseSubject
    -> CourseTopic
      -> TopicContent
```

Global `Subject` and `Topic` records remain reusable taxonomy/library records.
They are not learner curriculum until attached through `CourseSubject` and
`CourseTopic`.

---

# Terminology Invariant

Course is the only active educational-domain aggregate.

The following retired educational-domain terms must not be introduced into
active runtime code, database schema, APIs, permissions, UI, DTOs, services,
controllers, modules, guards, policies, tests, fixtures, seeds, events,
analytics, or current documentation:

- former educational catalog aggregate names retired by ADR-019;
- former educational catalog route families retired by ADR-019;
- former educational catalog permission namespaces retired by ADR-019;
- former educational catalog service, controller, module, and DTO names retired
  by ADR-019.

Historical ADRs, already-applied migration files, migration metadata, archived
evidence, and superseded documents may retain the former terminology only as
immutable historical context. They must not be imported, executed, exposed, or
used as active implementation guidance.

Any unclassified reintroduction of the former educational catalog terminology
fails Course-domain certification.

---

# Required Naming

Use Course terminology everywhere in the active educational domain.

Required examples:

```text
Course
Courses
Course
Courses
courseId
course_id
CoursesModule
CoursesController
CoursesService
CreateCourseDto
UpdateCourseDto
CourseStatus
CourseType
/api/v1/Courses
/Courses
/creator/Courses
/admin/Courses
course:create
course:update
course:publish
course:delete
```

Forbidden active educational-domain examples:

```text
retired educational catalog entity names from ADR-019
retired educational catalog routes from ADR-019
retired educational catalog permission namespaces from ADR-019
```

Historical ADRs and already-applied migration records may mention the old terms
only to document the former architecture.

---

# Supported Course Types

Courses may represent:

- Full exam preparation Courses.
- Subject-specific Courses.
- Test-preparation Courses.
- Practice bundles.
- Cohort-based Courses.
- Free learning Courses.
- Paid Courses, when monetization is enabled.
- Creator-owned Courses.
- Platform-owned Courses.

Marketplace Courses are a future separate domain and must not be mixed into the
Course implementation.

---

# Course Structure

The Course model is the root educational entity.

```text
Course
  -> CourseSubject
    -> TopicContent (subject placement)
      -> TopicContentProgress
    -> CourseTopic
      -> TopicContent (topic placement)
        -> TopicContentProgress
```

Rules:

- Every `CourseSubject` belongs to exactly one Course.
- Every `CourseTopic` belongs to exactly one `CourseSubject`.
- Every `TopicContent` belongs to exactly one `CourseSubject` and may belong to
  one `CourseTopic` under that same `CourseSubject`.
- A `TopicContent` without a `CourseTopic` is subject-level Course Content and
  must never be placed under an artificial Topic.
- Learner-visible curriculum is always Course-scoped.
- Progress authority exists at `TopicContentProgress`.
- Higher-level topic, subject, and Course progress is derived or projected from
  content progress.

Future hierarchy layers such as `CourseSection` require a separate ADR.

---

# Course Ownership Rule

Every Course must have:

- a single authoritative owner or centralized ownership policy;
- explicit creator/admin management authorization;
- immutable audit metadata for state-changing actions;
- enrollment access checks for learner access;
- a complete visibility and publishing lifecycle.

Creators may manage only Courses they own unless a centralized admin policy
allows otherwise.

Authorization must resolve the full parent chain for nested curriculum:

```text
Course
  -> CourseSubject
    -> TopicContent
    -> CourseTopic
      -> TopicContent
```

Never authorize a nested mutation by trusting only the final record ID.

---

# Course Lifecycle

Course lifecycle states must be explicit and enforced server-side.

Minimum lifecycle:

```text
DRAFT
PUBLISHED
ARCHIVED
```

Publishing a Course requires the Course publishing validator to pass.

Minimum Course publishing blockers:

- title missing;
- short description missing;
- language missing;
- difficulty missing;
- visibility missing;
- instructor/creator identity missing;
- required branding missing, if mandated by current policy;
- no published `CourseSubject`;
- no published `CourseTopic` under a published `CourseSubject`;
- no published `TopicContent` under a fully published chain.

Publishing validation must return structured blocker codes so the creator UI can
focus the correct editor section.

---

# Curriculum Status Rule

`CourseSubject`, `CourseTopic`, `TopicContent`, and Course-scoped `Test` and
`Question` records use:

```text
DRAFT
PUBLISHED
ARCHIVED
```

Visibility rules:

- Publishing a child under an unpublished parent does not make it visible to
  students.
- Archiving a parent hides descendants without rewriting child statuses.
- Restoring a parent does not automatically publish descendants.
- Published or consumed curriculum should normally be archived rather than
  hard-deleted.
- Draft curriculum with no learner activity may be hard-deleted when policy
  allows it.

Student visibility requires:

```text
Course is published
AND CourseSubject is active/published
AND (CourseTopic is absent OR CourseTopic is active/published)
AND TopicContent is published
AND student has Course access
```

This must be enforced in backend learner queries, not only hidden in the
frontend.

Course-scoped Tests follow the same optional-Topic visibility chain. A
whole-subject Test may contain subject-level Questions and Questions from active
Topics of the same CourseSubject. A topic-level Test may contain Questions only
from its exact CourseTopic. The backend must enforce these ancestry rules.

---

# Course Builder Rule

Creators build Courses inside the Course Builder workspace.

After Course creation, route creators into the Course context:

```text
/creator/Courses/:courseId
```

Creators should not be returned to the Course list while building curriculum.

The builder must separate:

- Course overview and branding;
- CourseSubject attachment;
- CourseTopic attachment;
- TopicContent editing;
- publishing validation;
- student/analytics/settings views where available.

Builder APIs may expose draft and archived curriculum only to authorized editors.
Learner APIs must use separate Course curriculum contracts.

---

# Course Branding Rule

Course branding is part of Course discovery and recognition, not curriculum
content.

Supported branding fields may include:

- cover asset;
- banner asset;
- logo/icon asset;
- intro video asset;
- theme color;
- tagline;
- preview duration.

Raw storage URLs must not be stored directly on Course records when the Asset
system is available. Course responses should return UI-ready branding objects
without exposing storage keys.

Publishing may require a READY cover image when the approved policy mandates it.

---

# Course Permissions

Use the `Course` permission namespace.

Required examples:

| Permission | Meaning |
|---|---|
| `course:create` | Create a Course |
| `course:update` | Update owned Course fields |
| `course:publish` | Publish or unpublish an owned Course when validation passes |
| `course:delete` | Archive or delete an owned Course according to policy |
| `course:pricing.update` | Change pricing when monetization policy allows it |
| `course:visibility.update` | Change visibility |
| `course:feature` | Feature a Course on discovery surfaces |

Do not introduce `course:*` permissions for educational behavior.

---

# Course Audit Events

State-changing Course actions must create audit records.

Required event examples:

| Action | Event |
|---|---|
| Create Course | `course.create` |
| Update Course | `course.update` |
| Publish Course | `course.publish` |
| Unpublish Course | `course.unpublish` |
| Archive Course | `course.archive` |
| Delete Course | `course.delete` |
| Change pricing | `course.pricing.update` |
| Change visibility | `course.visibility.update` |
| Add subject | `course.subject.create` |
| Reorder subjects | `course.subject.reorder` |
| Add topic | `course.topic.create` |
| Reorder topics | `course.topic.reorder` |
| Add content | `course.content.create` |
| Complete content | `course.content.complete` |
| Grant enrollment | `enrollment.grant` |
| Revoke enrollment | `enrollment.revoke` |

Audit metadata must use IDs, not user PII.

---

# Enrollment Rule

Enrollment is bound to Course access.

Rules:

- A user may have at most one active enrollment per Course.
- Duplicate enrollment attempts must return a clean conflict response.
- Enrollment checks must be enforced server-side for learner-only content.
- Creator/admin management access does not imply learner progress state.
- Historical enrollment/progress data must not be erased by curriculum archival.

---

# Learning Access Rule

Learner APIs must resolve Course access before returning curriculum.

Forbidden:

- standalone learner-facing Subject pages that imply Course enrollment;
- standalone learner-facing Topic pages that bypass Course access;
- frontend-only filtering of builder data;
- returning draft or archived curriculum through learner contracts.

Required learner endpoint pattern:

```text
GET /api/v1/Courses/:courseId/curriculum
```

The response must contain only accessible, published hierarchy and learner
progress for the requesting user.

---

# Course Progress Rule

`TopicContentProgress` is the canonical completion source.

Progress completion must be idempotent:

- repeated completion requests must not double-count XP;
- repeated completion requests must not duplicate streaks;
- repeated completion requests must not duplicate analytics events;
- repeated completion requests must not duplicate outbox events.

If a Course has zero required published `TopicContent`, progress is `0%`.

Higher-level progress may be computed or projected from content completion, but
projection tables are not authorization sources and are not authoritative
completion records.

---

# Course Analytics Rule

Course analytics projections are rebuildable and non-authoritative.

Analytics projections may support:

- views;
- starts;
- completions;
- average time;
- drop-off rate;
- average score;
- last activity.

Projection counters should store totals and counts, not averages of averages.

Analytics projections must not decide:

- authorization;
- enrollment;
- completion truth;
- publishing eligibility.

---

# Course API Rule

Use Course routes for educational APIs.

Required:

```text
GET    /api/v1/Courses
POST   /api/v1/Courses
GET    /api/v1/Courses/:courseId
PATCH  /api/v1/Courses/:courseId
DELETE /api/v1/Courses/:courseId
```

Nested curriculum APIs must stay Course-scoped:

```text
POST   /api/v1/Courses/:courseId/subjects
GET    /api/v1/Courses/:courseId/subjects
PATCH  /api/v1/Courses/:courseId/subjects/:courseSubjectId
DELETE /api/v1/Courses/:courseId/subjects/:courseSubjectId

POST   /api/v1/Courses/:courseId/subjects/:courseSubjectId/topics
GET    /api/v1/Courses/:courseId/subjects/:courseSubjectId/topics
PATCH  /api/v1/Courses/:courseId/subjects/:courseSubjectId/topics/:courseTopicId
DELETE /api/v1/Courses/:courseId/subjects/:courseSubjectId/topics/:courseTopicId

POST   /api/v1/Courses/:courseId/subjects/:courseSubjectId/topics/:courseTopicId/content
GET    /api/v1/Courses/:courseId/subjects/:courseSubjectId/topics/:courseTopicId/content
PATCH  /api/v1/Courses/:courseId/subjects/:courseSubjectId/topics/:courseTopicId/content/:contentId
DELETE /api/v1/Courses/:courseId/subjects/:courseSubjectId/topics/:courseTopicId/content/:contentId
```

Old educational Course routes must return the platform's normal not-found
behavior unless a separately approved ADR defines a temporary compatibility
boundary.

---

# Course Testing Rule

Course changes require focused regression around:

- Course CRUD;
- Course ownership and permissions;
- Course enrollment;
- Course Builder navigation;
- CourseSubject attachment;
- CourseTopic attachment;
- TopicContent validation;
- learner curriculum access;
- progress idempotency;
- old route not-found behavior;
- absence of active Course terminology.

Do not mark Course changes complete without evidence.

---

# Future Marketplace Boundary

Course terminology is reserved for a future marketplace domain.

When that domain is introduced, it must have its own ADR and must not reuse or
relabel the Course implementation.

Marketplace concepts such as order, cart, subscription, invoice, and coupon may
refer to marketplace Courses only after that domain exists.

Until then, active educational guidance must use Course terminology.

---

# Completion Rule

A Course-domain change is complete only when:

- code uses Course terminology;
- Prisma schema and generated client are aligned;
- database objects use Course terminology;
- permissions and telemetry use the `Course` namespace;
- current documentation uses Course terminology;
- historical Course references are classified;
- type-check, build, and affected tests pass;
- evidence is captured.

Build Courses that are correct, complete, secure, and learner-safe.
