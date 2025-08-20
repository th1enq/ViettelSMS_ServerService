#!/bin/sh

echo "Running migrations..."

# Run migrations
./migrate up

exec ./main
