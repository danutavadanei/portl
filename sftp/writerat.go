package sftp

import (
	"errors"
	"io"
)

type writerAt struct {
	offset int64
	pw     *io.PipeWriter
}

func (w *writerAt) WriteAt(p []byte, off int64) (n int, err error) {
	if off != w.offset {
		return 0, errors.New("non-sequential writes not supported in streaming mode")
	}
	w.offset += int64(len(p))
	return w.pw.Write(p)
}

func (w *writerAt) Close() error {
	return w.pw.Close()
}
