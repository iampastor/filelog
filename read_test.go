package filelog

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"
)

type PrintHandler struct{}

func (h *PrintHandler) HandleData(data []byte, metaData MetaData) {
	fmt.Printf("%+v: %s\n", metaData, string(data))
}

func Test_Reader(t *testing.T) {
	reader, err := NewReader(".", "test")
	if err != nil {
		t.Fatal(err)
	}

	err = reader.Handle(&PrintHandler{})
	if err != nil {
		t.Fatal(err)
	}
	reader.Close()
}

type BinaryHandler struct{}

func (h *BinaryHandler) HandleData(data []byte, metaData MetaData) {
	u := UserInfo{}
	var buf bytes.Buffer
	buf.Write(data)
	decoder := gob.NewDecoder(&buf)
	decoder.Decode(&u)
	fmt.Printf("%+v: %+v\n", metaData, u)
}

func Test_BinaryReader(t *testing.T) {
	reader, err := NewReaderWithCodec(".", "test", NewBinaryDecoder())
	if err != nil {
		t.Fatal(err)
	}

	err = reader.Handle(&BinaryHandler{})
	if err != nil {
		t.Fatal(err)
	}
	reader.Close()
}
