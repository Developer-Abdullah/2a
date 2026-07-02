#!/usr/bin/env bash
#
# End-to-end pipeline smoke test (roadmap Phase 5.1). Exercises:
#   owner login -> create tenant -> tenant login -> create app -> upload sample IPA
#   -> create version (pending_review) -> owner approve (enqueues signing) -> poll -> fetch manifest
#
# It does NOT need a Mac. It DOES need the docker-compose stack up (make setup) and a couple of tools:
#   curl, jq  (and zip OR python3 to build the sample IPA)
#
# Config via env (sensible dev defaults for the seeded stack):
#   CORE_API   (default http://localhost:8080)
#   ADMIN_API  (default http://localhost:8081)
#   OWNER_EMAIL / OWNER_PASSWORD   platform-owner creds (from the seed)
set -euo pipefail

CORE_API="${CORE_API:-http://localhost:8080}"
ADMIN_API="${ADMIN_API:-http://localhost:8081}"
OWNER_EMAIL="${OWNER_EMAIL:-admin@platform.com}"
OWNER_PASSWORD="${OWNER_PASSWORD:-Admin123!}"
SLUG="smoke$(date +%s)"

need() { command -v "$1" >/dev/null 2>&1 || { echo "missing required tool: $1" >&2; exit 1; }; }
need curl; need jq

say() { printf '\n=== %s ===\n' "$1"; }

# ---- sample IPA: a structurally valid zip containing Payload/Sample.app/ ----
sample_ipa() {
  local out="$1" work
  work="$(mktemp -d)"
  mkdir -p "$work/Payload/Sample.app"
  printf 'stub' > "$work/Payload/Sample.app/Sample"
  cat > "$work/Payload/Sample.app/Info.plist" <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>CFBundleIdentifier</key><string>com.smoke.sample</string>
  <key>CFBundleName</key><string>Sample</string>
  <key>CFBundleVersion</key><string>1.0</string>
</dict></plist>
PLIST
  if command -v zip >/dev/null 2>&1; then
    ( cd "$work" && zip -qr "$out" Payload )
  elif command -v python3 >/dev/null 2>&1; then
    python3 - "$work" "$out" <<'PY'
import sys, shutil, os
work, out = sys.argv[1], sys.argv[2]
base = out[:-4] if out.endswith('.zip') else out
shutil.make_archive(base, 'zip', work)
if base + '.zip' != out: os.replace(base + '.zip', out)
PY
  else
    echo "need zip or python3 to build the sample IPA" >&2; exit 1
  fi
  rm -rf "$work"
}

say "1. owner login"
OWNER_TOKEN=$(curl -fsS -X POST "$ADMIN_API/v1/admin/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$OWNER_EMAIL\",\"password\":\"$OWNER_PASSWORD\"}" | jq -r '.data.access_token')
[ -n "$OWNER_TOKEN" ] && [ "$OWNER_TOKEN" != null ] || { echo "owner login failed"; exit 1; }

say "2. create tenant '$SLUG'"
curl -fsS -X POST "$ADMIN_API/v1/admin/platform/tenants" \
  -H "Authorization: Bearer $OWNER_TOKEN" -H 'Content-Type: application/json' \
  -d "{\"slug\":\"$SLUG\",\"name\":\"Smoke $SLUG\",\"admin_email\":\"admin@$SLUG.local\",\"admin_password\":\"smoke-pass-123\"}" >/dev/null
echo "tenant created"

say "3. tenant admin login"
TENANT_TOKEN=$(curl -fsS -X POST "$ADMIN_API/v1/admin/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"admin@$SLUG.local\",\"password\":\"smoke-pass-123\"}" | jq -r '.data.access_token')
AUTH=(-H "Authorization: Bearer $TENANT_TOKEN" -H "X-Tenant-Slug: $SLUG")

say "4. create app"
APP_ID=$(curl -fsS -X POST "$ADMIN_API/v1/admin/apps" "${AUTH[@]}" -H 'Content-Type: application/json' \
  -d '{"bundle_identifier":"com.smoke.sample","name":"Sample","category":"other"}' | jq -r '.data.id')
echo "app_id=$APP_ID"

say "5. request upload URL + upload sample IPA"
UP=$(curl -fsS -X POST "$ADMIN_API/v1/admin/apps/$APP_ID/upload-url" "${AUTH[@]}")
PUT_URL=$(echo "$UP" | jq -r '.data.presigned_url'); KEY=$(echo "$UP" | jq -r '.data.expected_key')
IPA="$(mktemp -d)/sample.ipa"; sample_ipa "$IPA"
curl -fsS -X PUT --upload-file "$IPA" "$PUT_URL" >/dev/null
echo "uploaded -> $KEY"

say "6. create version (should land in pending_review)"
VER=$(curl -fsS -X POST "$ADMIN_API/v1/admin/apps/$APP_ID/versions" "${AUTH[@]}" -H 'Content-Type: application/json' \
  -d "{\"version\":\"1.0\",\"build_number\":\"1\",\"raw_ipa_s3_key\":\"$KEY\",\"size_bytes\":$(wc -c < "$IPA")}")
VERSION_ID=$(echo "$VER" | jq -r '.data.version_id')
echo "version_id=$VERSION_ID review_status=$(echo "$VER" | jq -r '.data.review_status')"

say "7. owner approves (enqueues signing)"
curl -fsS -X POST "$ADMIN_API/v1/admin/platform/reviews/$SLUG/$VERSION_ID/approve" \
  -H "Authorization: Bearer $OWNER_TOKEN" >/dev/null
echo "approved"

say "8. poll signing status"
STATUS=""
for _ in $(seq 1 20); do
  sleep 2
  STATUS=$(curl -fsS "$ADMIN_API/v1/admin/apps/$APP_ID/versions" "${AUTH[@]}" \
    | jq -r --arg id "$VERSION_ID" '.data.items[] | select(.id==$id) | .signing_status')
  echo "  signing_status=$STATUS"
  [ "$STATUS" = "signed" ] && break
  [ "$STATUS" = "failed" ] && break
done

say "9. manifest + itms-services link"
MANIFEST_URL="$CORE_API/v1/install/$VERSION_ID/manifest.plist"
ITMS="itms-services://?action=download-manifest&url=$(printf '%s' "$MANIFEST_URL" | sed 's#/#%2F#g; s#:#%3A#g; s#?#%3F#g; s#=#%3D#g; s#&#%26#g')"
echo "manifest: $MANIFEST_URL"
echo "itms:     $ITMS"

cat <<NOTE

--- result ---
Pipeline exercised: login -> tenant -> app -> upload -> pending_review -> approve -> signing dispatch -> manifest.

Final signing_status: ${STATUS:-unknown}
NOTE

if [ "$STATUS" != "signed" ]; then
  cat <<NOTE
NOTE: signing_status is not 'signed'. This is EXPECTED without a real Apple signing identity.
The seeded tenant certificate is a placeholder — the strict signing worker fails rather than
faking a signed build. To validate real signing + install you MUST provide, per tenant:
  - a real Apple Developer .p12 (AES-256-GCM encrypted -> S3 cert.p12.enc)
  - the matching .mobileprovision whose embedded UDIDs cover your test device
Then re-run. On a real iPhone, open the itms-services link in SAFARI (Chrome cannot trigger the
install prompt); the manifest must be served over HTTPS with an XML content type.
NOTE
fi
