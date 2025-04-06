package http

import (
	"net/http"

	"github.com/danutavadanei/portl/common"
	"go.uber.org/zap"
)

type Server struct {
	logger     *zap.Logger
	sm         *common.SessionManager
	listenAddr string
}

func NewServer(logger *zap.Logger, sm *common.SessionManager, listenAddr string) *Server {
	return &Server{
		logger:     logger,
		sm:         sm,
		listenAddr: listenAddr,
	}
}

func (s *Server) ListenAndServe() error {
	s.logger.Info("Starting HTTP server", zap.String("listen_addr", s.listenAddr))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /404", handle404(s.logger))
	mux.HandleFunc("POST /404", handle404(s.logger))
	mux.HandleFunc("GET /{id}", handleDownloadPage(s.logger, s.sm))
	mux.HandleFunc("POST /{id}", handleDownload(s.logger, s.sm))

	return http.ListenAndServe(s.listenAddr, mux)
}
