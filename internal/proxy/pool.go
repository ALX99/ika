package proxy

import (
	"sync"
)

var strSlicePool = &sync.Pool{
	New: func() any { return &[]string{} },
}
