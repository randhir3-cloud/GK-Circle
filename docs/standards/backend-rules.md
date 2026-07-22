# Go API Rules

- Use Go 1.23 and existing Fiber patterns under `api/`.
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
