#!/bin/sh
set -e

# default values
DB_HOST=${DB_HOST:-db}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-pr_review}

export PGPASSWORD="$DB_PASSWORD"

echo "Waiting for database at ${DB_HOST}:${DB_PORT}..."
until psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' >/dev/null 2>&1; do
  sleep 1
done

if [ -d /migrations ]; then
  for f in /migrations/*.sql; do
    [ -e "$f" ] || continue
    echo "Applying migration $f"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$f"
  done
fi

exec /usr/local/bin/pr-review
