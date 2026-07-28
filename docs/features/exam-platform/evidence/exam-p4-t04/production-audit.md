# EXAM-P4-T04 — Production Audit

## Risk summary

Additive snapshot tables and editor-gated APIs. No attempt lifecycle, scoring, or learner player. Snapshots are append-only compositions; live bank/collection edits do not mutate existing rows.

## Integrity

- Create is transactional (snapshot header + items).
- Unique `(snapshot_id, question_id)` and `(snapshot_id, position)`.
- DYNAMIC metadata-pending collections cannot be snapshotted until taxonomy resolution exists (EXAM-P10).

## Security

- All snapshot routes require quiz edit access for this contract-hook phase.
- Learner-safe endpoint omits answer keys (unit-proven).
- Future P5 may introduce authenticated learner attempt reads bound to attempt ownership.

## Rollback

`gk-circle migrate down 1` drops snapshot tables. No impact on questions/collections.

## Unverified

Live DB migration apply and browser smoke deferred to operator/manual review. Phase 4 not closed.
