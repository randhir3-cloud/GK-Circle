# Architecture Rules

The existing architecture is a Go/Fiber API and Nuxt/Vue web application supported by PostgreSQL, Ory Kratos, Redis, WebSockets, MinIO, Mailpit, and Docker Compose.

- Extend existing quiz, question, session, scoring, sharing, authentication, and reporting systems.
- Do not introduce parallel test engines or authentication systems.
- Keep business rules and authorization in the API.
- Keep the web application focused on interaction, presentation, and client state.
- Add SQL migrations; never rewrite migrations that may have run.
- Preserve backward compatibility unless an approved plan documents the break.
- New external providers, major frameworks, top-level services, destructive migrations, or cross-domain designs require an ADR and approval.
- Search `api/` and `app/` before creating new packages, composables, stores, components, or endpoints.
- PCS examination modes are configurations of the shared quiz engine, not separate engines.
- Upstream jovVix history and AGPL attribution remain part of the architecture provenance.
