# Testing Rules

Applicable verification includes:

- `go test ./...` for the API;
- `npm run lint` for the web application;
- `npm test -- --run` for frontend unit tests;
- `npm run build` for production compilation;
- `docker compose config --quiet` for Compose;
- container health checks and real authentication/test workflows for deployment.

Do not call a feature or deployment complete while relevant checks fail. Distinguish inherited baseline failures from regressions with evidence. Production verification must cover registration/login, practice/test creation, joining, answering, scoring, results, persistence, WebSockets, and rollback where affected.
