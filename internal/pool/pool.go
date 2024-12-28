package pool

import (
	"sync"
)

type BufferPool struct{ sync.Pool }

func NewBufferPool() *BufferPool {
	return &BufferPool{Pool: sync.Pool{
		New: func() any {
			s := make([]byte, 32*1024)
			return &s
		},
	}}
}

func (bp *BufferPool) Get() []byte {
	ptr := bp.Pool.Get().(*[]byte)
	return *ptr
}

func (bp *BufferPool) Put(b []byte) {
	bp.Pool.Put(&b)
}
