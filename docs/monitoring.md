# Uptime Monitoring (Uptime Kuma)

Uptime Kuma is a free, self-hosted status monitor. Run it next to the stack:

```yaml
# add to your ops compose file (not the app compose)
services:
  uptime-kuma:
    image: louislam/uptime-kuma:1
    volumes: [ "uptime-kuma:/app/data" ]
    ports: [ "3001:3001" ]
    restart: unless-stopped
volumes: { uptime-kuma: {} }
```

Open `http://<host>:3001`, create the admin account, then add monitors:

| Monitor | URL | Expect |
|---------|-----|--------|
| Core API readiness | `https://api.<domain>/healthz` | 200 |
| Admin API readiness | `https://admin-api.<domain>/healthz` | 200 |
| Dashboard | `https://dashboard.<domain>/` | 200 |

Notes:
- Watch `/healthz` (not `/health`): it pings the database, so a DB outage flips the monitor red.
- Interval 60s, retries 2. Add a notification channel (email / Telegram / webhook).
- Optionally add a keyword monitor on `/healthz` for `"status":"ok"`.
- Put Uptime Kuma behind the reverse proxy or restrict port 3001 by firewall — don't expose it open.
