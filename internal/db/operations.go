package db

import (
	"context"
	"fmt"
	"time"

	"diagnostic-client/pkg/models"

	"github.com/jackc/pgx/v5"
)

func (db *DB) SaveFiles(ctx context.Context, files []models.FileNode) error {
	batch := &pgx.Batch{}

	for _, file := range files {
		batch.Queue(`
            INSERT INTO files (path, name, type, size, mod_time, parent_path)
            VALUES ($1, $2, $3, $4, $5, $6)
            ON CONFLICT (path) DO UPDATE SET
                size = EXCLUDED.size,
                mod_time = EXCLUDED.mod_time
            RETURNING path`,
			[]interface{}{file.Path, file.Name, file.Type, file.Size, file.ModTime, file.ParentPath}...)
	}

	br := db.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range files {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("error saving file: %w", err)
		}
	}

	return nil
}

func (db *DB) GetFileTree(ctx context.Context, path string) ([]models.FileNode, error) {
	rows, err := db.pool.Query(ctx, `
        WITH RECURSIVE file_tree AS (
            SELECT path, name, type, size, mod_time, parent_path, is_scraped, 1 as level
            FROM files
            WHERE parent_path = $1 OR (parent_path IS NULL AND $1 = '/')
            
            UNION ALL
            
            SELECT f.path, f.name, f.type, f.size, f.mod_time, f.parent_path, f.is_scraped, ft.level + 1
            FROM files f
            INNER JOIN file_tree ft ON f.parent_path = ft.path
            WHERE ft.type = 'directory' AND ft.level < 2
        )
        SELECT 
            path, 
            name, 
            type, 
            size, 
            mod_time, 
            is_scraped,
            EXISTS(SELECT 1 FROM files WHERE parent_path = file_tree.path) as has_children,
            (SELECT COUNT(*) FROM files WHERE parent_path = file_tree.path) as child_count
        FROM file_tree
        ORDER BY path`, path)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.FileNode
	for rows.Next() {
		var f models.FileNode
		if err := rows.Scan(
			&f.Path, &f.Name, &f.Type, &f.Size, &f.ModTime, &f.IsScraped,
			&f.HasChildren, &f.ChildCount,
		); err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	return files, nil
}

func (db *DB) SaveLogs(ctx context.Context, logs []models.LogEntry) error {
	batch := &pgx.Batch{}

	for _, log := range logs {
		batch.Queue(`
            INSERT INTO logs (file_path, line, line_number, timestamp, level)
            VALUES ($1, $2, $3, $4, $5)`,
			[]interface{}{log.Filename, log.Line, log.LineNum, log.Timestamp, log.Level}...)
	}

	br := db.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range logs {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("error saving log: %w", err)
		}
	}

	return nil
}

func (db *DB) GetLogs(ctx context.Context, filePath string, beforeTime time.Time, limit int) ([]models.LogEntry, error) {
	rows, err := db.pool.Query(ctx, `
        SELECT file_path, line, line_number, timestamp, level
        FROM logs
        WHERE file_path = $1 AND timestamp < $2
        ORDER BY timestamp DESC, line_number DESC
        LIMIT $3`,
		filePath, beforeTime, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.LogEntry
	for rows.Next() {
		var l models.LogEntry
		if err := rows.Scan(
			&l.Filename, &l.Line, &l.LineNum, &l.Timestamp, &l.Level,
		); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	return logs, nil
}

func (db *DB) SearchLogs(ctx context.Context, query string, files []string, startTime, endTime time.Time) ([]models.LogEntry, error) {
	rows, err := db.pool.Query(ctx, `
        SELECT file_path, line, line_number, timestamp, level
        FROM logs
        WHERE 
            timestamp BETWEEN $1 AND $2
            AND ($3::text[] IS NULL OR file_path = ANY($3))
            AND search_vector @@ plainto_tsquery('english', $4)
        ORDER BY timestamp DESC
        LIMIT 1000`,
		startTime, endTime, files, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.LogEntry
	for rows.Next() {
		var l models.LogEntry
		if err := rows.Scan(
			&l.Filename, &l.Line, &l.LineNum, &l.Timestamp, &l.Level,
		); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	return logs, nil
}

func (db *DB) SaveNetworkPackets(ctx context.Context, packets []models.NetworkPacket) error {
	batch := &pgx.Batch{}

	for _, p := range packets {
		batch.Queue(`
            INSERT INTO network_packets (time, protocol, src_ip, dst_ip, src_port, dst_port, length, payload_size, tcp_flags)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			[]interface{}{p.Timestamp, p.Protocol, p.SrcIP, p.DstIP, p.SrcPort, p.DstPort, p.Length, p.PayloadSize, p.TCPFlags}...)
	}

	br := db.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range packets {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("error saving network packet: %w", err)
		}
	}

	return nil
}

func (db *DB) GetNetworkPackets(ctx context.Context, startTime, endTime time.Time, protocols []string) ([]models.NetworkPacket, error) {
	rows, err := db.pool.Query(ctx, `
        SELECT time, protocol, src_ip, dst_ip, src_port, dst_port, length, payload_size, tcp_flags
        FROM network_packets
        WHERE 
            time BETWEEN $1 AND $2
            AND ($3::text[] IS NULL OR protocol = ANY($3))
        ORDER BY time DESC
        LIMIT 1000`,
		startTime, endTime, protocols)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packets []models.NetworkPacket
	for rows.Next() {
		var p models.NetworkPacket
		if err := rows.Scan(
			&p.Timestamp, &p.Protocol, &p.SrcIP, &p.DstIP, &p.SrcPort, &p.DstPort,
			&p.Length, &p.PayloadSize, &p.TCPFlags,
		); err != nil {
			return nil, err
		}
		packets = append(packets, p)
	}

	return packets, nil
}
