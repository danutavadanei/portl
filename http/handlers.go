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
	"github.com/danutavadanei/portl/common"
	"go.uber.org/zap"
)

func handle404(l *zap.Logger) http.HandlerFunc {
	notFoundPage, err := template.ParseFS(static.Templates, "404.html")
	if err != nil {
		l.Error("failed to parse 404 template", zap.Error(err))
		return nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := notFoundPage.Execute(w, nil); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	}
}

func handleDownloadPage(l *zap.Logger, sm *common.SessionManager) http.HandlerFunc {
	downloadPage, err := template.ParseFS(static.Templates, "download.html")
	if err != nil {
		l.Error("failed to parse download template", zap.Error(err))
		return nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := sm.Load(r.PathValue("id")); !ok {
			http.Redirect(w, r, "/404", http.StatusFound)
			return
		}

		if err := downloadPage.Execute(w, nil); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	}
}

func handleDownload(l *zap.Logger, sm *common.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		b, ok := sm.Load(id)
		if !ok {
			http.Redirect(w, r, "/404", http.StatusFound)
			return
		}

		msgs, err := b.Subscribe()
		if err != nil {
			l.Error("failed to subscribe to broker", zap.Error(err))
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
				l.Debug("Creating directory", zap.String("path", msg.Path))

				header := &zip.FileHeader{
					Name:     msg.Path + "/",
					Method:   zip.Store,
					Modified: time.Now(),
				}
				if _, err := zw.CreateHeader(header); err != nil {
					l.Error("failed to write zip header for mkdir",
						zap.String("path", msg.Path),
						zap.Error(err),
					)
					return
				}
			case broker.Put:
				l.Debug("writing file", zap.String("path", msg.Path))

				header := &zip.FileHeader{
					Name:     msg.Path,
					Method:   zip.Store,
					Modified: time.Now(),
				}
				iw, err := zw.CreateHeader(header)
				if err != nil {
					l.Error("failed to write tar header for put", zap.String("path", msg.Path), zap.Error(err))
					return
				}

				if _, err := io.Copy(iw, msg.Data); err != nil {
					l.Error("failed to write data for put", zap.String("path", msg.Path), zap.Error(err))
					return
				}
			}
		}
	}
}
