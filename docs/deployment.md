# Deployment, Hosting & Operations

## Services

| Service | Port | Notes |
|---------|------|-------|
| core-api (`cmd/api`) | 8080 | Public API: enrollment, activation, OTA install, ratings, billing webhooks. |
| admin-api (`cmd/admin`) | 8081 | Dashboard/owner API. |
| signing-worker (`cmd/signing`) | — | Consumes the asynq signing queue; runs zsign. |
| notifier-worker (`cmd/notifier`) | — | Push dispatch + hourly billing reminder/sweep. |
| analytics-worker (`cmd/analytics`) | — | Async analytics. |
| admin-dashboard (Next.js) | 3000 | Owner + tenant admin UI. |
| postgres 16, redis 7, minio | — | Data / queue / object storage. |

## Hosting & domain checklist (Phase 0.3)

1. **VPS.** Hetzner (best price/perf for a bootstrapped EU/MENA SaaS) or DigitalOcean. Start:
   4 vCPU / 8 GB / 80 GB SSD (Postgres + Redis + MinIO + API + workers + dashboard on one box). Move
   Postgres and object storage off-box (managed PG + B2/Wasabi) as you grow.
2. **DNS.** `A`/`AAAA` for the marketing site + `api.`, `dashboard.`, and a wildcard `*.` for
   per-merchant subdomains (tenant resolution uses the first host label).
3. **Reverse proxy.** **Caddy** — automatic HTTPS via Let's Encrypt with near-zero config, including
   on-demand TLS for wildcard/per-tenant hosts. (Nginx is fine but you manage certs yourself.)
4. **TLS.** Let's Encrypt via Caddy auto-HTTPS. Confirmed simplest; wildcard needs a DNS-01
   challenge (Caddy DNS plugin for your registrar).
5. **Firewall (ufw).** Allow 22 (SSH, ideally key-only + fail2ban), 80, 443. **Block** 5432
   (Postgres), 6379 (Redis), 9000 (MinIO) from the public internet — they talk over the Docker
   network only.

## CI/CD

`.github/workflows/deploy.yml`: on push to `main` → golangci-lint + `go test -race -cover` + Next.js
build, then build & push service images to ghcr.io. Deploy by pulling the new images on the VPS and
`docker compose -f docker-compose.prod.yml up -d`. For a solo dev, a pull+restart over SSH (or a
Watchtower-style auto-pull) is sufficient; add blue/green only when uptime demands it.

## Health checks

Both APIs expose `GET /health` (liveness). `GET /healthz` (readiness) additionally pings the DB —
point Uptime Kuma and the reverse proxy at `/healthz` so a DB outage marks the service unhealthy.

## Backups

`scripts/backup.sh` runs `pg_dump` (custom format), keeps **7 daily + 4 weekly** locally, and
uploads to the S3-compatible store under a **separate bucket/prefix** from the IPA vault. Schedule it
via cron:

```
0 3 * * *  /opt/app/scripts/backup.sh >> /var/log/pg-backup.log 2>&1
```

Restore: `pg_restore --clean --if-exists -d "$DB_DSN" <dump>`.

## Monitoring (Uptime Kuma)

Self-hosted, free. Add HTTP(s) monitors for `https://api.<domain>/healthz`,
`https://admin-api.<domain>/healthz`, and the dashboard root; alert to email/Telegram. See
`docs/monitoring.md`.
