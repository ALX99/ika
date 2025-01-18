package pool_test

import (
	"net/http/httputil"
	"testing"

	"github.com/alx99/ika/internal/pool"
)

func BenchmarkBufferPool(b *testing.B) {
	var bp httputil.BufferPool = pool.NewBufferPool()
	for i := 0; i < b.N; i++ {
		buf := bp.Get()
		bp.Put(buf)
	}
}
