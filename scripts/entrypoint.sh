#!/bin/sh

echo "Running migrations..."

go run ./cmd/migrate up

echo "Running server..."

go run ./cmd/server/
