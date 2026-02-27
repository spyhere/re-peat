package editorview

import (
	"time"
)

func newPlayhead(updateTime time.Duration) *playhead {
	return &playhead{
		update: updateTime,
	}
}

type playhead struct {
	bytes     int64 // pcm bytes
	prevBytes int64
	update    time.Duration
}

func (p *playhead) set(pcm int64) {
	p.prevBytes = pcm
	p.bytes = pcm
}

func (p *playhead) reset() {
	p.bytes = p.prevBytes
}
