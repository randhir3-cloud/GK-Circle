# Frontend Coverage Matrix

| Area | Coverage |
|---|---|
| Course selector | Loading, error alert, exact option order, explicit emit |
| CourseNode selector | Multiple direct-child levels, API order, responsive grid, error, level/ID emit |
| LearningItem table/row | Exact item order, read-only position, fields, accessible table, edit/delete forwarding |
| Editor dialog | Required title, frozen types/states, defaults, description omission/null, metadata omission, saving/error states |
| Delete dialog | Confirmation, cancellation, pending controls, API error |
| Admin transport | Courses, roots, children, CRUD paths, ID encoding, methods/bodies, credentials, JSend unwrap/errors |
| Admin page | Prompts, lazy depth, empty state, CRUD refresh, failure states, stale-response protection |
| Sidebar | Capability-present and capability-absent visibility |
| Learner transport/pages | JSend, enrollment, API errors, order, empty state, stale responses, API adjacency |
| Renderer/URL | All frozen types, fallbacks, metadata states, immutability, order, escaping, safe URL behavior |

Focused result: 9 files, 63 tests, all passing.
