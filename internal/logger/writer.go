package logger

import (
	"bufio"
	"io"
	"sync"
	"sync/atomic"
)

type bufferedWriter struct {
	w               io.Writer
	bw              *bufio.Writer
	writeToBuffered atomic.Bool
	sync.RWMutex
}

func newBufferedWriter(w io.Writer) *bufferedWriter {
	bw := bufferedWriter{
		w:               w,
		bw:              bufio.NewWriterSize(w, 1<<23), // 8MB
		writeToBuffered: atomic.Bool{},
	}
	bw.writeToBuffered.Store(true)
	return &bw
}

// SetBuffered enables or disables buffering
func (b *bufferedWriter) SetBuffered(enabled bool) {
	b.writeToBuffered.Store(enabled)
}

func (b *bufferedWriter) Write(p []byte) (int, error) {
	if !b.writeToBuffered.Load() {
		return b.w.Write(p)
	}
	b.Lock()
	n, err := b.bw.Write(p)
	b.Unlock()
	return n, err
}

func (b *bufferedWriter) Flush() error {
	b.Lock()
	err := b.bw.Flush()
	b.Unlock()
	return err
}
