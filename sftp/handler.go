package sftp

import (
	"errors"
	"io"
	"path"

	"github.com/danutavadanei/portl/broker"
	"github.com/pkg/sftp"
	"go.uber.org/zap"
)

type Handler struct {
	logger     *zap.Logger
	mb         broker.Broker
	pathPrefix string
}

func NewHandler(logger *zap.Logger, mb broker.Broker) *Handler {
	return &Handler{
		logger: logger,
		mb:     mb,
	}
}

func (sh *Handler) Filewrite(req *sftp.Request) (io.WriterAt, error) {
	sh.logger.Debug("Filewrite: Received command for path",
		zap.String("path", req.Filepath),
		zap.String("method", req.Method),
	)

	if req.Method != "Put" {
		return nil, nil
	}

	pr, pw := io.Pipe()

	if err := sh.mb.Publish(broker.Message{
		Type: broker.Put,
		Path: sh.normalizePath(req.Filepath),
		Data: pr,
	}); err != nil {
		return nil, err
	}

	return newWriterAt(pw), nil
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
	sh.logger.Debug("Filecmd: Received command for path",
		zap.String("path", req.Filepath),
		zap.String("method", req.Method),
	)

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
