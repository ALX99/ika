package router

import (
	"iter"
	"slices"
)

func collectIters[T any](iters ...iter.Seq[T]) []T {
	var t []T
	for _, it := range iters {
		t = append(t, slices.Collect(it)...)
	}
	return t
}
