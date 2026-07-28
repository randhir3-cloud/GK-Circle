# COURSE-P2-T14 Closure Audit

Audited: 2026-07-27

| Frozen acceptance item | Implemented | Automated verification | Runtime evidence | Result |
|---|---|---|---|---|
| Ordered published learner items at any CourseNode depth | Yes | Yes | Depth-4 PostgreSQL/API/UI correlation and desktop/mobile screenshots | PASS |
| Draft exclusion, empty state, and unauthorized error contract | Yes | Yes | Published/draft sibling correlation, successful empty node, enrollment-required and signed-out states | PASS |
| Evidence bundle | Yes | Hashes verified | Runtime records and seven screenshots present | PASS |

## Source audit

The learner list/detail routes, transport composable, generation-ID stale
protection, API-order rendering, data-driven renderer, supported/fallback block
dispatch, URL safety, metadata immutability, escaped text, accessibility
attributes, and API-provided adjacency are present. T12 enrollment handling is
covered on both pages. The closure run found and fixed one page-root
shrink-to-fit regression by adding `w-full flex-1` to both learner routes, with
page-test coverage.

## Runtime audit

- Signed in through the normal local Kratos flow with the approved QA identity.
- An unenrolled deep-node request returned `course enrollment required`; the UI
  enrollment action persisted and reloaded the learner list successfully.
- PostgreSQL ancestry proved the selected CourseNode was depth 4.
- Persisted position 0 `PUBLISHED` rendered before a temporary position 2
  `PUBLISHED` item; the position 1 `DRAFT` sibling never appeared.
- Representative HEADING, TEXT, LINK, CALLOUT, hidden TEXT, and DIVIDER blocks
  were written through the existing authenticated admin API. The learner API/UI
  preserved visible block order and omitted the hidden block.
- External HTTPS links retained `_blank` plus `noopener noreferrer`;
  root-relative links remained same-tab.
- Previous/next links contained backend-provided IDs only. First/last boundaries
  omitted the absent control.
- Refresh preserved rendered data. Successful empty and signed-out denial states
  were observed. Browser warning/error logs were empty.
- Screenshots were captured at 1280x900 and 360x800.
- The temporary adjacent item was deleted and the retained item was restored to
  null description plus `{"version":1,"blocks":[]}`; residue count was zero.

## Verification decision

Focused tests, lint, build, backend tests, Compose checks, final image build,
local service health, authenticated runtime, screenshots, evidence integrity,
status consistency, and scoped diff checks pass. The full-suite failures
reproduce the unchanged unrelated inherited baseline.

Status: `VERIFIED`

Database Migration: NONE

Breaking Changes: NO

Production Behaviour Changes: responsive learner-page root sizing only
