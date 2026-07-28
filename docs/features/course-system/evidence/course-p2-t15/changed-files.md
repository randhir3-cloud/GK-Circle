# COURSE-P2-T15 Changed Files

## T15 test changes

- `api/controllers/api/v1/learning_item_backend_suite_test.go`
  - Focused admin CRUD, validation, authorization, exact JSend, and public error
    regression coverage.
- `api/models/learning_item_backend_suite_test.go`
  - Accepted block-enum, unsupported-type, create-default, and partial-update
    regression coverage.
- `api/controllers/api/v1/learner_learning_item_controller_test.go`
  - Strengthened exact learner JSend and public response-field assertions in a
    pre-existing untracked Course-system test file.

## T15 evidence and minimum tracking

- `docs/features/course-system/evidence/course-p2-t15/`
- `docs/development/modules/course-system/phases/phase-02-learning-items.md`
- `docs/development/modules/course-system/CURRENT_STATUS.md`
- `docs/development/modules/course-system/HANDOFF.md`
- Derived status artifacts updated only by the repository sync command.

## Baseline separation

Baseline commit:
`eeac599f05eaf936c7f61db4a3deeac3c9063f59`

The repository already contained a large dirty `api/` and `app/` working tree
before this execution. Its captured porcelain-state digest was:

`130a121cf472cc75feb045af082944748387052865947e806c7789c7a5971fa1`

Those pre-existing production and frontend changes are not attributed to T15.
The final production guard removes only the two new T15 test-file status entries
and compares the remaining `api/`/`app/` state with this baseline.

Final filtered state: 71 lines, matching the 71-line baseline.

Final filtered digest:
`130a121cf472cc75feb045af082944748387052865947e806c7789c7a5971fa1`

Production code changes introduced by T15: NONE.
