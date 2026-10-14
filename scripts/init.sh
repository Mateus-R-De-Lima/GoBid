#!/bin/bash

set -e

echo "Waiting for database to be ready..."

until pg_isready \
    -h "$GOBID_DATABASE_HOST" \
    -p "$GOBID_DATABASE_PORT" \
    -U "$GOBID_DATABASE_USER" \
    -d "$GOBID_DATABASE_NAME"; do
    sleep 1
done

echo "Creating database if doesn't exist..."

export PGPASSWORD="$GOBID_DATABASE_PASSWORD"

psql \
    -h "$GOBID_DATABASE_HOST" \
    -p "$GOBID_DATABASE_PORT" \
    -U "$GOBID_DATABASE_USER" \
    -d postgres \
    -tc "SELECT 1 FROM pg_database WHERE datname = '$GOBID_DATABASE_NAME'" \
    | grep -q 1 \
    || psql \
        -h "$GOBID_DATABASE_HOST" \
        -p "$GOBID_DATABASE_PORT" \
        -U "$GOBID_DATABASE_USER" \
        -d postgres \
        -c "CREATE DATABASE \"$GOBID_DATABASE_NAME\";"

unset PGPASSWORD

echo "Applying database migrations..."

cd /app/internal/store/pgstore/migrations
tern migrate

echo "Starting the application..."

cd /app

exec ./main