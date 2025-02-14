package router

import (
	"net/http/httputil"
	"testing"
)

func BenchmarkBufferPool(b *testing.B) {
	var bp httputil.BufferPool = newBufferPool()
	for b.Loop() {
		buf := bp.Get()
		bp.Put(buf)
	}
}
