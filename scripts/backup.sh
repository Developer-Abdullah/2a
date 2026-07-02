#!/usr/bin/env bash
#
# Postgres backup: pg_dump (custom format) -> local retention (7 daily + 4 weekly) -> upload to the
# S3-compatible store under a SEPARATE prefix from the IPA vault. Intended to run from cron:
#
#   0 3 * * *  /opt/app/scripts/backup.sh >> /var/log/pg-backup.log 2>&1
#
# Required env (same values the app uses):
#   DB_DSN                       postgres://user:pass@host:5432/db?sslmode=disable
#   S3_ENDPOINT S3_ACCESS_KEY S3_SECRET_KEY S3_REGION
#   BACKUP_BUCKET                bucket for backups (MUST differ from the IPA bucket)
# Optional:
#   BACKUP_DIR                   local dir (default /var/backups/postgres)
#   BACKUP_PREFIX                key prefix in the bucket (default db-backups)
set -euo pipefail

: "${DB_DSN:?DB_DSN is required}"
: "${S3_ENDPOINT:?S3_ENDPOINT is required}"
: "${S3_ACCESS_KEY:?S3_ACCESS_KEY is required}"
: "${S3_SECRET_KEY:?S3_SECRET_KEY is required}"
: "${BACKUP_BUCKET:?BACKUP_BUCKET is required}"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/postgres}"
BACKUP_PREFIX="${BACKUP_PREFIX:-db-backups}"
S3_REGION="${S3_REGION:-us-east-1}"

mkdir -p "$BACKUP_DIR/daily" "$BACKUP_DIR/weekly"
stamp="$(date +%Y%m%d-%H%M%S)"
dow="$(date +%u)"   # 1=Mon .. 7=Sun
daily_file="$BACKUP_DIR/daily/db-$stamp.dump"

echo "[backup] dumping database -> $daily_file"
pg_dump --format=custom --no-owner --dbname="$DB_DSN" --file="$daily_file"

# Promote Sunday's dump to the weekly set.
if [ "$dow" = "7" ]; then
  cp "$daily_file" "$BACKUP_DIR/weekly/db-$stamp.dump"
fi

# Retention: keep 7 newest daily, 4 newest weekly.
prune() { # dir keep
  ls -1t "$1"/*.dump 2>/dev/null | tail -n "+$(( $2 + 1 ))" | xargs -r rm -f
}
prune "$BACKUP_DIR/daily" 7
prune "$BACKUP_DIR/weekly" 4

# Upload today's dump. Uses the AWS CLI if present (works with any S3-compatible endpoint), else mc.
key="$BACKUP_PREFIX/$(date +%Y/%m)/db-$stamp.dump"
echo "[backup] uploading -> s3://$BACKUP_BUCKET/$key"
if command -v aws >/dev/null 2>&1; then
  AWS_ACCESS_KEY_ID="$S3_ACCESS_KEY" AWS_SECRET_ACCESS_KEY="$S3_SECRET_KEY" AWS_DEFAULT_REGION="$S3_REGION" \
    aws --endpoint-url "$S3_ENDPOINT" s3 cp "$daily_file" "s3://$BACKUP_BUCKET/$key"
elif command -v mc >/dev/null 2>&1; then
  mc alias set backup "$S3_ENDPOINT" "$S3_ACCESS_KEY" "$S3_SECRET_KEY" >/dev/null
  mc cp "$daily_file" "backup/$BACKUP_BUCKET/$key"
else
  echo "[backup] WARNING: neither aws nor mc found; dump kept locally only" >&2
  exit 1
fi

echo "[backup] done"
# Restore: pg_restore --clean --if-exists --no-owner -d "$DB_DSN" <dump>
