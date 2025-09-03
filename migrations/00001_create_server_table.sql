-- +goose Up
CREATE TYPE server_status AS ENUM ('UNKNOWN', 'ONLINE', 'OFFLINE');

CREATE TABLE servers (
    server_id VARCHAR(32) PRIMARY KEY,
    server_name VARCHAR(64) UNIQUE NOT NULL,
    ipv4 VARCHAR(15) NOT NULL UNIQUE,
    status server_status NOT NULL DEFAULT 'UNKNOWN',
    location VARCHAR(128),
    os VARCHAR(32),
    interval_time INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_servers_name ON servers (server_name);

-- +goose Down
DROP TABLE IF EXISTS servers;
DROP TYPE IF EXISTS server_status;