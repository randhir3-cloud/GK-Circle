# GK Circle — Developer Verification Guide

This guide defines the canonical verification policy, prerequisites, setup instructions, commands, and troubleshooting guidelines for validating the GK Circle platform across different operating systems.

---

## 1. Canonical Verification Policy

To ensure that pull requests and merges do not introduce regressions:
* **Docker is the canonical local environment** for backend verification on all platforms.
* **Native Ubuntu (Linux) execution is used in CI** for fast, native compilation and isolated testing.
* **No production code or runtime logic** must be modified to facilitate testing.
* **Local verification** must pass entirely before code is proposed for code review or merged.

---

## 2. Verification Matrix

| Platform | Backend Unit Tests | Race Tests | Frontend | E2E (Playwright) |
| :--- | :--- | :--- | :--- | :--- |
| **Windows Native** | Optional; may be blocked by WDAC/SAC | Optional; may be blocked | Supported | Supported |
| **Windows Docker** | **Canonical** | **Canonical** | Supported | Supported |
| **WSL (Linux)** | Supported | Supported | Supported | Supported |
| **Linux (Native)** | Supported | Supported | Supported | Supported |
| **macOS Docker** | Supported | Supported | Supported | Supported |

---

## 3. Prerequisites

Ensure you have the following installed locally:
* **Docker Desktop** (Windows/macOS) or **Docker Engine with Compose** (Linux)
* **Node.js** (v22.9.0 - see [app/.tool-versions](file:///e:/GK%20Circle%20v2/app/.tool-versions))
* **Go** (v1.23.0 - see [api/go.mod](file:///e:/GK%20Circle%20v2/api/go.mod)) - only required for native host runs.
* **Make** (standard build tool, optional but recommended)

---

## 4. Platform Workflows

### Windows Docker Workflow (Canonical for Windows)
Due to host-level security controls (such as Windows Smart App Control or AppLocker), Go test binaries compiled on-the-fly inside the `%TEMP%` folder may be blocked. To bypass this safely without weakening host security, run all backend tests inside the canonical `api-verify` Docker container.

1. Ensure Docker Desktop is running.
2. From the repository root, run the Makefile verification target:
   ```bash
   make -C api docker-verify
   ```
3. Or run specific docker compose commands:
   ```bash
   docker compose --profile verify run --rm api-verify go test ./...
   ```

### Windows WSL Workflow
If you prefer not to use Docker, you can run tests natively inside the Windows Subsystem for Linux (WSL), which is unaffected by Windows host Code Integrity policies.

1. Open your WSL terminal (e.g., Ubuntu).
2. Navigate to the project directory:
   ```bash
   cd /mnt/c/path/to/GK-Circle-v2/api
   ```
3. Run native Go commands:
   ```bash
   go test ./...
   go test -race ./...
   go vet ./...
   ```

### Linux & macOS Workflows
On Linux and macOS, you can choose to run tests either natively on the host or inside Docker:

**Docker (Canonical)**:
```bash
make -C api docker-verify
```

**Native Host**:
```bash
cd api
go vet ./...
go test ./...
go test -race ./...
```

---

## 5. Command Reference

### Backend Verification (Go API)
Run these commands inside the `api/` directory (or use `make -C api <target>` from root):

* **Run Vet**:
  ```bash
  make docker-vet
  # Or: docker compose --profile verify run --rm api-verify go vet ./...
  ```
* **Run Unit Tests**:
  ```bash
  make docker-test
  # Or: docker compose --profile verify run --rm api-verify go test ./...
  ```
* **Run Race Detector**:
  ```bash
  make docker-test-race
  # Or: docker compose --profile verify run --rm api-verify go test -race ./...
  ```
* **Run All Verification Targets**:
  ```bash
  make docker-verify
  ```

### Frontend Verification (Nuxt App)
Run these commands inside the `app/` directory:

* **Install Dependencies**:
  ```bash
  npm ci
  ```
* **Run Lint**:
  ```bash
  npm run lint
  ```
* **Production Build Validation**:
  ```bash
  npm run build
  ```

### E2E Integration Verification (Playwright)
Run these commands to validate full integration E2E flows:

1. Create a local environment file `.env` containing your safe development credentials (do not commit this).
2. Start the local Docker Compose stack:
   ```bash
   docker compose up -d --build
   ```
3. Install Playwright browser dependencies (first time only):
   ```bash
   cd app
   npx playwright install --with-deps
   ```
4. Run E2E tests:
   ```bash
   npx playwright test
   ```
5. Tear down the stack:
   ```bash
   docker compose down -v --remove-orphans
   ```

---

## 6. Continuous Integration (CI) Job Behavior

The CI pipeline is defined in [.github/workflows/verify.yaml](file:///e:/GK%20Circle%20v2/.github/workflows/verify.yaml). It is split into three isolated, sequential jobs:
1. **`backend-verify`**:
   * Executed on native Ubuntu runners for performance.
   * Runs `go vet`, `go test`, and `go test -race`.
   * Smoke-tests the `api-verify` docker compose service to prevent local environment breakage.
2. **`frontend-verify`**:
   * Uses Node version extracted dynamically from `app/.tool-versions`.
   * Runs `npm ci`, `npm run lint`, and `npm run build`.
3. **`playwright-e2e`**:
   * Runs only after both `backend-verify` and `frontend-verify` pass.
   * Starts the compose stack and polls health check readiness.
   * Executes Playwright tests.
   * Collects container logs and system diagnostics on failure.
   * Cleans up the stack.

---

## 7. Diagnostics and Troubleshooting

### Health & Readiness Polling
During local or CI E2E runs, services must be checked for readiness before running Playwright. The canonical check endpoints and service names are:
* **PostgreSQL (`db` service)**: Port `5432` internally. Check with `docker compose exec -T db pg_isready -U gk_circle -d gk_circle`.
* **Redis (`redis` service)**: Port `6379` internally. Check with `docker compose exec -T redis redis-cli ping`.
* **Ory Kratos (`kratos` service)**: Port `4433` public endpoint. Check with `http://localhost:4433/health/ready`.
* **API (`api` service)**: Port `3010` (mapped to container `3000`). Check with `http://localhost:3010/api/healthz/`.
* **Web (`web` service)**: Port `3000` (mapped to container `5000`). Check with `http://localhost:3000/`.
* **Mailpit (`mailpit` service)**: Port `8025`. Check with `http://localhost:8025/`.

### Log and Diagnostic Collection
If E2E tests fail or a service does not start, collect debug logs *before* stopping the containers:
```bash
# Get status of all containers
docker compose ps --all

# Collect logs without console color codes
docker compose logs --no-color > compose-debug.log
```

### Docker Cache Cleanup
If you experience caching issues or want to release disk space:
```bash
# Remove unused verification volumes
docker volume rm gk-circle-v2_go-build-cache gk-circle-v2_go-module-cache

# Remove all unused compose volumes
docker compose down -v
```

### WDAC / Smart App Control Troubleshooting
If you encounter `An Application Control policy has blocked this file` on Windows when running host commands:
* **Reason**: Windows Smart App Control (SAC) blocks any unsigned binary with no reputation. Go test binaries are unsigned and compiled on-the-fly.
* **Correction**: Do **not** disable SAC or AppLocker. Switch to the Windows Docker workflow (`make -C api docker-verify`) or run within WSL.
* **Event log location**: Check `Applications and Services Logs -> Microsoft -> Windows -> CodeIntegrity -> Operational` for Event ID `3077` blocks.
