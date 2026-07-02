# Billing & Subscriptions

Subscription billing runs behind a provider-agnostic interface (`internal/billing`), with adapters
for **MyFatoorah** (Kuwait/GCC, KWD) and **Paymob** (Egypt). Providers are enabled by presence of
their credentials (`billing.FromEnv`).

> Provider specifics (endpoints, integration/iframe ids, and the exact webhook signature field
> ordering) are implemented against each provider's documented contract but **must be verified
> against a live sandbox account** before go-live. The durable parts — idempotency, subscription
> state, grace logic, signature-verify-before-trust — are provider-independent and tested.

## Plans & pricing

`public.tenant_plans` carries `price_amount NUMERIC(12,3)` (KWD is 3-decimal), `currency`, and
`billing_interval` (`monthly`/`yearly`). Price is data, never hardcoded (seed sets 5.000 KWD/year).

## Checkout

`POST /v1/admin/billing/checkout {provider, plan_id, callback_url?}` (tenant admin) →
`{redirect_url}`. The service resolves the plan price, calls the provider's hosted checkout, records
a **pending** `subscription_payments` row keyed by `(provider, provider_ref)`, and returns the
redirect. `GET /v1/admin/billing/subscription` returns `{status, expires_at, grace_until,
distribution_open}` for the billing page.

## Webhook contract

`POST /billing/webhook/:provider` (public, unauthenticated, **outside** `/v1` so no tenant resolver).

- The provider's adapter **verifies the signature first** (Paymob: HMAC-SHA512 over the documented
  ordered field concat; MyFatoorah: HMAC-SHA256/base64 over the documented fields). Invalid/missing
  signature → 401, `ErrInvalidSignature` — never applied, never retried.
- On a verified successful payment, `MarkPaidAndExtend` runs in one transaction: it locks the
  payment row, **no-ops if already `paid`** (idempotent under redelivery), else marks it paid and
  extends the tenant subscription.

Idempotency anchor: the `(provider, provider_ref)` payment row and its `UNIQUE idempotency_key`. A
webhook delivered twice transitions the row once; the second delivery affects zero rows.

## Subscription state machine

`tenants.subscription_status`:

```
trialing ──(payment)──▶ active ──(expiry, within grace)──▶ past_due ──(grace elapsed)──▶ expired
   ▲                       │                                                                │
   └───────────────────────┴──────────────────(payment)───────────────────────────────────┘
```

- `subscription_expires_at` = `max(now, current expiry) + interval` on payment (early renewals stack).
- `grace_until` = new expiry + `GRACE_PERIOD_DAYS` (default **14**).

## Grace policy

Existing signed builds keep distributing regardless of subscription state. New submissions
(upload-url + create-version) are gated by `billingGuard`: allowed while trialing, active, or inside
the grace window; blocked (HTTP 402) once past grace. Decision logic: `snapshotAllows` (unit-tested).

## Renewal reminders & sweep

The notifier worker runs an hourly pass (`billing.Service.SendRenewalReminders` +
`SweepStatuses`): it records a `reminder` event for tenants expiring within 7 days (deduped per
expiry) and advances lapsed statuses. Reminders are event rows for now (no email channel yet).
