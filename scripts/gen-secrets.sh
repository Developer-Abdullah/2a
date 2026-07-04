#!/usr/bin/env bash
# Generate strong production secrets + an ES256 (P-256) JWT keypair for Double A.
# Prints env lines you paste into the server's .env. Requires `openssl`.
#
#   bash scripts/gen-secrets.sh
#
# Re-running generates NEW secrets — only do that on first setup or a deliberate rotation
# (rotating JWT keys invalidates all existing admin/device tokens).
set -euo pipefail

rand() { openssl rand -hex "$1"; }
b64() { openssl rand -base64 "$1" | tr -d '\n'; }

# ES256 P-256 keypair. PEMs are emitted as single-line with literal \n so they fit one env var.
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
openssl ecparam -name prime256v1 -genkey -noout -out "$tmp/priv.pem"
openssl ec -in "$tmp/priv.pem" -pubout -out "$tmp/pub.pem" 2>/dev/null
esc() { awk '{printf "%s\\n", $0}' "$1"; }

echo "# ── generated $(date -u +%FT%TZ) — paste into .env, then keep it secret ──"
echo "DB_PASSWORD=$(rand 24)"
echo "S3_SECRET_KEY=$(rand 24)"
echo "AES_ENCRYPTION_KEY=$(rand 32)"          # 64 hex chars = 32 bytes
echo "NEXTAUTH_SECRET=$(b64 32)"
echo "JWT_PRIVATE_KEY_PEM=\"$(esc "$tmp/priv.pem")\""
echo "JWT_PUBLIC_KEY_PEM=\"$(esc "$tmp/pub.pem")\""
echo
echo "# Reminder: also set DB_DSN's password to the DB_PASSWORD above."
