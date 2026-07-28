# EXAM-P1-T02 — Repository Audit Evidence

Status: VERIFIED
- Started: 2026-07-27
- Verified: 2026-07-27
- Re-audited: 2026-07-27 (authorisation refresh)

## Frozen Acceptance

- [x] Capability status matrix is recorded under exam-platform evidence (COMPLETE / PARTIAL / UNSAFE / MISSING / OUT OF SCOPE).
- [x] Gaps against the PCS MVP study loop are listed with file-path references.
- [x] Evidence pack exists under `docs/features/exam-platform/evidence/exam-p1-t02/`.

## Standards loaded

AGENTS.md, CLAUDE.MD, docs/standards/index.md, architecture-rules, security-rules, ai-rules, testing-rules, backend-rules, frontend-rules, course-rules, live-exam-rules, ADR-024.

## Task understanding

EXAM-P1-T02 is a **repository audit and evidence formalisation** task. It does not authorise production feature work. It records what exists, what is partial, what is unsafe, what is missing, and what is out of scope against the PCS study loop defined in the Product Roadmap and ADR-024.

## Audit baseline

| Field | Value |
|---|---|
| Branch | `chore/ci-verification` |
| Commit (HEAD) | `eeac599f05eaf936c7f61db4a3deeac3c9063f59` |
| Working tree | Dirty — exam-platform P1/P2 work present uncommitted at audit time |
| Course System P2 | VERIFIED (prerequisite) |
| ADR-024 | Accepted (EXAM-P1-T01) |

Audit method: read-only inspection of migrations, routes, models, controllers, Nuxt pages, and governance docs. See [audit-command-log.md](audit-command-log.md).

## Evidence index

| Document | Purpose |
|---|---|
| [capability-matrix.md](capability-matrix.md) | Domain capability statuses |
| [pcs-mvp-gaps.md](pcs-mvp-gaps.md) | Study-loop gaps with file paths |
| [key-artifacts.md](key-artifacts.md) | Canonical code and doc references |
| [audit-command-log.md](audit-command-log.md) | Commands and inspection steps |
| [production-audit.md](production-audit.md) | Deployment impact of this task |

## Summary findings

1. **Course System foundation is strong** — courses, nodes, learning items, enrollments, and learner reads are implemented and tested.
2. **P1 Course Builder path exists in tree** — admin builder, publish API, learner catalog, enroll publish gate (implementation attributed to EXAM-P1-T03; present at audit).
3. **PCS examination loop remains incomplete** — collections, self-paced runtime, player, analytics, revision, and coverage are still future phases.
4. **Inherited live-quiz review endpoints remain UNSAFE** — unauthenticated routes expose correct answers (EXAM-P2-T03 / EXAM-P7).
5. **Self-paced schema is orphan** — `assessment_*` tables exist; no Go consumer models at audit time.

## Production source modified by EXAM-P1-T02: NO

(Documentation and audit evidence only.)
