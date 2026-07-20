# jovVix — Docker Evaluation Guide

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) ≥ 4.x
- Ports **3000** (web), **3010** (API), **4433** (Kratos), **8025** (Mailpit), **9001** (MinIO) available

## How the Override Works

The upstream file `docker-compose.yaml` is **not modified**.
`docker-compose.override.yml` is automatically applied by Docker Compose on every `docker compose` command, so no extra flags are needed.

## Quick Start

```powershell
# 1. Copy and review env files
Copy-Item .env.example .env
Copy-Item api\.env.example api\.env.docker   # if not already present

# 2. Start the full stack
docker compose up -d

# 3. Open the quiz platform
#    http://localhost:3000
```

## Default Credentials

| Service     | URL                       | Credentials                    |
|-------------|---------------------------|--------------------------------|
| Web UI      | http://localhost:3000     | Register on first use          |
| API         | http://localhost:3010     | JWT via Kratos login           |
| Kratos      | http://localhost:4433     | Identity management            |
| Mailpit     | http://localhost:8025     | View emails locally            |
| MinIO       | http://localhost:9001     | ROOTNAME / CHANGEME123         |

## Commands

```powershell
# Start stack
docker compose up -d

# Stop stack
docker compose down

# Full reset (removes all data)
docker compose down -v

# View all logs
docker compose logs -f

# View specific service logs
docker compose logs api
docker compose logs web
docker compose logs db

# Check service health
docker compose ps
```

## Optional: Database Admin Tools

```powershell
# Start pgAdmin alongside the main stack
docker compose --profile tools up -d

# pgAdmin URL: http://localhost:5051
# Email:    admin@local.dev  
# Password: pgadmin_secret
#
# Add server in pgAdmin:
#   Host:     db
#   Port:     5432
#   Database: jovvix
#   Username: jovvix
#   Password: jovvix
```

## How to Reset Database

```powershell
docker compose down -v
docker compose up -d
```

## Backup & Restore

```powershell
.\backup.ps1
.\restore.ps1 .\backups\jovvix_db_YYYYMMDD_HHMMSS.sql
```

## Network & Ports

| Service | Host Port | Notes                        |
|---------|-----------|------------------------------|
| Web UI  | 3000      | Nuxt 3 frontend              |
| API     | 3010      | Go backend REST API          |
| Kratos  | 4433      | Public identity endpoint     |
| Kratos  | 4434      | Admin endpoint               |
| Mailpit | 8025      | Email UI                     |
| MinIO   | 9001      | Object storage console       |
| DB      | internal  | Not exposed to host          |
| Redis   | internal  | Not exposed to host          |

All services share the isolated `jovvix-network` bridge network.
