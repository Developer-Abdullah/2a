#!/usr/bin/env bash
# Timestamped Postgres backup for Double A. Run on the server (cron-friendly).
#
#   bash scripts/backup-db.sh                 # -> ./backups/doublea-YYYYmmdd-HHMMSS.sql.gz
#   BACKUP_DIR=/mnt/backups bash scripts/backup-db.sh
#
# Cron (daily 03:30, keep 14 days):
#   30 3 * * * cd /opt/doublea && BACKUP_DIR=/mnt/backups bash scripts/backup-db.sh >> /var/log/doublea-backup.log 2>&1
set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-./backups}"
CONTAINER="${DB_CONTAINER:-platform-postgres}"
KEEP_DAYS="${KEEP_DAYS:-14}"

# Load DB_USER / DB_NAME from .env if present.
if [ -f .env ]; then
  set -a; . ./.env; set +a
fi
DB_USER="${DB_USER:-platform_admin}"
DB_NAME="${DB_NAME:-platform_db}"

mkdir -p "$BACKUP_DIR"
stamp="$(date -u +%Y%m%d-%H%M%S)"
out="$BACKUP_DIR/doublea-$stamp.sql.gz"

echo "Backing up $DB_NAME from $CONTAINER -> $out"
docker exec "$CONTAINER" pg_dump -U "$DB_USER" "$DB_NAME" | gzip > "$out"
echo "Done ($(du -h "$out" | cut -f1))."

# Prune old backups.
find "$BACKUP_DIR" -name 'doublea-*.sql.gz' -type f -mtime "+$KEEP_DAYS" -print -delete || true
