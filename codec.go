package filelog

import (
	"bufio"
	"bytes"
	"encoding/binary"
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

const (
	binaryHeadLength = 4
)

type BinaryEncoder struct {
	writer *bufio.Writer
}

func NewBinaryEncoder() *BinaryEncoder {
	return &BinaryEncoder{
		writer: bufio.NewWriter(ioutil.Discard),
	}
}

func (w *BinaryEncoder) WriteHeader(data []byte) (int, error) {
	buf := make([]byte, binaryHeadLength)
	binary.BigEndian.PutUint32(buf, uint32(len(data)))
	return w.writer.Write(buf)
}

func (w *BinaryEncoder) WriteBody(data []byte) (int, error) {
	n, err := w.writer.Write(data)
	if err != nil {
		return 0, err
	}
	w.writer.Flush()
	return n, nil
}

func (w *BinaryEncoder) Reset(nw io.Writer) {
	w.writer.Reset(nw)
}

type BinaryDecoder struct {
	reader  *bufio.Reader
	dataLen uint32
}

func NewBinaryDecoder() *BinaryDecoder {
	return &BinaryDecoder{
		reader: bufio.NewReader(bytes.NewReader(nil)),
	}
}

func (r *BinaryDecoder) ReadHeader() (int, error) {
	buf := make([]byte, binaryHeadLength)
	n, err := io.ReadFull(r.reader, buf)
	if err != nil {
		return 0, err
	}
	r.dataLen = binary.BigEndian.Uint32(buf)
	return n, err
}

func (r *BinaryDecoder) ReadBody() (int, []byte, error) {
	data := make([]byte, r.dataLen)
	n, err := io.ReadFull(r.reader, data)
	return n, data, err
}

func (r *BinaryDecoder) Reset(nr io.Reader) {
	r.reader.Reset(nr)
}
