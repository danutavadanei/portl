package sftp

import (
	"errors"
	"github.com/danutavadanei/portl/broker"
	"github.com/pkg/sftp"
	"io"
	"log"
	"path"
)

type Handler struct {
	mb         broker.Broker
	pathPrefix string
}

func NewHandler(mb broker.Broker) *Handler {
	return &Handler{mb: mb}
}

func (sh *Handler) Filewrite(req *sftp.Request) (io.WriterAt, error) {
	log.Printf("Filewrite: Received command for path: %s (method=%s)", req.Filepath, req.Method)

	if req.Method != "Put" {
		return nil, nil
	}

	data := make(chan []byte)

	if err := sh.mb.Publish(broker.Message{
		Type: broker.Put,
		Path: sh.normalizePath(req.Filepath),
		Data: data,
	}); err != nil {
		return nil, err
	}

	return &writerAt{data: data}, nil
}

func (sh *Handler) Fileread(*sftp.Request) (io.ReaderAt, error) {
	return nil, nil
}

func (sh *Handler) Filelist(request *sftp.Request) (sftp.ListerAt, error) {
	switch request.Method {
	case "Stat":
		return listerat{fakeDir{}}, nil
	}

	return nil, errors.New("unsupported")
}

func (sh *Handler) Filecmd(req *sftp.Request) error {
	log.Printf("Filecmd: Received command for path: %s (method=%s)", req.Filepath, req.Method)

	if req.Method == "Mkdir" {
		p := sh.normalizePath(req.Filepath)
		if p == "" {
			return nil
		}

		if err := sh.mb.Publish(broker.Message{
			Type: broker.Mkdir,
			Path: p,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (sh *Handler) normalizePath(p string) string {
	return path.Clean(p)[1:]
}
