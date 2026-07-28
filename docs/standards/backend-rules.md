# Go API Rules

- Use Go 1.25 and existing Fiber patterns under `api/`.
- Keep the module path `github.com/randhir3-cloud/GK-Circle-v2/api`.
- Validate input and enforce authorization server-side.
- Return consistent existing API response formats.
- Use parameterized queries through the established data-access libraries.
- Keep database changes in timestamped up/down SQL migration files.
- Run migrations with `gk-circle migrate up`; do not use Prisma commands.
- Never log credentials, tokens, cookies, personal data, or answer keys unnecessarily.
- Preserve WebSocket authentication, session isolation, and reconnect behavior.
- Add or update Go tests for changed logic and run `go test ./...`.
- Keep health endpoints lightweight and independent of privileged data.

## Real-data completion

- Reads used as completion evidence must come from the established repository and PostgreSQL path, not hardcoded values or in-memory substitutes.
- Writes must use normal validation, authorization, ownership, transaction, and error-handling paths.
- Verify mutations with read-after-write and, where applicable, rollback or conflict behavior.
- Do not return fabricated `2xx` responses, swallow persistence errors, or substitute successful test doubles for real integration evidence.
- Use the existing Go/Fiber and SQL architecture; do not introduce Prisma, NestJS, or a parallel backend.
