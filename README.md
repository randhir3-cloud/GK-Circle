# GK Circle

GK Circle is a State PCS examination preparation platform for practice sets, mock examinations, subject tests, current-affairs tests, previous-year questions, live competitions, rankings, and performance analytics.

The project is built on the open-source [Improwised/jovVix](https://github.com/Improwised/jovVix) real-time quiz foundation and is being adapted for PCS examination workflows while preserving its established quiz, session, scoring, authentication, and reporting systems.

## Current capabilities

- Create and manage question sets through the UI or CSV imports.
- Run individual practice sessions and real-time multiplayer tests.
- Support text, image, code, single-answer, and survey-style questions.
- Publish selected tests for guest access.
- Provide live scoring, leaderboards, answer analysis, and reports.
- Share tests with other registered users.
- Register, sign in, verify email, and recover accounts through Ory Kratos.
- Operate through Docker Compose on desktop and server environments.

## PCS product direction

The existing quiz engine is the shared technical engine for:

- daily practice;
- prelims mock examinations;
- subject and topic tests;
- current-affairs tests;
- previous-year-question practice;
- timed live competitions.

Commission, state, exam stage, subject, topic, language, year, difficulty, and test mode will remain configurable rather than being hardcoded to one State PCS syllabus.

## Architecture

| Area | Technology |
|---|---|
| Web | Nuxt 3, Vue 3, Pinia, Tailwind CSS |
| API | Go 1.23, Fiber v2 |
| Database | PostgreSQL 15 with versioned SQL migrations |
| Authentication | Ory Kratos |
| Cache/live coordination | Redis-compatible service |
| Real-time interaction | WebSockets |
| Object storage | MinIO |
| Local email | Mailpit |
| Runtime | Docker Compose |

The source is organized under `app/` for the Nuxt frontend and `api/` for the Go API.

## Quick start

Requirements:

- Docker Desktop or Docker Engine with Compose
- Git

```powershell
git clone https://github.com/randhir3-cloud/GK-Circle-v2.git
cd GK-Circle-v2
Copy-Item .env.example .env
```

Replace every `change_me` value in `.env`, then start the stack:

```powershell
docker compose up --build
```

Default local endpoints:

| Service | URL |
|---|---|
| Web | `http://localhost:3000` |
| API | `http://localhost:3010/api/v1` |
| Kratos public API | `http://localhost:4433` |
| Mailpit | `http://localhost:8025` |

Never commit `.env` or database backups.

## Development

Frontend:

```powershell
cd app
npm ci
npm run dev
```

API (requires Go 1.23 and supporting services):

```powershell
cd api
go run . migrate up
go run . api
```

## Verification Workflow

For local verification, Docker is the canonical environment for backend validation.

### Makefile Targets
You can run verification tasks depending on your current working directory:

* **From the repository root**:
  ```powershell
  make -C api docker-vet
  make -C api docker-test
  make -C api docker-test-race
  make -C api docker-verify  # Runs all three checks
  ```
* **From the `api/` directory**:
  ```powershell
  make docker-vet
  make docker-test
  make docker-test-race
  make docker-verify  # Runs all three checks
  ```

### Direct Docker Compose Commands
If you do not have `make` installed, you can run the commands directly using `docker compose` from the repository root:

* **Run all checks (vet, test, race) in a single run**:
  ```powershell
  docker compose --profile verify run --rm api-verify sh -c "go vet ./... && go test ./... && go test -race ./..."
  ```
* **Run individual checks**:
  ```powershell
  docker compose --profile verify run --rm api-verify go vet ./...
  ```
  ```powershell
  docker compose --profile verify run --rm api-verify go test ./...
  ```
  ```powershell
  docker compose --profile verify run --rm api-verify go test -race ./...
  ```

### Windows Security Policy Warning
> [!IMPORTANT]
> Windows Smart App Control or WDAC may block unsigned temporary Go test executables generated under `%TEMP%\go-build`. This is a host application-control policy restriction, not a GK Circle test failure. Docker or WSL is the supported backend verification environment on affected Windows systems. Do not disable Smart App Control, Defender, WDAC, or AppLocker on your host system.

## NUC deployment

The NUC deployment uses `docker-compose.nuc.yml` and the scripts under `scripts/`. Production configuration belongs only in `/home/randhir/apps/gk-circle-v2/.env` on the server.

The previous GK Circle production stack and its data must remain intact until GK Circle v2 passes health, authentication, persistence, WebSocket, and public HTTPS verification.

## Repository remotes

- `origin`: `https://github.com/randhir3-cloud/GK-Circle-v2.git`
- `upstream`: `https://github.com/Improwised/jovVix.git`

Use `upstream` to review suitable maintenance changes from the original project. Do not overwrite GK Circle-specific product decisions during upstream synchronization.

## Licence and attribution

GK Circle retains the GNU Affero General Public License v3.0 supplied with the inherited project. See [LICENSE.txt](LICENSE.txt).

Original jovVix project copyright and contributor attribution remain with the original authors and contributors. GK Circle modifications are maintained in this repository. Removal or alteration of licence and attribution material requires legal review.
