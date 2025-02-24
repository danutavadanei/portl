package http

import (
	"github.com/danutavadanei/portl/common"
	"net/http"
)

type Server struct {
	sm         *common.SessionManager
	listenAddr string
}

func NewServer(sm *common.SessionManager, listenAddr string) *Server {
	return &Server{
		sm:         sm,
		listenAddr: listenAddr,
	}
}

func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/{id}", stream(s.sm))

	return http.ListenAndServe(s.listenAddr, mux)
}
