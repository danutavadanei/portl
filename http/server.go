package http

import (
	"github.com/danutavadanei/portl/broker"
	"github.com/danutavadanei/portl/config"
	"net/http"

	"go.uber.org/zap"
)

type Server struct {
	logger     *zap.Logger
	store      *broker.Store
	listenAddr string
}

func NewServer(logger *zap.Logger, store *broker.Store, cfg *config.Config) *Server {
	return &Server{
		logger:     logger,
		store:      store,
		listenAddr: cfg.HttpListenAddr,
	}
}

func (s *Server) Serve() error {
	s.logger.Info("Starting HTTP server", zap.String("listen_addr", s.listenAddr))

	mux := http.NewServeMux()

	handler404 := s.handle404()

	mux.HandleFunc("GET /404", handler404)
	mux.HandleFunc("POST /404", handler404)
	mux.HandleFunc("GET /{id}", s.handleDownloadPage())
	mux.HandleFunc("POST /{id}", s.handleDownload)

	return http.ListenAndServe(s.listenAddr, mux)
}
