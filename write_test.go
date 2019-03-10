package filelog

import (
	"os"
	"sync"
	"testing"
)

var data = make([]byte, 1024)

func Test_Writer(t *testing.T) {
	w := NewWriter(".", "test")
	for i := 0; i < 10; i++ {
		w.Write([]byte("test data"))
	}
	w.Close()
}

func Benchmark_WriteNoLock(b *testing.B) {
	f, err := os.OpenFile("/tmp/write_benchmark_nolock", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		b.Fatalf("open file %s", err.Error())
	}
	for i := 0; i < b.N; i++ {
		f.Write(data)
	}
	f.Close()
}

func Benchmark_WriteWithLock(b *testing.B) {
	f, err := os.OpenFile("/tmp/write_benchmark_lock", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
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
	f, err := os.OpenFile("/tmp/write_benchmark_sync", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
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
	f, err := os.OpenFile("/tmp/write_benchmark_lock_sync", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
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
