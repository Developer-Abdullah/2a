# Object Storage

S3-compatible storage (MinIO in dev; Backblaze B2 / Wasabi in prod) holds all binary artifacts.
The client (`pkg/storage/s3.go`) uses AWS SDK Go v2 with a custom endpoint and path-style addressing.

## Configuration (env)

`S3_ENDPOINT`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_REGION`, `S3_BUCKET`. An empty endpoint uses the
AWS default; a non-empty endpoint sets `BaseEndpoint` + `UsePathStyle=true` (required for MinIO/B2).

## Key layout (`pkg/storage/paths.go`)

All keys are prefixed by tenant id:

| Kind | Key pattern |
|------|-------------|
| Raw IPA | `{tenantID}/raw-ipa/{appID}/{versionID}.ipa` |
| Signed IPA | `{tenantID}/signed-ipa/{appID}/{versionID}-{certID}.ipa` |
| Manifest | `{tenantID}/manifests/{appID}/{versionID}.plist` |
| Certificate (encrypted) | `{tenantID}/certificates/{certID}/cert.p12.enc` |
| Export CSV | `{tenantID}/exports/{adminID}/{ts}-export.csv` |

## Upload flow (no proxying through the API)

The dashboard requests a **presigned PUT URL** (`POST /v1/admin/apps/:id/upload-url`, 1h lifetime)
and uploads the raw IPA straight to storage. It then creates the version row referencing the key
(`POST /v1/admin/apps/:id/versions`). This avoids streaming large IPAs through the API.

Because the DB row references an S3 key rather than being written in the same transaction as the
bytes, a failed upload leaves no signed artifact and the version simply never leaves
`pending`/`pending_review`; nothing orphaned is served.

## Downloads

OTA downloads are served via short-lived presigned GET URLs (15 min for the IPA, 1h for icons),
issued by the install endpoints — the bucket itself stays private. See
[ota-distribution.md](ota-distribution.md).

## Certificate assets at rest

`.p12` files are stored **AES-256-GCM encrypted** (`cert.p12.enc`) and the p12 password is stored as
GCM ciphertext in `tenant_certificates.encrypted_password`. The signing worker decrypts both with
`AES_ENCRYPTION_KEY` just-in-time inside its scratch dir. See [signing-pipeline.md](signing-pipeline.md).
