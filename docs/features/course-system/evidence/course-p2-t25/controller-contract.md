# COURSE-P2-T25 Controller Contract

The learner controller remains a transport boundary:

- Kratos middleware rejects unauthenticated list and detail requests with 401
  before any repository query.
- Authenticated list calls the published-only repository method and serializes
  rows in repository-returned order.
- Authenticated detail calls both the published item lookup and published
  adjacent-item lookup, preserving the T22 `learning_item` / `previous` /
  `next` wrapper.
- The controller maps not-found to 404 and persistence failures to 500 using
  stable public fields.
- Learner DTOs omit persistence ordering and ownership fields.

The adversarial sqlmock rows intentionally label a returned row `DRAFT`. Its
presence in the response proves the controller does not apply a second
publish-state filter after repository delegation.
