package logger

import (
	"testing"
	"time"
)

type slowWriter struct{}

func (w *slowWriter) Write(p []byte) (n int, err error) {
	time.Sleep(100 * time.Millisecond)
	return len(p), nil
}

func BenchmarkWriter(b *testing.B) {
	w := newBufferedWriter(&slowWriter{})
	time.Sleep(1 * time.Second)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w.Write([]byte("testaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	}
	w.Flush()
}
