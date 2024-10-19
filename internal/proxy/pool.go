package proxy

import (
	"sync"

	"github.com/valyala/bytebufferpool"
)

var strSlicePool = &sync.Pool{
	New: func() any { return &[]string{} },
}

type bufferPool struct{ bytebufferpool.Pool }

func (bp *bufferPool) Get() []byte  { return bp.Pool.Get().B }
func (bp *bufferPool) Put(b []byte) { bp.Pool.Put(&bytebufferpool.ByteBuffer{B: b}) }
