package editor

type scroll struct {
	leftB           int // left border of samples to skip what's outside of visible range
	rightB          int // right border of samples
	maxLvl          int
	samplesPerPx    float32
	minSamplesPerPx float32
	maxSamplesPerPx float32
}

func (s scroll) getSamplesFromPx(x float32) int {
	return s.leftB + int(s.samplesPerPx*x)
}
