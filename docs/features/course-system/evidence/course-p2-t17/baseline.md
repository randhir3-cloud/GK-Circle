# T17 Baseline

- Starting commit: `eeac599f05eaf936c7f61db4a3deeac3c9063f59`
- Branch: `chore/ci-verification`
- Playwright config: `app/playwright.config.ts`
- Project: `chromium`
- Existing helpers reused:
  `app/tests/e2e/fixtures/authenticated-user.ts`
- Existing E2E suite:
  `app/tests/e2e/runtime-verification.spec.ts`
- Focused T17 baseline: specification absent
- Full Playwright baseline: 10/10 passed after the sandbox permitted Chromium

Environment proof before mutation:

- `PLAYWRIGHT_BASE_URL`: `http://localhost:3000`
- `PLAYWRIGHT_API_BASE_URL`: `http://localhost:3010`
- Kratos public/admin: localhost ports 4433/4434
- Compose project: `gkcirclev2`
- PostgreSQL: internal Compose network only, no published host port
- API and API/database health: HTTP 200
- Web, Kratos, and Mailpit health: HTTP 200

The starting tree was already heavily dirty from earlier Course-system work.
The complete, literal pre-change `git status --short` output is preserved in
[baseline-status.txt](baseline-status.txt). T17 ownership is isolated in
[changed-files.md](changed-files.md); no pre-existing file is attributed to
T17.
