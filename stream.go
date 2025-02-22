package portl

import (
	"errors"
	"io"
	"sync"
)

type MsgType uint8

const (
	OpenFile  MsgType = 1
	WriteFile MsgType = 2
)

type Msg struct {
	MsgType MsgType
	Path    string
	Data    chan []byte
}

type Stream struct {
	Mu   sync.Mutex
	Msgs chan Msg
}

func NewStream() *Stream {
	return &Stream{
		Msgs: make(chan Msg),
	}
}

func (s *Stream) OpenFile(path string) {
	s.Msgs <- Msg{MsgType: OpenFile, Path: path}
}

func (s *Stream) WriteFile(path string) io.WriterAt {
	data := make(chan []byte)

	go func() {
		s.Msgs <- Msg{MsgType: WriteFile, Path: path, Data: data}
	}()

	return &writerAt{data: data}
}

func (s *Stream) Close() {
	close(s.Msgs)
}

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
