# Multi-Tenancy & Isolation Guarantee

## Model

Schema-per-tenant. Every merchant has a Postgres schema `tenant_<slug>`; platform data is in
`public`. The tenant is resolved per request (subdomain or `X-Tenant-Slug`) into a `domain.Tenant`
that carries its `SchemaName`.

## How queries are scoped

Repositories build queries against the resolved schema by quoting it with
`postgres.QuoteIdentifier` and interpolating it into the table reference, e.g.:

```go
schema := postgres.QuoteIdentifier(tenant.SchemaName)
fmt.Sprintf(`SELECT ... FROM %s.applications WHERE ...`, schema)
```

`QuoteIdentifier` double-quotes and escapes the identifier; schema names are additionally
constrained at creation to `^[a-z0-9_]+$` (see `public.tenants` CHECK), so the interpolated
identifier cannot carry injection.

## The guarantee

- A data query targets exactly one tenant schema, chosen from the authenticated request context —
  not from user-supplied table names.
- `enforceTenantScope` prevents a tenant admin from resolving a tenant other than the one in its JWT.
- Cross-tenant reads exist only where explicitly intended and owner-gated: the platform review queue
  (`ListPendingReviews`) and tenant listing iterate schemas deliberately, behind
  `RequirePlatformOwner`.

## Onboarding / migrations

`internal/provision.ProvisionTenantSchema` creates the schema and replays every embedded
`migrations/tenant/*.up.sql` in lexical order on a single pinned connection (so `search_path` holds
for the whole run). Adding a tenant migration file is automatically picked up (`//go:embed
tenant/*.sql`). New migrations must be written to apply cleanly to existing populated schemas.

## Verifying isolation

The intended cross-tenant test: create tenants A and B, authenticate as B, and confirm B cannot read
A's rows (the resolver + `enforceTenantScope` reject it, and B's queries target only `tenant_b`).
This requires a live Postgres; run it against the docker-compose stack.
