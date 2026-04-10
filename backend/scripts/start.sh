#!/bin/sh
set -eu

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is required"
  exit 1
fi

echo "Waiting for Postgres..."
until pg_isready -d "$DATABASE_URL" >/dev/null 2>&1; do
  sleep 1
done

echo "Running migrations..."
/app/migrate

echo "Running seed..."
/app/seed

echo "Starting API server..."
exec /app/api
