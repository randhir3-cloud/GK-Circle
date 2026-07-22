# GK Circle Architecture

## Runtime

```text
Browser
  -> Nuxt 3 web application
      -> Go/Fiber REST API
      -> Go WebSocket endpoint
      -> Ory Kratos public API

Go API
  -> PostgreSQL application schema
  -> Redis-compatible cache/coordination
  -> MinIO object storage

Ory Kratos
  -> PostgreSQL kratos schema
  -> SMTP provider (Mailpit locally)
```

## Source layout

- `app/`: Nuxt frontend, components, pages, stores, composables, and tests.
- `api/`: Go API, commands, models, handlers, WebSockets, and SQL migrations.
- `api/pkg/kratos/`: authentication schema and configuration.
- `scripts/`: local and NUC operational tooling.
- `docker-compose.yaml`: base development stack.
- `docker-compose.override.yml`: local port and tooling overrides.
- `docker-compose.nuc.yml`: isolated NUC candidate/production stack.

## Data

The v2 deployment starts with a fresh `gk_circle` PostgreSQL database. The paused previous GK Circle database has a different schema and is preserved as rollback; it is not attached to this application.

Schema evolution uses the existing ordered SQL migrations in `api/database/migrations/`.

## Product adaptation

The inherited `quiz` model remains the shared assessment engine. PCS-specific taxonomies and modes should extend it additively. Separate practice, mock-exam, current-affairs, and live-test engines are prohibited.
