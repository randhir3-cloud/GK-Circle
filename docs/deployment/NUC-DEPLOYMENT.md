# NUC Deployment

GK Circle v2 deploys to `/home/randhir/apps/gk-circle-v2` using `docker-compose.nuc.yml`.

## Safety model

- The paused previous GK Circle stack remains intact as rollback.
- GK Circle v2 uses its own containers, network, volumes, and fresh database.
- Production secrets exist only in the server `.env`.
- The public `gkcircle.com` route changes only after candidate verification.

## First deployment

1. Clone `randhir3-cloud/GK-Circle-v2` into the v2 directory.
2. Check out the approved commit.
3. Copy `.env.example` to `.env` and replace every placeholder with a strong value.
4. Run `docker compose --env-file .env -f docker-compose.nuc.yml config --quiet`.
5. Build and start the candidate stack.
6. Verify database, migration, Kratos, API, web, Redis, and MinIO health.
7. Verify a real registration/login and test workflow.
8. Switch the Cloudflare/public reverse-proxy route to the candidate web/API ports.
9. Verify `https://gkcircle.com` and retain the old route for rollback.

## Rollback

If critical verification fails, restore the previous public route and stop the v2 candidate without removing its volumes. Do not delete either stack during initial cutover.
