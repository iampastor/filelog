package filelog

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
)

type Encoder interface {
	WriteHeader(data []byte) (int, error)
	WriteBody(data []byte) (int, error)
	Reset(w io.Writer)
}

type LineEncoder struct {
	writer *bufio.Writer
}

func NewLineEncoder() *LineEncoder {
	return &LineEncoder{
		writer: bufio.NewWriter(ioutil.Discard),
	}
}

func (w *LineEncoder) WriteHeader(data []byte) (int, error) {
	return 0, nil
}

func (w *LineEncoder) WriteBody(data []byte) (int, error) {
	n, err := w.writer.Write(data)
	if err != nil {
		return 0, err
	}
	w.writer.WriteByte('\n')
	w.writer.Flush()
	return n + 1, nil
}

func (w *LineEncoder) Reset(nw io.Writer) {
	w.writer.Reset(nw)
}

type Decoder interface {
	ReadHeader() (int, error)
	ReadBody() (int, []byte, error)
	Reset(r io.Reader)
}

type LineDecoder struct {
	reader *bufio.Reader
}

func NewLineDecoder() *LineDecoder {
	return &LineDecoder{
		reader: bufio.NewReader(bytes.NewReader(nil)),
	}
}

func (r *LineDecoder) ReadHeader() (int, error) {
	return 0, nil
}

func (r *LineDecoder) ReadBody() (int, []byte, error) {
	data, err := r.reader.ReadBytes('\n')
	if err != nil {
		return 0, nil, err
	}
	return len(data), data[:len(data)-1], nil
}

func (r *LineDecoder) Reset(nr io.Reader) {
	r.reader.Reset(nr)
}
