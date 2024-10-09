package proxy

import "sync"

var pool = &bufferPool{
	sync.Pool{
		New: func() any {
			b := make([]byte, 5*1024)
			return &b
		},
	},
}

type bufferPool struct{ sync.Pool }

func (bp *bufferPool) Get() []byte {
	return *bp.Pool.Get().(*[]byte)
}

func (bp *bufferPool) Put(b []byte) {
	bp.Pool.Put(&b)
}
