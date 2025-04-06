package http

import (
	"archive/zip"
	"fmt"
	"github.com/danutavadanei/portl/static"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/danutavadanei/portl/broker"
	"go.uber.org/zap"
)

func (s *Server) handle404() http.HandlerFunc {
	notFoundPage, err := template.ParseFS(static.Templates, "404.html")
	if err != nil {
		s.logger.Error("failed to parse 404 template", zap.Error(err))
		return nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := notFoundPage.Execute(w, nil); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	}
}

func (s *Server) handleDownloadPage() http.HandlerFunc {
	downloadPage, err := template.ParseFS(static.Templates, "download.html")
	if err != nil {
		s.logger.Error("failed to parse download template", zap.Error(err))
		return nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := s.store.Load(r.PathValue("id")); !ok {
			http.Redirect(w, r, "/404", http.StatusFound)
			return
		}

		if err := downloadPage.Execute(w, nil); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	}
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	b, ok := s.store.Load(id)
	if !ok {
		http.Redirect(w, r, "/404", http.StatusFound)
		return
	}

	msgs, err := b.Subscribe()
	if err != nil {
		s.logger.Error("failed to subscribe to broker", zap.Error(err))
		http.Redirect(w, r, "/404", http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", id))

	zw := zip.NewWriter(w)
	defer zw.Close()

	for msg := range msgs {
		switch msg.Type {
		case broker.Mkdir:
			s.logger.Debug("Creating directory", zap.String("path", msg.Path))

			header := &zip.FileHeader{
				Name:     msg.Path + "/",
				Method:   zip.Store,
				Modified: time.Now(),
			}
			if _, err := zw.CreateHeader(header); err != nil {
				s.logger.Error("failed to write zip header for mkdir",
					zap.String("path", msg.Path),
					zap.Error(err),
				)
				return
			}
		case broker.Put:
			s.logger.Debug("writing file", zap.String("path", msg.Path))

			header := &zip.FileHeader{
				Name:     msg.Path,
				Method:   zip.Store,
				Modified: time.Now(),
			}
			iw, err := zw.CreateHeader(header)
			if err != nil {
				s.logger.Error("failed to write tar header for put", zap.String("path", msg.Path), zap.Error(err))
				return
			}

			if _, err := io.Copy(iw, msg.Data); err != nil {
				s.logger.Error("failed to write data for put", zap.String("path", msg.Path), zap.Error(err))
				return
			}
		}
	}
}
