# T17 Source Mapping

| Verified behavior | Production source | E2E source |
|---|---|---|
| Course and direct-child selection | Admin LearningItem page and Course selectors | `findDeepPath`, `selectDeepNode` |
| Admin DRAFT creation | Admin editor/page and Course transport composable | Create response and persisted row assertions |
| Draft exclusion | Learner list page/composable and Go learner repository | Pre-publication learner list assertion |
| Publication | Admin editor PATCH flow | PATCH response and row state assertion |
| API ordering | Learner list response and list page | Exact DOM ID array equality |
| Learner detail and renderer | Learner detail page and renderer | Title and `No content available.` assertions |
| Previous/next authority | Learner detail API and page | URLs compared only with API neighbor IDs |
| Enrollment | Learner enrollment controller/composable/UI | Existing enrollment CTA when required |
| Signed-out denial | Kratos middleware and learner page error state | HTTP 401 plus visible alert |
| Cleanup | Admin DELETE endpoint and PostgreSQL | DELETE 200, GET 404, DB zero count |

The T17 test uses the repository’s existing Playwright engine and login helper.
It introduces no parallel E2E framework, runtime mock, response interception,
production hook, frontend data source, backend API, DTO, or persistence change.
