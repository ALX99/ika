package proxy

import "sync"

var (
	bufPool = &bufferPool{
		sync.Pool{
			New: func() any {
				b := make([]byte, 32*1024)
				return &b
			},
		},
	}
	strSlicePool = &sync.Pool{
		New: func() any { return &[]string{} },
	}
)

type bufferPool struct{ sync.Pool }

func (bp *bufferPool) Get() []byte {
	return *bp.Pool.Get().(*[]byte)
}

func (bp *bufferPool) Put(b []byte) {
	bp.Pool.Put(&b)
}
