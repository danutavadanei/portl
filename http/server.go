package http

import (
	"github.com/danutavadanei/portl/static"
	"html/template"
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
	s.logger.Info("Starting HTTP server",
		zap.String("listen_addr", s.listenAddr),
	)

	mux := http.NewServeMux()

	downloadPage, err := template.ParseFS(static.Templates, "download.html")
	if err != nil {
		return err
	}
	notFoundPage, err := template.ParseFS(static.Templates, "404.html")
	if err != nil {
		return err
	}

	mux.HandleFunc("GET /404", func(w http.ResponseWriter, r *http.Request) {
		if err := notFoundPage.Execute(w, nil); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("POST /404", func(w http.ResponseWriter, r *http.Request) {
		if err := notFoundPage.Execute(w, nil); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("GET /{id}", func(w http.ResponseWriter, _ *http.Request) {
		if err := downloadPage.Execute(w, nil); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("POST /{id}", stream(s.logger, s.sm))

	return http.ListenAndServe(s.listenAddr, mux)
}
