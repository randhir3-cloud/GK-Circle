# EXAM-P6-T01 — Production audit

## Breaking Changes: NO

Additive instructions GET and Nuxt learner pages. Existing attempt create/resume/get contracts unchanged.

## Database migration status

- No migration.

## Runtime risk

- Instructions endpoint reuses create entitlement rules (published SELF_PACED / editor preview).
- Does not expose snapshot items or answer keys.
- Attempt shell is informational only until T02.

## Rollback

- Revert application revision; no schema change.
