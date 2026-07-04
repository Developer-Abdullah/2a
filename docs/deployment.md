# Deployment, Hosting & Operations

## Quick launch (single "Double A" store)

The store runs as one tenant. Everything sits behind **Caddy**, which gets HTTPS certificates
automatically. On a fresh VPS (4 vCPU / 8 GB, Ubuntu, Docker installed):

1. **DNS.** Point four `A`/`AAAA` records at the server: `doublea.store`, `api.`, `admin.`,
   `admin-api.` (any domain — set them in `.env`).
2. **Clone + configure.**
   ```bash
   git clone <repo> /opt/doublea && cd /opt/doublea
   cp .env.prod.example .env
   bash scripts/gen-secrets.sh >> secrets.txt   # paste the lines into .env, then rm secrets.txt
   # edit .env: set the *_DOMAIN vars, ACME_EMAIL, STOREFRONT_URL, payment + SMTP keys
   ```
3. **Launch.**
   ```bash
   docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
   ```
   Caddy issues certs on first request. Storefront → `https://doublea.store`, admin →
   `https://admin.doublea.store`.
4. **Firewall (ufw).** Allow `22, 80, 443`; block `5432, 6379, 8080, 8081, 3000, 3001` from the
   public internet — they only talk over the Docker network; Caddy is the sole public entry.
5. **Payment webhooks.** In the Paymob / MyFatoorah dashboards set the webhook URL to
   `https://api.doublea.store/billing/webhook/paymob` (and `/myfatoorah`). HTTPS is required.
6. **Backups.** `crontab -e` → `30 3 * * * cd /opt/doublea && bash scripts/backup-db.sh`.

Key files: [`deploy/Caddyfile`](../deploy/Caddyfile), `docker-compose.prod.yml`,
[`.env.prod.example`](../.env.prod.example), `scripts/gen-secrets.sh`, `scripts/backup-db.sh`.

**Single-store resolution:** `STORE_TENANT_SLUG=store` forces every request (including header-less
payment webhooks and the device's `itms-services` OTA fetch) to resolve to the one store — set it or
those callers 400.

---


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
