package sftp

import (
	"io"
	"os"
)

type listerat struct{}

func (f listerat) ListAt(_ []os.FileInfo, _ int64) (int, error) {
	return 0, io.EOF
}
