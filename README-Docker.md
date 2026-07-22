# GK Circle Docker Guide

## Local stack

Copy `.env.example` to `.env`, replace every `change_me` value, and run:

```powershell
docker compose up -d --build
docker compose ps
```

Local endpoints:

| Service | URL |
|---|---|
| Web | `http://localhost:3000` |
| API | `http://localhost:3010/api/v1` |
| Kratos | `http://localhost:4433` |
| Mailpit | `http://localhost:8025` |
| MinIO console | `http://localhost:9001` |

Database and Redis ports are internal by default. Do not use example credentials outside local development.

## Commands

```powershell
docker compose config --quiet
docker compose up -d --build
docker compose logs -f api web
docker compose ps
```

Use `backup.ps1` before any local migration or restore exercise. Volume removal is destructive and requires explicit confirmation and a verified backup.

## NUC

The isolated NUC stack uses `docker-compose.nuc.yml`, a server-only `.env`, and gateway port 3200 during candidate verification. See `docs/deployment/NUC-DEPLOYMENT.md`.
