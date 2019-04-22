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
	maxFileSize  int64
	syncInterval time.Duration
	fd           *os.File
	position     int64
	filename     string
	mu           sync.Mutex
	codec        Encoder
	rotateSize   bool
	rotateHour   bool
}

func NewWriter(dir string, name string) *Writer {
	return NewWriterWithEncoder(dir, name, NewLineEncoder())
}

func NewWriterWithEncoder(dir string, name string, codec Encoder) *Writer {
	w := &Writer{
		Dir:   dir,
		Name:  name,
		codec: codec,
	}

	return w
}

func NewSizeWriter(dir string, name string, maxFilesize int64) *Writer {
	return NewSizeWriterWithEncoder(dir, name, maxFilesize, NewLineEncoder())
}

func NewSizeWriterWithEncoder(dir string, name string, maxFilesize int64, codec Encoder) *Writer {
	w := NewWriterWithEncoder(dir, name, codec)
	if maxFilesize == 0 {
		maxFilesize = defaultMaxFileSize
	}
	w.maxFileSize = maxFilesize
	w.rotateSize = true
	return w
}

func NewHourWriter(dir string, name string) *Writer {
	return NewHourWriterWithEncoder(dir, name, NewLineEncoder())
}

func NewHourWriterWithEncoder(dir string, name string, codec Encoder) *Writer {
	w := NewWriterWithEncoder(dir, name, codec)
	w.rotateHour = true
	return w
}

func (l *Writer) Write(b []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	var err error
	if l.rotateSize {
		err = l.rotateBySize()
	} else if l.rotateHour {
		err = l.rorateByHour()
	} else {
		err = l.rotateByDay()
	}
	if err != nil {
		return 0, err
	}
	hn, err := l.codec.WriteHeader(b)
	bn, err := l.codec.WriteBody(b)
	l.position += int64(hn + bn)
	return hn + bn, err
}

func (w *Writer) Sync() {
	w.mu.Lock()
	w.fd.Sync()
	w.mu.Unlock()
}

func (w *Writer) syncLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			w.Sync()
		}
	}
}

func (l *Writer) rotate(filename string) error {
	fpath := filepath.Join(l.Dir, filename)
	fd, err := os.OpenFile(fpath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	if l.fd != nil {
		l.fd.Close()
	}
	l.fd = fd
	l.position = 0
	l.filename = filename
	l.codec.Reset(fd)
	return nil
}

func (l *Writer) rotateBySize() error {
	if l.fd == nil || l.position > l.maxFileSize {
		filename := l.getSizeFilename()
		return l.rotate(filename)
	}
	return nil
}

func (l *Writer) rotateByDay() error {
	filename := l.getDayFilename()
	if l.fd == nil || l.filename != filename {
		return l.rotate(filename)
	}
	return nil
}

func (l *Writer) rorateByHour() error {
	filename := l.getHourFilename()
	if l.fd == nil || l.filename != filename {
		return l.rotate(filename)
	}
	return nil
}

func (l *Writer) getDayFilename() string {
	t := time.Now().Format("2006-01-02")
	return fmt.Sprintf("%s%s.log", l.Name, t)
}

func (l *Writer) getSizeFilename() string {
	t := time.Now().Unix()
	return fmt.Sprintf("%s%d.log", l.Name, t)
}

func (l *Writer) getHourFilename() string {
	t := time.Now().Format("2006-01-02-15")
	return fmt.Sprintf("%s%s.log", l.Name, t)
}

func (l *Writer) Close() error {
	return l.fd.Close()
}
