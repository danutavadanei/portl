package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
)

// to read: /Users/danut/go/pkg/mod/github.com/pkg/sftp@v1.13.7/request-example.go

func main() {
	ps := &pipes{m: make(map[string]*streamFile)}

	go func() {
		if err := runSshServer(ps); err != nil {
			log.Fatalf("Failed to run SSH server: %v", err)
		}
	}()

	go func() {
		if err := runHttpServer(ps); err != nil {
			log.Fatalf("Failed to run HTTP server: %v", err)
		}
	}()

	select {}
}

func runHttpServer(ps *pipes) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/{id}/", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		ps.mu.Lock()
		sf, ok := ps.m[id]
		ps.mu.Unlock()
		if !ok {
			http.Error(w, "Session ID not found", http.StatusNotFound)
			return
		}

		sf.ReadHTTP(w)
	})

	return http.ListenAndServe("127.0.0.1:8090", mux)
}

// streamFile is the "bridge" between SFTP writes and HTTP reads.
type streamFile struct {
	mu            sync.Mutex
	currentOffset int64
	dataChan      chan []byte   // SFTP side pushes chunks here
	doneChan      chan struct{} // signals end-of-upload or abort
	closed        bool
	fileName      string
}

// newStreamFile creates the pipeline with unbuffered channels
func newStreamFile() *streamFile {
	return &streamFile{
		dataChan: make(chan []byte),
		doneChan: make(chan struct{}),
	}
}

// WriteAt is called by the SFTP server code whenever the client sends file data.
// We assume sequential writes: offset must match currentOffset, or we return error.
func (sf *streamFile) WriteAt(p []byte, off int64) (n int, err error) {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	if sf.closed {
		return 0, errors.New("write on closed streamFile")
	}
	if off != sf.currentOffset {
		return 0, errors.New("non-sequential writes not supported in streaming mode")
	}
	sf.currentOffset += int64(len(p))

	// Send chunk to dataChan
	// If no HTTP reader is connected yet, this blocks until it starts reading.
	chunkCopy := make([]byte, len(p))
	copy(chunkCopy, p) // to avoid reusing p
	sf.dataChan <- chunkCopy
	return len(p), nil
}

// Close signals that the SFTP upload is finished or aborted.
func (sf *streamFile) Close() error {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if !sf.closed {
		sf.closed = true
		close(sf.doneChan)
		close(sf.dataChan)
	}
	return nil
}

func (sf *streamFile) ReadHTTP(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", sf.fileName))

	// Simple loop: read chunk from dataChan, write to response
	for {
		select {
		case chunk, ok := <-sf.dataChan:
			if !ok {
				// dataChan closed, so SFTP upload ended
				return
			}
			if _, err := w.Write(chunk); err != nil {
				log.Printf("HTTP write error: %v", err)
				return
			}

		case <-sf.doneChan:
			// If doneChan is closed, we're finished
			return
		}
	}
}

type pipes struct {
	mu sync.Mutex
	m  map[string]*streamFile
}

type myFileWriter struct {
	sf *streamFile
}

func (w *myFileWriter) Filewrite(req *sftp.Request) (io.WriterAt, error) {
	if req.Method != "Put" {
		return nil, os.ErrPermission
	}

	log.Printf("Receiving upload for path: %s (method=%s)", req.Filepath, req.Method)
	return w.sf, nil
}

// FileReader: we *deny* reads (so GET/download won't work).
type myFileReader struct{}

func (r *myFileReader) Fileread(req *sftp.Request) (io.ReaderAt, error) {
	return nil, os.ErrPermission // or a custom error
}

type listerat []os.FileInfo

// Modeled after strings.Reader's ReadAt() implementation
func (f listerat) ListAt(ls []os.FileInfo, offset int64) (int, error) {
	var n int
	if offset >= int64(len(f)) {
		return 0, io.EOF
	}
	n = copy(ls, f[offset:])
	if n < len(ls) {
		return n, io.EOF
	}
	return n, nil
}

// FileLister: we *deny* directory listings (ls).
type myFileLister struct{}

