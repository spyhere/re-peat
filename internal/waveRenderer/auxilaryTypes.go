package waverenderer

type audio struct {
	sampleRate   int
	samplesPerPx int
	channels     int
	pcmLen       int
	pcmMonoLen   int
	seconds      float32
	secsPerByte  float32
}

type scroll struct {
	originX      float32 // position of cursor when scrolling
	pxPerSec     float32
	minPxPerSec  float32
	maxPxPerSec  float32
	leftB        int     // left border of samples to skip what's outside of visible range
	rightB       int     // right border of samples
	zoomExpDelta float32 // zoom exponent delta
	maxZoomExp   float32 // max zoom exponent delta
}
