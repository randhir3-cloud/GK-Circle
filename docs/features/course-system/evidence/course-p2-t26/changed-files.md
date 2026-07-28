# COURSE-P2-T26 Changed Files

## T26-owned documentation

- `docs/api/course-learning-items-v1.md` — canonical API examples and contract.
- `docs/development/modules/course-system/README.md` — canonical API link.
- `docs/development/PROJECT_INDEX.md` — canonical API link and task status.
- Course-system status, handoff, changelog, AI context, phase status, and
  generated status artifacts — updated only for truthful T26 tracking.
- `docs/features/course-system/evidence/course-p2-t26/` — canonical evidence.

## Production-change guard

Baseline commit:
`eeac599f05eaf936c7f61db4a3deeac3c9063f59`

The repository was already dirty in `api/` and `app/` before T26 began,
including the existing Course/LearningItem implementation and frontend work.
That baseline was captured with `git status --short -- api app`.

T26-owned changes under `api/` or `app/`: NONE.

The final guard compares the ending `api/` and `app/` status to the captured
baseline rather than incorrectly attributing pre-existing dirty files to T26.
The ending status matched that baseline.
