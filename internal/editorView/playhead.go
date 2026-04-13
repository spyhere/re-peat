package editorview

import (
	"time"
)

func newPlayhead(updateTime time.Duration) *playhead {
	return &playhead{
		update: updateTime,
	}
}

// NOTE: Should it be inside AppState? reset playhead on new file?
type playhead struct {
	samples     int
	prevSamples int
	update      time.Duration
}

func (p *playhead) set(samples int) {
	p.prevSamples = samples
	p.samples = samples
}

func (p *playhead) reset() {
	p.samples = p.prevSamples
}
