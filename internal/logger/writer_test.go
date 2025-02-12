package logger

import (
	"testing"
	"time"

	"github.com/matryer/is"
)

type slowWriter struct{}

func (w *slowWriter) Write(p []byte) (n int, err error) {
	time.Sleep(100 * time.Millisecond)
	return len(p), nil
}

func BenchmarkWriter(b *testing.B) {
	is := is.New(b)
	w := newBufferedWriter(&slowWriter{})
	time.Sleep(1 * time.Second)
	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, err := w.Write([]byte("testaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
		is.NoErr(err)
	}
	if err := w.Flush(); err != nil {
		b.Fatal(err)
	}
	is.NoErr(w.Flush())
}
