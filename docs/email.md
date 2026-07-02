# Email delivery of activation codes

When an order is fulfilled (admin confirms a manual payment, or a payment webhook lands), the storefront
queues a code-delivery email. The **notifier worker** (`cmd/notifier`) consumes it and sends via SMTP.

Codes are always shown on the customer's order page, so email is optional — but it adds credibility and
gives the customer a permanent record.

## How it flows

1. `shop.Service` (in both core-api and admin-api) mints the codes, then calls `OrderEmailer.NotifyFulfilled`,
   which enqueues an `email:send` asynq task on the `notifications` queue.
2. `cmd/notifier` picks up the task in `handleEmail`. If `SMTP_HOST` is set it sends over SMTP;
   otherwise it just logs the delivery (so the pipeline works in dev without a mail server).
3. The email is only queued on a **fresh** fulfillment, so re-confirming an order never re-sends codes.

## Configure SMTP

Set these in `.env` (the `notifier-worker` service reads it via `env_file`):

| Var | Meaning |
|-----|---------|
| `SMTP_HOST` | SMTP server host. Leave blank to log-only. |
| `SMTP_PORT` | Default `587` (STARTTLS). Port 465 / implicit TLS is **not** supported. |
| `SMTP_USER` | Login user. |
| `SMTP_PASS` | Password or SMTP key. |
| `SMTP_FROM` | From address. Defaults to `SMTP_USER` when blank (Gmail requires this). |

### Gmail

1. Enable 2-Step Verification on the Google account.
2. Create an **App Password**: https://myaccount.google.com/apppasswords (use it as `SMTP_PASS`, not your login password).
3. Set:
   ```
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USER=you@gmail.com
   SMTP_PASS=<16-char app password>
   SMTP_FROM=you@gmail.com
   ```

### Brevo (Sendinblue)

```
SMTP_HOST=smtp-relay.brevo.com
SMTP_PORT=587
SMTP_USER=<your brevo login>
SMTP_PASS=<your SMTP key>
SMTP_FROM=orders@yourdomain.com   # a verified sender in Brevo
```

## Apply and test

```bash
docker compose up -d --build notifier-worker   # pick up new env
```

Then confirm an order from the admin dashboard and watch the worker logs:

```bash
docker compose logs -f notifier-worker
```

- With SMTP set: `Email delivered  to=... subject=...`
- Without SMTP: `Email delivery (SMTP not configured; logged only)`
