# IPA Signing Pipeline

Signing runs asynchronously **after owner approval** (see [review-workflow.md](review-workflow.md)),
driven by a Redis-backed queue (asynq).

## Flow

1. Owner approves a version → the review controller enqueues an `signing:ipa_sign` task
   (`{JobID, VersionID, CertID, TenantID}`) and the `signing_jobs` row moves toward processing.
2. The **signing worker** (`cmd/signing`, its own process/image) consumes the task
   (`internal/signing/worker.go`):
   1. Loads the signing context (raw IPA key + the tenant's active certificate) via
      `GetSigningContext`, which cross-joins `public.tenant_certificates`.
   2. Downloads the raw IPA and the `.mobileprovision` to a per-job scratch dir.
   3. Downloads the AES-256-GCM-encrypted `.p12`, **decrypts** it and the p12 password with
      `AES_ENCRYPTION_KEY` (`pkg/crypto`).
   4. Runs **zsign** (`pkg/zsign`): `zsign -k cert.p12 -p <password> -m app.mobileprovision -o
      signed.ipa original.ipa`. The command is executed as an argv slice (no shell), so
      tenant-controlled filenames cannot inject.
   5. Uploads the signed IPA to `signed-ipa/...` and marks the job `completed` + version `signed`.
   6. Any failure marks the job `failed` with the captured reason — **there is no silent
      success/copy fallback**. A version with no cert or no raw IPA fails explicitly.

## Certificate contract

`tenant_certificates` stores, per tenant: `s3_key_p12_enc` (GCM-encrypted p12), `s3_key_mobileprovision`
(plaintext profile), and `encrypted_password` (GCM ciphertext of the p12 password). All GCM material
uses `AES_ENCRYPTION_KEY` (32 bytes hex). The worker never persists decrypted material outside its
scratch dir, which is removed on completion.

## Container

`docker/signing.Dockerfile` builds zsign from source into the worker image, so `zsign` is on `PATH`.
Override with `ZSIGN_BINARY`; override scratch location with `SIGNING_WORK_DIR`.

## Tests

`internal/signing/worker_test.go` drives the worker with a mock signer + fake storage and asserts:
happy path (decrypt → sign → upload → completed), no-certificate → job failed (no sign), and signer
failure → job failed. It does not shell out to a real zsign.

## What still needs a real Apple cert

Producing an installable signed build requires a valid Apple Developer certificate + provisioning
profile whose embedded device UDIDs cover the target device. That can only be validated end-to-end
with real Apple credentials (flagged in the Phase 5 smoke test, [../docs/deployment.md](deployment.md)).
