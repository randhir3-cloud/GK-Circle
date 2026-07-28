# COURSE-P2-T06 API Endpoints

All routes require `KratosAuthenticated` plus in-controller quiz-admin allowlist
(`authorizeCourseAdmin`). Owner is derived from authenticated identity.

| Method | Path | Handler |
|---|---|---|
| POST | `/api/v1/admin/courses/:course_id/nodes/:node_id/learning-items` | Create |
| GET | `/api/v1/admin/courses/:course_id/nodes/:node_id/learning-items` | List (`position ASC, id ASC`) |
| GET | `/api/v1/admin/courses/:course_id/nodes/:node_id/learning-items/:item_id` | Get |
| PATCH | `/api/v1/admin/courses/:course_id/nodes/:node_id/learning-items/:item_id` | Update (presence-aware) |
| DELETE | `/api/v1/admin/courses/:course_id/nodes/:node_id/learning-items/:item_id` | Delete |

## DTOs

### Create (`ReqCreateAdminLearningItem`)

Required: `title`, `item_type`  
Optional: `description`, `metadata`  
Not accepted as authority: `course_node_id` (path node wins)

### Update (`ReqUpdateAdminLearningItem`)

Presence-aware: `title`, `item_type`, `description`, `metadata`  
Empty patch → 400 (`ErrLearningItemUpdateRequired`)

## Repository mapping

| HTTP | Repository |
|---|---|
| Create | `CreateLearningItem` |
| List | `ListLearningItemsByNode` |
| Get | `GetLearningItemByID` |
| Update | `UpdateLearningItem` |
| Delete | `DeleteLearningItem` |

## Error mapping (selected)

| Model error | HTTP |
|---|---|
| `ErrLearningItemNotFound` | 404 |
| `ErrLearningItemNodeNotFound` / `ErrCourseNodeNotFound` | 404 |
| `ErrLearningItemCrossCourse` / `ErrCourseNotFound` | 404 |
| `ErrLearningItemUpdateRequired` | 400 |
| Metadata / placeholder / visibility errors | 400 |
| Missing identity | 401 |
| Non-allowlisted | 403 |
