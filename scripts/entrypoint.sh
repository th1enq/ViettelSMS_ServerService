#!/bin/sh

# Function to wait for database
wait_for_db() {
    echo "Waiting for database to be ready..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if nc -z postgres 5432; then
            echo "Database is ready!"
            return 0
        fi
        echo "Database not ready, waiting... (attempt $attempt/$max_attempts)"
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo "Database failed to become ready after $max_attempts attempts"
    exit 1
}

# Wait for database to be ready
wait_for_db

echo "Running migrations..."

# Run migrations
./migrate up

# Check if migrations were successful
if [ $? -eq 0 ]; then
    echo "Migrations completed successfully"
else
    echo "Migrations failed"
    exit 1
fi

# Start the main application
echo "Starting the main application..."
exec ./main
