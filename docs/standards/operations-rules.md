# Operations Rules

- NUC application path: `/home/randhir/apps/gk-circle-v2`.
- Production uses `docker-compose.nuc.yml` and a server-local `.env`.
- Deploy only committed code from `randhir3-cloud/GK-Circle`.
- Pull with fast-forward only.
- Back up PostgreSQL before migrations once production data exists.
- Run Go SQL migrations before serving new code.
- Verify container health, public HTTPS, authentication, API, persistence, and WebSockets after deployment.
- Keep the previous production stack intact during initial v2 cutover.
- Rollback traffic to the previous stack if critical verification fails.
- Never delete containers, volumes, databases, or backups without explicit approval.
- Report commit SHA, image/build state, migration outcome, health results, and rollback path.
