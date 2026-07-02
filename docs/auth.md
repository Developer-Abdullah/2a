# Authentication & Authorization

## Actors

- **Platform owner** — a `platform_admins` row with `role = super_admin` and **no tenant binding**.
  Manages merchants and reviews/approves builds across all tenants.
- **Tenant admin** — a `platform_admins` row bound to one tenant (its slug travels in the JWT).
  Operates only within its own tenant.
- **End user** — a device holder. Does **not** log in with a password; identified by an activation
  code + device, then issued a JWT bound to the device.

## Tokens (JWT, ES256)

Signed with a P-256 ECDSA key (`internal/auth/jwt.go`). `CustomClaims`:

| Claim | Meaning |
|-------|---------|
| `user_id` | subject id |
| `role` | `super_admin`, `user`, `refresh`, … |
| `tenant_id` | tenant **slug** for a tenant admin; empty for the platform owner |
| `device_id` | bound device; `admin-dashboard` for dashboard sessions |

Access tokens last 24h; refresh 30d. Admin passwords are hashed with **bcrypt**
(`platform_repository.go`, `AuthenticateAdmin`).

## Middleware chain

1. `securityMiddleware` (per-router) — CORS, security headers, body-size cap, and a coarse per-IP
   rate limit (120/min).
2. `tenant.ResolverMiddleware` — resolves the tenant from `X-Tenant-Slug` / `X-Tenant-ID` / subdomain
   and rejects inactive tenants. (Data routes only.)
3. `auth.RequireAuth(pubKey)` — validates the JWT and enforces the `X-Device-ID` binding.
4. `enforceTenantScope` — a tenant admin may only act on its own tenant; the owner passes freely.
5. `auth.RequirePlatformOwner` — gates the `/v1/admin/platform/*` group to the owner (defense in
   depth alongside the per-handler `requireOwner`).

## Brute-force protection

On top of the coarse limiter, sensitive endpoints get a tight per-IP cap
(`internal/ratelimit`, 10/min): admin login (`POST /v1/admin/auth/login`) and activation-code
validation (`POST /v1/activation/validate`). The limiter is in-memory (per replica); swap to a
Redis counter when the limit must hold cluster-wide — the swap is localized to `internal/ratelimit`.
