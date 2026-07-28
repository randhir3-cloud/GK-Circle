# EXAM-P4-T02 — Production Audit

## Risk summary

Frontend-only feature mounted behind existing quiz edit permission (`canEditQuiz`). All mutations go through authenticated T01 collection APIs. No schema changes.

## Security

- Collection routes already require Kratos + quiz edit access.
- Resolve preview displays question IDs and bank titles only; answer keys are not requested from resolve.
- No alternate client-side collection store.

## Rollback

Remove or hide `VisualTestBuilder` from the quiz manage page. Collection data created via the UI remains in T01 tables and can be managed via API.

## Unverified at runtime

IronBee browser MCP (`ironbee-dt-browser`) was not available in this agent session. Unit/component tests + production build verify client wiring; full browser smoke against a live stack is deferred to manual review.
