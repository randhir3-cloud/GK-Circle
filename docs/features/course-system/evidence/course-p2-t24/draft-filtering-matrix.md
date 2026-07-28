# COURSE-P2-T24 Draft Filtering Matrix

| Surface | Published item | Draft item | Expected result |
|---|---|---|---|
| Learner list (`ListPublishedLearningItemsByNode`) | Included | Excluded | 200; published-only array |
| Learner get (`GetPublishedLearningItemByID`) | Returned | Hidden | 200 / 404 |
| Admin list (`ListLearningItemsByNode`) | Included | Included | 200 |
| Admin get (`GetLearningItemByID`) | Returned | Returned | 200 |

## Endpoints

- Learner: `GET /api/v1/learner/courses/:course_id/nodes/:node_id/learning-items[+/:item_id]`
- Admin: `GET /api/v1/admin/courses/:course_id/nodes/:node_id/learning-items[+/:item_id]`

## Filter ownership

SQL `publish_state = PUBLISHED` is applied only on learner published repository methods. Controllers do not post-filter.