func (l *myFileLister) Filelist(request *sftp.Request) (sftp.ListerAt, error) {
	switch request.Method {
	case "Stat":
		file, err := os.Stat(".")
		if err != nil {
			return nil, err
		}
		return listerat{file}, nil
	}

	return nil, errors.New("unsupported")
}

// FileCmder: we *deny* rename, link, symlink, etc.
type myFileCmd struct{}

func (c *myFileCmd) Filecmd(req *sftp.Request) error {
	return nil
}

func runSshServer(ps *pipes) error {
	cfg := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			// nothing for now
			return nil, nil
		},
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			// nothing for now
			return nil, nil
		},
		BannerCallback: func(conn ssh.ConnMetadata) string {
			// hash the session id using sha256
			hash := sha256.New()
			hash.Write(conn.SessionID())
			sessionID := fmt.Sprintf("%x", hash.Sum(nil))

			ps.mu.Lock()
			ps.m[sessionID] = newStreamFile()
			ps.mu.Unlock()

			// This string is sent to the SSH client before authentication
			return fmt.Sprintf("Welcome to my SSH server. Your session ID is %s\n", sessionID)
		},
	}

	bytes, err := os.ReadFile("./keys/ssh.pem")
	if err != nil {
		return fmt.Errorf("failed to read private key: %v", err)
	}

	key, err := ssh.ParsePrivateKey(bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}

	cfg.AddHostKey(key)

	listener, err := net.Listen("tcp", "127.0.0.1:2222")
	if err != nil {
		return fmt.Errorf("failed to listen on 127.0.0.1:2222: %v", err)
	}

	log.Print("SSH Server listening on 127.0.0.1:2222")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection: %v", err)
			continue
		}
		go handleIncomingSshConnection(conn, cfg, ps)
	}
}

func handleIncomingSshConnection(conn net.Conn, cfg *ssh.ServerConfig, ps *pipes) {
	shConn, chans, reqs, err := ssh.NewServerConn(conn, cfg)
	if err != nil {
		log.Printf("Failed to handshake: %v", err)
		return
	}
	defer shConn.Close()

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if t := newChannel.ChannelType(); t != "session" {
			newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("Could not accept channel %v", err)
			return
		}

		// Get the session ID
		hash := sha256.New()
		hash.Write(shConn.SessionID())
		sessionID := fmt.Sprintf("%x", hash.Sum(nil))
		sf, ok := ps.m[sessionID]
		if !ok {
			log.Printf("Session ID not found in pipes")
			return
		}

		go handleSession(channel, requests, sf)
	}
}
func handleSession(channel ssh.Channel, in <-chan *ssh.Request, sf *streamFile) {
	defer channel.Close()

	for req := range in {
		switch req.Type {
		case "subsystem":
			subsystem := parseSubsystem(req.Payload)
			if subsystem == "sftp" {
				// Accept the request
				req.Reply(true, []byte("Hello World!"))
				log.Println("Starting SFTP subsystem")

				handlers := sftp.Handlers{
					FileGet:  &myFileReader{},
					FilePut:  &myFileWriter{sf: sf},
					FileCmd:  &myFileCmd{},
					FileList: &myFileLister{},
				}

				// Start SFTP server
				sftpServer := sftp.NewRequestServer(channel, handlers)
				//sftpServer, _ := sftp.NewServer(channel)

				// Serve blocks until EOF or error
				if err := sftpServer.Serve(); err == io.EOF {
					log.Println("SFTP client exited session.")
				} else if err != nil {
					log.Printf("SFTP server completed with error: %v", err)
				}

				return
			}
			// If not "sftp", reject
			req.Reply(false, nil)

		default:
			// Reject all other request types
			req.Reply(false, nil)
		}
	}
}

func parseSubsystem(payload []byte) string {
	if len(payload) < 4 {
		return ""
	}
	length := (uint32(payload[0]) << 24) |
		(uint32(payload[1]) << 16) |
		(uint32(payload[2]) << 8) |
		(uint32(payload[3]))
	if int(length) > len(payload)-4 {
		return ""
	}
	return string(payload[4 : 4+length])
}
