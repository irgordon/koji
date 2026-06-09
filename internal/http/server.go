package http

import (
	"context"
	"database/sql"
	"koji/internal/config"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func New(cfg config.Config, database *sql.DB) (*Server, error) {
	mux, err := NewMuxRouter(cfg, database)
	if err != nil {
		return nil, err
	}

	httpServer := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{httpServer: httpServer}, nil
}

func (s *Server) Addr() string {
	return s.httpServer.Addr
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
