package sftp

import (
	"io"
	"sync"
)

type writerAt struct {
	offset int64
	pw     io.WriteCloser

	mu sync.Mutex

	outOfOrderBytes map[int64][]byte
}

func newWriterAt(pw io.WriteCloser) *writerAt {
	return &writerAt{
		pw:              pw,
		outOfOrderBytes: make(map[int64][]byte),
	}
}

// WriteAt buffers out-of-order writes and writes them to w.pw in
// ascending order of offsets once the correct offset is reached.
func (w *writerAt) WriteAt(p []byte, off int64) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Store the incoming chunk (possibly out-of-order) in the map
	w.outOfOrderBytes[off] = p

	// Try to flush data from the current offset forward
	for {
		chunk, ok := w.outOfOrderBytes[w.offset]
		if !ok {
			break // no chunk is available at the current offset yet
		}

		// Write the chunk at w.offset
		written, writeErr := w.pw.Write(chunk)
		if writeErr != nil {
			return 0, writeErr
		}
		if written < len(chunk) {
			return 0, io.ErrShortWrite
		}

		// Advance the offset and remove the written chunk from the map
		w.offset += int64(len(chunk))
		delete(w.outOfOrderBytes, w.offset-int64(len(chunk)))
	}

	// Return the length of the chunk just received (p).
	// Even if it hasn't been flushed yet (because its offset was too high),
	// from the caller's perspective, we've "accepted" it.
	return len(p), nil
}

func (w *writerAt) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// If there are any out-of-order bytes left, return an error
	if len(w.outOfOrderBytes) > 0 {
		return io.ErrUnexpectedEOF
	}

	return w.pw.Close()
}
