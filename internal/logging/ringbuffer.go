package logging

import (
	"bytes"
	"sync"
)

type ringBuffer struct {
	SeenErr bool
	logs    []byte
	size    int
	idx     int
	full    bool
	mu      sync.Mutex
}

func newRingBuffer(size int) *ringBuffer {
	return &ringBuffer{
		logs: make([]byte, size),
		size: size,
	}
}

func (rb *ringBuffer) Write(p []byte) (int, error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	n := len(p)
	if n > rb.size {
		copy(rb.logs[0:], p[n-rb.size:])
		rb.idx = 0
		rb.full = true
		return n, nil
	}

	vacant := rb.size - rb.idx
	copy(rb.logs[rb.idx:], p[:min(vacant, n)])
	rb.idx += n
	hasLeft := n - vacant
	if hasLeft > 0 {
		copy(rb.logs[0:], p[vacant:])
		rb.idx = hasLeft
		rb.full = true
	}
	return n, nil
}

func (rb *ringBuffer) snapshot() []byte {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	result := make([]byte, 0, rb.size)
	if rb.full {
		result = append(result, rb.logs[rb.idx:]...)
	}
	result = append(result, rb.logs[:rb.idx]...)

	crIdx := bytes.IndexByte(result, '\n')
	if crIdx >= 0 && rb.full {
		result = result[crIdx:]
	}
	return result
}
