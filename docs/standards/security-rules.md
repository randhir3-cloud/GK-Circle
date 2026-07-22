# Security Rules

- Ory Kratos remains the authentication and session provider.
- Authorization must be enforced by the Go API, not trusted from the client.
- Commit only placeholder environment examples. Real `.env` files stay outside Git.
- Use strong unique production values for database, Redis, JWT, Kratos cookie/cipher, MinIO, SMTP, and admin credentials.
- Production must use HTTPS and secure cookie settings.
- Restrict Kratos admin, PostgreSQL, Redis, and MinIO administration ports from public exposure.
- Validate uploads, CSV input, WebSocket messages, identifiers, and return URLs.
- Back up databases before migrations and verify restoration procedures.
- Do not disable validation, authentication, CORS protections, or tests to make a deployment pass.
- Scan staged changes for private keys and provider tokens before every push.
