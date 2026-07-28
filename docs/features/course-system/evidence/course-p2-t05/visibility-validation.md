# COURSE-P2-T05 Visibility Validation

## Model

Optional per-block field:

```json
"visibility": { "mode": "ALL" }
```

Typed modes:

- `ALL`
- `AUTHENTICATED`
- `PREMIUM`
- `INSTRUCTOR`
- `HIDDEN`

## Rules

| Input | Outcome |
|---|---|
| Omitted | Persist `{"mode":"ALL"}` |
| Valid mode object | Persist that mode |
| `null` / non-object | `ErrLearningItemVisibilityInvalid` |
| Missing `mode` | `ErrLearningItemVisibilityInvalid` |
| Unknown / non-string mode | `ErrLearningItemVisibilityModeInvalid` |

## Pipeline position

Envelope → block → data object → placeholders → **visibility** → persist.

No permission checks or learner filtering are performed.
