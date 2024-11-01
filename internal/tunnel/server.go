package tunnel

import (
	"context"
	"log"
	"net"

	"diagnostic-client/internal/config"
)

type Server struct {
	cfg      *config.Config
	listener net.Listener
	handler  *Handler
}

func NewServer(cfg *config.Config, handler *Handler) (*Server, error) {
	listener, err := net.Listen("tcp", cfg.AgentAddr)
	if err != nil {
		return nil, err
	}

	return &Server{
		cfg:      cfg,
		listener: listener,
		handler:  handler,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	log.Printf("Tunnel server listening on %s", s.cfg.AgentAddr)

	go func() {
		<-ctx.Done()
		if s.listener != nil {
			s.listener.Close()
		}
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				log.Printf("Accept error: %v", err)
				continue
			}
		}

		log.Printf("New agent connected from: %s", conn.RemoteAddr())
		go s.handler.HandleConnection(ctx, conn)
	}
}
