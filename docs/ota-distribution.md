# OTA Distribution (itms-services)

iOS installs enterprise/ad-hoc apps over the air via the `itms-services://` protocol, which points
Safari at an HTTPS-served `manifest.plist`.

## Endpoints (`internal/api/controllers/install_controller.go`)

| Method & path | Auth | Purpose |
|---------------|------|---------|
| `GET /v1/install/:version_id/manifest.plist` | public | Serves the Apple manifest XML. |
| `GET /v1/install/:version_id/download` | public | 302 → 15-min presigned URL for the signed IPA. |
| `GET /v1/apps/:id/install-info` | user JWT | Returns the manifest URL for the client to wrap in `itms-services://`. |

The install/download endpoints are intentionally unauthenticated: iOS fetches them directly with no
app-controlled headers. The IPA bytes stay private behind the short-lived presigned redirect.

## Manifest generation (`internal/signing/manifest.go`)

`GenerateManifestPlist` renders the exact Apple DTD structure (`items` → `assets`
[`software-package`, `display-image`, `full-size-image`] + `metadata` with `bundle-identifier`,
`bundle-version`, `kind=software`, `title`). It uses `text/template` (not `html/template`) so the
strict XML is not entity-escaped.

**Content type:** the manifest is served as XML. Apple is strict here — serve over HTTPS with an
XML content type and do not alter the node structure.

**Download URL:** the manifest's `software-package` URL must point at the API's stable
`/v1/install/:version_id/download` redirect, never at a raw (expiring) presigned URL, so the manifest
stays valid.

## Device gating

Restricting installs to activated UDIDs is ultimately enforced at the **provisioning-profile** level
during signing (the profile embeds the permitted device UDIDs) — not merely at the link level. A
link is not a security boundary; the signed profile is.
