# Testing Rules

Applicable verification includes:

- `go test ./...` for the API;
- `npm run lint` for the web application;
- `npm test -- --run` for frontend unit tests;
- `npm run build` for production compilation;
- `docker compose config --quiet` for Compose;
- container health checks and real authentication/test workflows for deployment.

Do not call a feature or deployment complete while relevant checks fail. Distinguish inherited baseline failures from regressions with evidence. Production verification must cover registration/login, practice/test creation, joining, answering, scoring, results, persistence, WebSockets, and rollback where affected.

## Real-system completion evidence

Feature completion evidence must exercise the applicable production-shaped path:

`Nuxt UI -> real HTTP route -> Go/Fiber controller -> repository -> PostgreSQL`

Authentication evidence must use Ory Kratos and the applicable server-side authorization policy. A successful UI message, mocked request, intercepted response, in-memory repository, hardcoded array, or fixture-only test does not prove integration or persistence.

Mocks, stubs, fakes, and isolated fixtures are allowed in unit and component tests when the test boundary is explicit. They must not be presented as end-to-end, runtime, integration, persistence, or feature-completion evidence.

## Persisted development seeds

Development seed data is allowed only when all of the following are true:

- it is created by a documented repository seed, migration-safe QA helper, or the normal authorized product workflow;
- it persists in the local PostgreSQL database and satisfies normal constraints;
- it is clearly local or test-only and contains no production personal data, credentials, tokens, or secrets;
- its purpose, prerequisites, identifiers, and cleanup procedure are documented;
- verification still uses the normal authentication, API, repository, and read paths.

Seeds are setup data, not mocked completion evidence. Ad hoc database edits, fabricated API responses, and undocumented fixture injection are not acceptable runtime evidence.

## Read and write verification

For a read feature, verify the database record exists, the backend query returns it, the authorized API contract contains it, the frontend requests and receives the real route, and the rendered value matches the API without a parallel static source. Verify ordering, filtering, visibility, and authorization against server-owned rules where applicable.

For a write feature, verify the browser request, server validation and authorization, the database mutation, read-after-write through the normal repository/API path, refresh or reopen behavior, and database persistence. Verify that no duplicate mutation occurred, cancellation caused no mutation, and errors did not display false success. Where transactions, ordering, ownership, or authorization apply, verify those outcomes as well.

## Authentication automation

Agents must first attempt an authorized local browser workflow with an existing approved browser session, QA identity, or documented QA-account mechanism. They must not invent credentials, extract secrets, bypass Kratos, weaken authorization, disable security controls, or manufacture an identity outside the approved flow. If a required local identity or fixture is missing and creation is authorized and in scope, create it through the documented normal persistence and authentication system and resume verification. Otherwise, record the exact blocker.

## Completion loop

When verification exposes a repository-scope blocker:

1. reproduce and classify it as implementation, configuration, governance, or external;
2. fix it only when the current authorization covers the change;
3. rerun the failed check and the relevant regression checks;
4. resume the original workflow;
5. repeat until the acceptance path passes or an exact stop condition remains.

Formal module certification still requires the human procedure in `module-certification-standard.md`. Authorized agent-run task verification may satisfy task acceptance evidence, but it does not replace formal certification or an explicitly required human sign-off.
