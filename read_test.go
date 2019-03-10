package filelog

import (
	"fmt"
	"testing"
)

type PrintHandler struct{}

func (h *PrintHandler) HandleData(data []byte, metaData MetaData) {
	fmt.Printf("%+v: %s\n", metaData, string(data))
}

func Test_Read(t *testing.T) {
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
