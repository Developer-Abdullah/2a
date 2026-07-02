# Database Schema & Isolation Strategy

## Isolation choice: schema-per-tenant

Each merchant (tenant) gets a dedicated Postgres schema (`tenant_<slug>`). Platform-wide data lives
in `public`. This was chosen over shared-schema + `tenant_id` row filtering because:

- **Hard isolation by default.** A query runs against one tenant's schema; there is no `WHERE
  tenant_id = ?` to forget. Cross-tenant leakage requires an explicit, obvious cross-schema join.
- **Per-tenant operations** (backup/restore, drop on offboarding, per-tenant migration) are simple.
- **Scale fits.** The target is < 100 merchants; the schema count stays well within Postgres limits,
  and the per-schema DDL cost at onboarding is negligible.

Trade-off: migrations must be applied to every tenant schema (handled by
`internal/provision.ProvisionTenantSchema`, which replays the embedded `migrations/tenant/*.up.sql`
in order on a pinned connection). See [multi-tenancy.md](multi-tenancy.md).

## Migration layout (golang-migrate)

- `migrations/public/NNN_*.up.sql` / `.down.sql` — platform schema (applied once).
- `migrations/tenant/NNN_*.up.sql` / `.down.sql` — the template applied into each tenant schema.

## Public schema (platform)

| Table | Purpose |
|-------|---------|
| `tenants` | Merchant record: slug, plan, isolation mode, `schema_name`, operational `status`, and subscription state (`subscription_status`, `subscription_expires_at`, `grace_until`). |
| `tenant_plans` | Plan tiers with feature limits **and pricing** (`price_amount`, `currency`, `billing_interval`). |
| `tenant_config` | Per-tenant display config (app name, support email). |
| `platform_admins` | Platform owner + tenant-bound admins (bcrypt password hash, role). |
| `tenant_certificates` | Encrypted signing assets per tenant (`s3_key_p12_enc`, `s3_key_mobileprovision`, `encrypted_password`). |
| `audit_logs` | Platform-level audit. |
| `subscription_payments` | Idempotent payment ledger (see [billing.md](billing.md)). |
| `subscription_events` | Renewal reminders + subscription state-change log. |

## Tenant schema (per merchant)

Key tables: `users`, `devices`, `activation_codes`, `applications`, `application_versions`
(with `signing_status` **and** `review_status`), `signing_jobs`, `installations`, `updates`,
`notifications`, `analytics_events`/`analytics_daily`, `clone_records`, `admin_users`, `audit_logs`
(tamper-proof trigger), `platform_ratings`, `udid_enrollment_sessions`.

Cross-schema foreign keys point from tenant tables to `public.tenant_certificates` /
`public.tenants` for signing and enrollment.
