# COURSE-P2-T24 Discovery Leak Analysis

## Learner get draft ≡ missing

Both draft and missing learner gets return:

- HTTP 404
- public `data` message = `constants.ErrLearningItemNotFound`
- no draft-specific wording (“unpublished”, etc.)

Stable public fields (`status`, `data`) match; bodies need not be byte-identical.

## Learner list

- Draft IDs, titles, and `publish_state=DRAFT` absent from JSON
- No placeholder/redacted draft rows
- No draft-inclusive counts or pagination totals (API returns a plain array under `data`)

## Admin visibility

- Authorised admin list/get continue to return drafts
- Non-admin admin-route access remains forbidden
- Learner filtering does not alter admin repository queries (no published-only predicate)

## Ownership

All reads remain scoped by `course_id` + `course_node_id`. No global-by-ID learner lookup.

## Architecture

No new visibility engine, enrollment, unlock, frontend, routes, DTOs, or migrations.
