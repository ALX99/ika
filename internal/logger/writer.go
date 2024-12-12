package logger

import (
	"bufio"
	"sync"
)

type bufferedWriter struct {
	w *bufio.Writer
	sync.RWMutex
}

func newBufferedWriter(w *bufio.Writer) *bufferedWriter {
	return &bufferedWriter{w: w}
}

func (b *bufferedWriter) Write(p []byte) (int, error) {
	b.Lock()
	n, err := b.w.Write(p)
	b.Unlock()
	return n, err
}

func (b *bufferedWriter) Flush() error {
	b.Lock()
	err := b.w.Flush()
	b.Unlock()
	return err
}
