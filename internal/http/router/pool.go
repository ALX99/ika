package router

import "sync"

type bufferPool struct{ sync.Pool }

func newBufferPool() *bufferPool {
	return &bufferPool{Pool: sync.Pool{
		New: func() any {
			s := make([]byte, 32*1024)
			return &s
		},
	}}
}

func (bp *bufferPool) Get() []byte {
	ptr := bp.Pool.Get().(*[]byte)
	return *ptr
}

func (bp *bufferPool) Put(b []byte) {
	bp.Pool.Put(&b)
}
