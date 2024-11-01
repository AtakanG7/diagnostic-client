-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS btree_gin;

-- File system table
CREATE TABLE files (
    path TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('file', 'directory')),
    size BIGINT,
    mod_time TIMESTAMP WITH TIME ZONE,
    parent_path TEXT REFERENCES files(path),
    is_scraped BOOLEAN DEFAULT false,
    children_count INTEGER DEFAULT 0,
    has_children BOOLEAN DEFAULT false
);

CREATE INDEX idx_files_parent ON files(parent_path);
CREATE INDEX idx_files_type ON files(type) WHERE type = 'directory';

-- Log entries
CREATE TABLE logs (
    id BIGSERIAL PRIMARY KEY,
    file_path TEXT REFERENCES files(path),
    line TEXT NOT NULL,
    line_number INTEGER NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    level TEXT CHECK (level IN ('error', 'warning', 'info')),
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('english', line)) STORED
);

CREATE INDEX idx_logs_file_line ON logs(file_path, line_number);
CREATE INDEX idx_logs_timestamp ON logs(timestamp);
CREATE INDEX idx_logs_level ON logs(level);
CREATE INDEX idx_logs_search ON logs USING GIN(search_vector);

-- Network packets
CREATE TABLE network_packets (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    protocol TEXT NOT NULL,
    src_ip INET,
    dst_ip INET,
    src_port INTEGER,
    dst_port INTEGER,
    length INTEGER,
    payload_size INTEGER,
    tcp_flags TEXT
);

-- Convert to hypertable
SELECT create_hypertable('network_packets', 'time');

CREATE INDEX idx_network_protocol ON network_packets(protocol, time DESC);
CREATE INDEX idx_network_ips ON network_packets(src_ip, dst_ip);
