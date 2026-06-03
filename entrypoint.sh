#!/bin/sh
set -e

echo "Running goose migrations..."
/usr/local/bin/goose -dir /root/db/migrations postgres "host=${DB_HOST} port=${DB_PORT} user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_NAME} sslmode=disable" up

echo "Starting API server..."
exec /root/server