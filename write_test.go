package filelog

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"
)

type UserInfo struct {
	Username string
	Age      int
	Birthday time.Time
	Homepage string
	Image    []byte
}

var data = make([]byte, 1024)

var userInfo = &UserInfo{
	Username: "iampastor",
	Age:      12,
	Birthday: time.Now(),
	Homepage: "https://www.google.com",
	Image:    make([]byte, 1024),
}

func Test_Writer(t *testing.T) {
	w := NewWriter(".", "test")
	data, _ := json.Marshal(userInfo)
	for i := 0; i < 1; i++ {
		w.Write(data)
	}
	w.Close()
}

func Test_SizeWriter(t *testing.T) {
	w := NewSizeWriter(".", "sizetest", 1024)
	data, _ := json.Marshal(userInfo)
	for i := 0; i < 10; i++ {
		w.Write(data)
		time.Sleep(time.Second)
	}
	w.Close()
}
func Test_HourWriter(t *testing.T) {
	w := NewHourWriter(".", "hourtest")
	data, _ := json.Marshal(userInfo)
	for i := 0; i < 10; i++ {
		w.Write(data)
		time.Sleep(time.Second)
	}
	w.Close()
}
func Test_BinaryWriter(t *testing.T) {
	w := NewWriterWithEncoder(".", "test", NewBinaryEncoder())
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	encoder.Encode(userInfo)
	for i := 0; i < 10; i++ {
		w.Write(buf.Bytes())
	}
	w.Close()
}

func Benchmark_WriteNoLock(b *testing.B) {
	f, err := os.OpenFile("/tmp/write_benchmark_nolock", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		b.Fatalf("open file %s", err.Error())
	}
	for i := 0; i < b.N; i++ {
		f.Write(data)
	}
	f.Close()
}

func Benchmark_WriteWithLock(b *testing.B) {
	f, err := os.OpenFile("/tmp/write_benchmark_lock", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		b.Fatalf("open file %s", err.Error())
	}
	var m sync.Mutex
	for i := 0; i < b.N; i++ {
		m.Lock()
		f.Write(data)
		m.Unlock()
	}
	f.Close()
}

func Benchmark_WriteWithSync(b *testing.B) {
	f, err := os.OpenFile("/tmp/write_benchmark_sync", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		b.Fatalf("open file %s", err.Error())
	}
	for i := 0; i < b.N; i++ {
		f.Write(data)
		f.Sync()
	}
	f.Close()
}

func Benchmark_WriteWithLockSync(b *testing.B) {
	f, err := os.OpenFile("/tmp/write_benchmark_lock_sync", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		b.Fatalf("open file %s", err.Error())
	}
	var m sync.Mutex
	for i := 0; i < b.N; i++ {
		m.Lock()
		f.Write(data)
		f.Sync()
		m.Unlock()
	}
	f.Close()
}
