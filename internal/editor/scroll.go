package editor

const maxScrollLvl = 5

func newScroll() scroll {
	return scroll{
		maxLvl: maxScrollLvl,
	}
}

type scroll struct {
	leftB           int // left border of samples to skip what's outside of visible range
	rightB          int // right border of samples
	maxLvl          int
	samplesPerPx    float32
	minSamplesPerPx float32
	maxSamplesPerPx float32
}
