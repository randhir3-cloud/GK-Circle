# COURSE-P2-T13 Visibility Runtime Semantics

## Access context

```go
type LearnerVisibilityAccess struct {
  Authenticated     bool
  PremiumAuthorized bool
}
```

Production helper `AuthenticatedLearnerVisibilityAccess()` returns:

- `Authenticated: true` (learner routes already require Kratos identity)
- `PremiumAuthorized: false` (no Course premium entitlement exists in `api/` today)

Unit tests still prove PREMIUM retention when `PremiumAuthorized: true`.

## Stable filter rules (omit without placeholders)

| Mode | Learner output |
|---|---|
| ALL | keep |
| AUTHENTICATED | keep iff `Authenticated` |
| PREMIUM | keep iff `PremiumAuthorized` |
| INSTRUCTOR | omit |
| HIDDEN | omit |
| unknown | omit and continue |

## Binding projection rules

1. Deep-copy metadata: parse → filter → `json.Marshal` → assign newly allocated buffer on a returned item copy. Caller/DB buffers remain unchanged.
2. List projection is atomic: any projection error returns `nil` list + wrapped error (no partial learner list).
3. Malformed stored metadata → wrapped `ErrLearningItemMetadataInvalid` family error.
4. No blocks retained → preserve metadata `version` and emit non-nil empty `blocks` (`[]`, not `null`).

## Wiring

```
Learner GET/List
  → requireAuthenticatedLearner
  → GetPublished / ListPublished (SQL publish filter)
  → ProjectLearningItemForLearner
  → LearnerLearningItemResponse metadata
```

Admin `GetLearningItemByID` / `ListLearningItemsByNode` remain unfiltered.
