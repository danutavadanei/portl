package sftp

import "errors"

type writerAt struct {
	offset int64
	data   chan []byte
}

func (w *writerAt) WriteAt(p []byte, off int64) (n int, err error) {
	if off != w.offset {
		return 0, errors.New("non-sequential writes not supported in streaming mode")
	}
	w.offset += int64(len(p))
	w.data <- p
	return len(p), nil
}

func (w *writerAt) Close() error {
	close(w.data)
	return nil
}
