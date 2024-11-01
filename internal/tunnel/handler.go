package tunnel

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"diagnostic-client/internal/config"
	"diagnostic-client/internal/db"
	"diagnostic-client/pkg/models"
)

type MessageType string

const (
	TypeMetrics   MessageType = "metrics"
	TypeLogList   MessageType = "log_list"
	TypeLogData   MessageType = "log_data"
	TypeLogSearch MessageType = "log_search"
)

type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Handler struct {
	cfg             *config.Config
	db              *db.DB
	networkStreamCh chan []models.NetworkPacket
	logStreamCh     chan models.LogEntry
	fileUpdateCh    chan models.FileNode
	batchMutex      sync.Mutex
	networkBatch    []models.NetworkPacket
	lastBatchTime   time.Time
}

func NewHandler(cfg *config.Config, db *db.DB) *Handler {
	return &Handler{
		cfg:             cfg,
		db:              db,
		networkStreamCh: make(chan []models.NetworkPacket, cfg.NetworkBufferSize),
		logStreamCh:     make(chan models.LogEntry, cfg.LogBufferSize),
		fileUpdateCh:    make(chan models.FileNode, 1000),
		networkBatch:    make([]models.NetworkPacket, 0, cfg.BatchSize),
		lastBatchTime:   time.Now(),
	}
}

func (h *Handler) HandleConnection(ctx context.Context, conn net.Conn) {
	log.Printf("[TUNNEL] New connection from: %s", conn.RemoteAddr())
	defer conn.Close()

	decoder := json.NewDecoder(conn)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			var msg Message
			if err := decoder.Decode(&msg); err != nil {
				log.Printf("[TUNNEL] Error decoding message: %v", err)
				return
			}

			// Log raw message for debugging
			rawMsg, _ := json.Marshal(msg)
			log.Printf("[TUNNEL] Received message: Type=%s, Raw=%s", msg.Type, string(rawMsg))

			if err := h.processMessage(ctx, msg); err != nil {
				log.Printf("[TUNNEL] Error processing message: %v", err)
			}
		}
	}
}

func (h *Handler) processMessage(ctx context.Context, msg Message) error {
	switch msg.Type {
	case TypeMetrics:
		var metrics struct {
			Timestamp string                 `json:"timestamp"`
			Packets   []models.NetworkPacket `json:"packets"`
		}
		if err := json.Unmarshal(msg.Payload, &metrics); err != nil {
			return fmt.Errorf("unmarshal metrics: %w", err)
		}
		log.Printf("[TUNNEL] Received metrics: timestamp=%s, packet_count=%d",
			metrics.Timestamp, len(metrics.Packets))

		h.batchMutex.Lock()
		h.networkBatch = append(h.networkBatch, metrics.Packets...)
		h.batchMutex.Unlock()

	case TypeLogList:
		var files []models.FileNode
		if err := json.Unmarshal(msg.Payload, &files); err != nil {
			return fmt.Errorf("unmarshal file list: %w", err)
		}
		log.Printf("[TUNNEL] Received file list: count=%d", len(files))
		for _, f := range files {
			log.Printf("[TUNNEL] File: path=%s type=%s", f.Path, f.Type)
		}

		if err := h.db.SaveFiles(ctx, files); err != nil {
			return fmt.Errorf("save files: %w", err)
		}

	case TypeLogData:
		var logs []models.LogEntry
		if err := json.Unmarshal(msg.Payload, &logs); err != nil {
			return fmt.Errorf("unmarshal logs: %w", err)
		}
		log.Printf("[TUNNEL] Received log entries: count=%d", len(logs))
		if len(logs) > 0 {
			log.Printf("[TUNNEL] Sample log: file=%s line_num=%d",
				logs[0].Filename, logs[0].LineNum)
		}

		if err := h.db.SaveLogs(ctx, logs); err != nil {
			return fmt.Errorf("save logs: %w", err)
		}

		for _, entry := range logs {
			select {
			case h.logStreamCh <- entry:
			default:
				log.Printf("[TUNNEL] Log channel full, dropping entry")
			}
		}
	}

	return nil
}

func (h *Handler) NetworkStream() <-chan []models.NetworkPacket {
	return h.networkStreamCh
}

func (h *Handler) LogStream() <-chan models.LogEntry {
	return h.logStreamCh
}

func (h *Handler) FileUpdates() <-chan models.FileNode {
	return h.fileUpdateCh
}
