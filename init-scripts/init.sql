CREATE DATABASE diagnostic;
\c diagnostic

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS btree_gin;
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- File system table
CREATE TABLE files (
    path TEXT PRIMARY KEY,
    parent_path TEXT,
    name TEXT NOT NULL,
    is_directory BOOLEAN NOT NULL DEFAULT false,
    size BIGINT NOT NULL DEFAULT 0,
    mod_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_gzipped BOOLEAN NOT NULL DEFAULT false,
    is_scraped BOOLEAN NOT NULL DEFAULT false
);

-- Indexes for tree operations
CREATE INDEX idx_files_parent ON files(parent_path);
CREATE INDEX idx_files_directory ON files(is_directory) WHERE is_directory = true;
CREATE INDEX idx_files_scraped ON files(is_scraped) WHERE is_scraped = true;

-- Log entries
CREATE TABLE logs (
    id BIGSERIAL PRIMARY KEY,
    file_path TEXT REFERENCES files(path) ON DELETE CASCADE,
    line TEXT NOT NULL,
    line_number INTEGER NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    level TEXT DEFAULT 'info',
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('english', line)) STORED
);

CREATE INDEX idx_logs_file_line ON logs(file_path, line_number);
CREATE INDEX idx_logs_timestamp ON logs(timestamp);
CREATE INDEX idx_logs_level ON logs(level);
CREATE INDEX idx_logs_search ON logs USING GIN(search_vector);

-- Network packets
CREATE TABLE network_packets (
    time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    protocol TEXT NOT NULL,
    src_ip INET,
    dst_ip INET,
    src_port INTEGER,
    dst_port INTEGER,
    length INTEGER DEFAULT 0,
    payload_size INTEGER DEFAULT 0,
    tcp_flags TEXT
);

SELECT create_hypertable('network_packets', 'time', chunk_time_interval => INTERVAL '1 hour');

CREATE INDEX idx_network_protocol ON network_packets(protocol, time DESC);
CREATE INDEX idx_network_ips ON network_packets(src_ip, dst_ip);