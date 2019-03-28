package filelog

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/pkg/errors"
)

type MetaData struct {
	filename string
	position int64
	f        *os.File
}

func (m *MetaData) write() error {
	err := m.f.Truncate(0)
	if err != nil {
		return errors.Wrap(err, "truncate metadata")
	}
	_, err = m.f.Seek(0, 0)
	if err != nil {
		return errors.Wrap(err, "seek metadata")
	}
	_, err = m.f.Write([]byte(fmt.Sprintf("%s\n%d\n", m.filename, m.position)))
	return errors.Wrap(err, "write metadata")
}

func (m *MetaData) read() error {
	r := bufio.NewReader(m.f)
	fname, err := r.ReadString('\n')
	if err != nil {
		return errors.Wrap(err, "read metadata filename")
	}

	position, err := r.ReadString('\n')
	if err != nil {
		return errors.Wrap(err, "read metadata position")
	}

	pos, err := strconv.ParseInt(position[:len(position)-1], 10, 64)
	if err != nil {
		return errors.Wrap(err, "parse metadata")
	}
	m.filename = fname[:len(fname)-1]
	m.position = pos
	return nil
}

func (m *MetaData) close() error {
	return m.f.Close()
}

type Reader struct {
	Dir      string
	Name     string
	metaData *MetaData
	codec    Decoder
	closed   bool
	mu       sync.RWMutex
}

type ReadHandler interface {
	HandleData(data []byte, metaData MetaData)
}

func NewReader(dir string, name string) (*Reader, error) {
	return NewReaderWithCodec(dir, name, NewLineDecoder())
}

func NewReaderWithCodec(dir string, name string, codec Decoder) (*Reader, error) {
	r := &Reader{
		Dir:   dir,
		Name:  name,
		codec: codec,
	}

	metaData, err := r.getMetaData()
	if err != nil {
		return r, err
	}
	r.metaData = metaData
	return r, nil
}

func (r *Reader) getMetaData() (*MetaData, error) {
	metaData := &MetaData{}
	mfileName := r.getMetaDataFileName()
	_, err := os.Stat(mfileName)
	if err != nil {
		if os.IsNotExist(err) {
			f, err := os.Create(mfileName)
			if err != nil {
				return metaData, err
			}
			metaData.f = f
			return metaData, nil
		} else {
			return nil, err
		}
	}
	f, err := os.OpenFile(mfileName, os.O_RDWR, 0666)
	if err != nil {
		return nil, errors.Wrapf(err, "open metadata file %s", mfileName)
	}
	metaData.f = f
	err = metaData.read()
	return metaData, err
}

func (r *Reader) listFiles() ([]string, error) {
	var files []string

	list, err := ioutil.ReadDir(r.Dir)
	if err != nil {
		return files, errors.Wrap(err, "list files")
	}
	filePattern := r.getFileNamePattern()
	for _, fi := range list {
		match, err := filepath.Match(filePattern, fi.Name())
		if err != nil {
			return nil, errors.Wrap(err, "match files")
		}
		if match && fi.Name() >= r.metaData.filename {
			files = append(files, fi.Name())
		}
	}
	return files, nil
}

func (r *Reader) getFileNamePattern() string {
	return fmt.Sprintf("%s*.log", r.Name)
}

func (r *Reader) getMetaDataFileName() string {
	return fmt.Sprintf("%s.metadata", r.Name)
}

func (r *Reader) Handle(handler ReadHandler) error {
	files, err := r.listFiles()
	if err != nil {
		return err
	}
	for _, file := range files {
		if !r.isClosed() {
			r.metaData.filename = file
			err := r.readFile(file, handler)
			if err != nil {
				break
			}
			r.metaData.position = 0
		}
	}
	r.metaData.close()
	if r.isClosed() {
		return io.ErrUnexpectedEOF
	} else {
		return nil
	}
}

func (r *Reader) readFile(fname string, handler ReadHandler) error {
	f, err := os.Open(fname)
	if err != nil {
		return errors.Wrapf(err, "open file %s", fname)
	}
	defer f.Close()
	if r.metaData.position != 0 {
		_, err = f.Seek(r.metaData.position, 0)
		if err != nil {
			return errors.Wrapf(err, "seek file %s", fname)
		}
	}
	r.codec.Reset(f)

	for !r.isClosed() {
		hn, err := r.codec.ReadHeader()
		if err != nil {
			if err == io.EOF {
				return nil
			} else {
				return errors.Wrapf(err, "read header %s", fname)
			}
		}
		bn, data, err := r.codec.ReadBody()
		if err != nil {
			if err == io.EOF {
				return nil
			} else {
				return errors.Wrapf(err, "read file %s", fname)
			}
		}
		handler.HandleData(data, *r.metaData)
		r.metaData.position += int64(hn + bn)
		if err := r.metaData.write(); err != nil {
			return errors.Wrap(err, "write metadata")
		}
	}
	return nil
}

func (r *Reader) isClosed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.closed
}

func (r *Reader) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
}
