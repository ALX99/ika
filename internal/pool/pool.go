package pool

import "github.com/valyala/bytebufferpool"

type BufferPool struct{ bytebufferpool.Pool }

func (bp *BufferPool) Get() []byte  { return bp.Pool.Get().B }
func (bp *BufferPool) Put(b []byte) { bp.Pool.Put(&bytebufferpool.ByteBuffer{B: b}) }
