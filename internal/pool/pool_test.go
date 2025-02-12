package pool_test

import (
	"net/http/httputil"
	"testing"

	"github.com/alx99/ika/internal/pool"
)

func BenchmarkBufferPool(b *testing.B) {
	var bp httputil.BufferPool = pool.NewBufferPool()
	for b.Loop() {
		buf := bp.Get()
		bp.Put(buf)
	}
}
