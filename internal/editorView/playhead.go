package editorview

type playhead struct {
	samples       int
	playbackStart int
}

func (p *playhead) set(samples int) {
	p.playbackStart = samples
	p.samples = samples
}

func (p *playhead) reset() {
	p.samples = p.playbackStart
}
