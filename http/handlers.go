package http

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/danutavadanei/portl/broker"
	"github.com/danutavadanei/portl/common"
	"go.uber.org/zap"
)

func stream(logger *zap.Logger, sm *common.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		b, ok := sm.Load(id)
		if !ok {
			http.Error(w, "Session ID not found", http.StatusNotFound)
			return
		}

		msgs, err := b.Subscribe()
		if err != nil {
			http.Error(w, "Failed to subscribe to broker", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", id))

		zw := zip.NewWriter(w)
		defer zw.Close()

		for msg := range msgs {
			switch msg.Type {
			case broker.Mkdir:
				logger.Debug("Creating directory",
					zap.String("path", msg.Path),
				)
				header := &zip.FileHeader{
					Name:     msg.Path + "/",
					Method:   zip.Store,
					Modified: time.Now(),
				}
				if _, err := zw.CreateHeader(header); err != nil {
					logger.Error("failed to write zip header for mkdir",
						zap.String("path", msg.Path),
						zap.Error(err),
					)
					return
				}
			case broker.Put:
				logger.Debug("Writing file",
					zap.String("path", msg.Path),
				)
				header := &zip.FileHeader{
					Name:     msg.Path,
					Method:   zip.Store,
					Modified: time.Now(),
				}

				iw, err := zw.CreateHeader(header)
				if err != nil {
					logger.Error("failed to write tar header for put",
						zap.String("path", msg.Path),
						zap.Error(err),
					)
					return
				}

				if _, err := io.Copy(iw, msg.Data); err != nil {
					logger.Error("failed to write data for put",
						zap.String("path", msg.Path),
						zap.Error(err),
					)
					return
				}
			}
		}
	}
}
