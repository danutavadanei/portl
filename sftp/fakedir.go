package sftp

import (
	"os"
	"time"
)

type fakeDir struct{}

func (f fakeDir) Name() string {
	return "/."
}

func (f fakeDir) Size() int64 {
	return 0
}

func (f fakeDir) Mode() os.FileMode {
	return os.ModeDir
}

func (f fakeDir) ModTime() time.Time {
	return time.Now()
}

func (f fakeDir) IsDir() bool {
	return true
}

func (f fakeDir) Sys() any {
	return nil
}
