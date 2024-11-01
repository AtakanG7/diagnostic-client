package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"diagnostic-client/internal/config"
	"diagnostic-client/internal/db"
	"diagnostic-client/internal/tunnel"
	"diagnostic-client/internal/websocket"
)

type Server struct {
	cfg    *config.Config
	db     *db.DB
	tunnel *tunnel.Handler
	ws     *websocket.Handler
	http   *Handler
	server *http.Server
}

func NewServer(cfg *config.Config, db *db.DB) *Server {
	// Initialize components
	tunnelHandler := tunnel.NewHandler(cfg, db)
	wsHandler := websocket.NewHandler(cfg, tunnelHandler)
	httpHandler := NewHandler(db)

	// Create server with routing
	mux := http.NewServeMux()

	// WebSocket endpoint
	mux.HandleFunc("/ws", wsHandler.ServeWS)

	// REST endpoints
	mux.HandleFunc("/api/files", httpHandler.GetFiles)
	mux.HandleFunc("/api/logs", httpHandler.GetLogs)
	mux.HandleFunc("/api/logs/search", httpHandler.SearchLogs)
	mux.HandleFunc("/api/network/metrics", httpHandler.GetNetworkMetrics)

	// Create HTTP server with timeouts
	server := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		cfg:    cfg,
		db:     db,
		tunnel: tunnelHandler,
		ws:     wsHandler,
		http:   httpHandler,
		server: server,
	}
}

func (s *Server) Run(ctx context.Context) error {
	// Start tunnel server in background
	tunnelServer, err := tunnel.NewServer(s.cfg, s.tunnel)
	if err != nil {
		log.Printf("Tunnel server error: %v", err)
		return err
	}
	go func() {
		if err := tunnelServer.Run(ctx); err != nil {
			log.Printf("Tunnel server error: %v", err)
		}
	}()

	// Start HTTP server
	go func() {
		log.Printf("HTTP server listening on %s", s.cfg.ServerAddr)
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutting down servers...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Graceful shutdown
	return s.server.Shutdown(shutdownCtx)
}
