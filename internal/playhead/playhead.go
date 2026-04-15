package playhead

type Transport struct {
	Samples       int
	playbackStart int
}

func (p *Transport) Set(samples int) {
	p.playbackStart = samples
	p.Samples = samples
}

func (p *Transport) Reset() {
	p.Samples = p.playbackStart
}
