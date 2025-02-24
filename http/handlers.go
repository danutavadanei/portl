package http

import (
	"archive/zip"
	"fmt"
	"github.com/danutavadanei/portl/broker"
	"github.com/danutavadanei/portl/common"
	"log"
	"net/http"
	"time"
)

func stream(sm *common.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		b, ok := sm.Load(id)
		if !ok {
			http.Error(w, "Session ID not found", http.StatusNotFound)
			return
		}

		msgs, err := b.(broker.Broker).Subscribe()
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
				log.Printf("Creating directory: %s", msg.Path)
				header := &zip.FileHeader{
					Name:     msg.Path + "/",
					Method:   zip.Store,
					Modified: time.Now(),
				}
				if _, err := zw.CreateHeader(header); err != nil {
					log.Printf("failed to write zip header for mkdir %s: %s", msg.Path, err)
					return
				}
			case broker.Put:
				log.Printf("Writing file: %s", msg.Path)
				header := &zip.FileHeader{
					Name:     msg.Path,
					Method:   zip.Store,
					Modified: time.Now(),
				}

				iw, err := zw.CreateHeader(header)
				if err != nil {
					log.Printf("failed to write tar header for put %s: %s", msg.Path, err)
					return
				}

				for data := range msg.Data {
					if _, err := iw.Write(data); err != nil {
						log.Printf("failed to write data for put %s: %s", msg.Path, err)
						return
					}
				}
			}
		}
	}
}
