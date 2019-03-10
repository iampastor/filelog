package filelog

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	defaultMaxFileSize = 128 * 1024 * 1024 // 128 MB
)

type Writer struct {
	Dir          string
	Name         string
	MaxFileSize  int64
	SyncInterval time.Duration
	fd           *os.File
	position     int64
	mu           sync.Mutex
	codec        Encoder
}

func NewWriter(dir string, name string) *Writer {
	return NewWriterWithEncoder(dir, name, NewLineEncoder())
}

func NewWriterWithEncoder(dir string, name string, codec Encoder) *Writer {
	return &Writer{
		Dir:         dir,
		Name:        name,
		MaxFileSize: defaultMaxFileSize,
		codec:       codec,
	}
}

func (l *Writer) Write(b []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	err := l.rotate()
	if err != nil {
		return 0, err
	}
	hn, err := l.codec.WriteHeader(b)
	bn, err := l.codec.WriteBody(b)
	if l.SyncInterval == 0 {
		l.fd.Sync()
	}
	l.position += int64(hn + bn)
	return hn + bn, err
}

func (w *Writer) Sync() {
	w.mu.Lock()
	w.fd.Sync()
	w.mu.Unlock()
}

func (l *Writer) rotate() error {
	if l.fd == nil || l.position > l.MaxFileSize {
		filename := l.getFilename()
		fpath := filepath.Join(l.Dir, filename)
		fd, err := os.OpenFile(fpath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return err
		}
		if l.fd != nil {
			l.fd.Close()
		}
		l.fd = fd
		l.position = 0
		l.codec.Reset(fd)
	}
	return nil
}

func (l *Writer) getFilename() string {
	t := time.Now().Unix()
	return fmt.Sprintf("%s_%d.data", l.Name, t)
}

func (l *Writer) Close() error {
	return l.fd.Close()
}
