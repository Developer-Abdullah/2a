# UDID Collection & Enrollment

End users enroll their device so its UDID can be captured (and later covered by a provisioning
profile). The flow is tenant-scoped: the tenant is resolved from the subdomain / `X-Tenant-Slug` on
every `/v1/enroll*` route.

## Endpoints (public)

| Method & path | Purpose |
|---------------|---------|
| `GET /v1/enroll` | Issues a one-time session and returns a signed `.mobileconfig` (`application/x-apple-aspen-config`). |
| `POST /v1/enroll/callback?token=…` | Device posts its attributes back after installing the profile. |
| `GET /v1/enroll/status?token=…` | The enrollment web page polls for completion. |

## Session lifecycle (`udid_enrollment_sessions`, per tenant)

1. `GET /v1/enroll` creates a session: a random 64-char `one_time_token`, client IP, and a 15-minute
   `expires_at`. It builds an enrollment profile (`pkg/mobileconfig`) whose `PayloadContent.URL`
   points to the callback with that token, and signs it with the tenant's enrollment cert. If no
   usable cert is configured it falls back to an **unsigned** profile (dev convenience; iOS shows an
   "unsigned" warning but still installs).
2. The device POSTs to the callback. `parseCallback` accepts a **PKCS#7-signed plist** (Apple's
   profile-service response), a raw plist, or a simple form post (`udid`/`product`/`version`) for a
   web-form flow. The UDID is **hashed (SHA-256, normalized)** — the raw identifier is never stored.
3. `CompleteSession` marks the session `completed` with `udid_hash` + `device_type`, but only while
   the session is unexpired (`WHERE expires_at > NOW()`); an expired/unknown token yields 409.
4. Polling returns `{completed, device_type}`.

## Per-tenant limits

Device activation (turning an enrolled/known device into an active install) is governed by
`activation_codes` (`max_devices`, `current_device_count`, `max_uses`) — enforced in the activation
flow, configurable per code/plan, not hardcoded.

## Notes / hardening

- Full attestation (verifying Apple's PKCS#7 signature chain on the callback) is not performed; the
  payload is parsed for the UDID. Add chain verification before treating enrollment as proof of
  device authenticity.
- The enrollment cert provider (`service.DBEnrollmentCertProvider`) is currently a placeholder;
  wiring real per-tenant enrollment certs makes the served profile trusted (no unsigned warning).
