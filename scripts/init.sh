#!/bin/bash

echo "Waiting for database to be ready..."

until pg_isready -h db -U postgres -d gobid; do
    sleep 1
done


echo "Creating database if doesn't exist..."

psql -h db -U postgres -c "CREATE DATABASE gobid;" || true

echo "Applying database migrations..."

cd /app/internal/store/pgstore/migrations
tern migrate

echo "Starting the application..."

cd /app

exec ./main