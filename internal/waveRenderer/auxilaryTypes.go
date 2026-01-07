package waverenderer

type audio struct {
	sampleRate  int
	channels    int
	pcmLen      int
	pcmMonoLen  int
	seconds     float32
	secsPerByte float32
}

type scroll struct {
	minX        float32
	maxX        float32
	deltaX      float32
	deltaY      float32
	originX     float32
	minPxPerSec float32
	maxPxPerSec float32
	leftB       int
	rightB      int
}
