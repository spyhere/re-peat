package logging

import (
	"log/slog"
	"sync"
	"time"
)

type logEntry struct {
	Time  time.Time
	Level slog.Level
	Msg   string
}

type ringBuffer struct {
	entries []logEntry
	size    int
	idx     int
	full    bool
	mu      sync.Mutex
}

func newRingBuffer(size int) *ringBuffer {
	return &ringBuffer{
		entries: make([]logEntry, size),
		size:    size,
	}
}

func (rb *ringBuffer) add(e logEntry) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	e.Time = time.Now()

	rb.entries[rb.idx] = e
	rb.idx = (rb.idx + 1) % rb.size

	if rb.idx == 0 {
		rb.full = true
	}
}

func (rb *ringBuffer) snapshot() []logEntry {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	var result []logEntry
	if rb.full {
		result = append(result, rb.entries[rb.idx:]...)
	}
	result = append(result, rb.entries[:rb.idx]...)
	return result
}
