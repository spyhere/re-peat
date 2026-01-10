package waverenderer

import (
	"math"

	"github.com/spyhere/re-peat/internal/constants"
)

type audio struct {
	sampleRate   int
	samplesPerPx int
	channels     int
	pcmLen       int64
	pcmMonoLen   int
	seconds      float32
	secsPerByte  float32
}

func (a audio) getSecondsFromSamples(samplesIdx int) float64 {
	return float64(samplesIdx) / float64(a.sampleRate)
}
func (a audio) getNextSecond(second float64) (nextSecond float64, sammplesIdx int) {
	nextSecond = math.Ceil(second)
	return nextSecond, int(nextSecond * float64(a.sampleRate))
}
func (a audio) getSamplesFromPCM(pcmBytes int64) int {
	return int(pcmBytes / int64(a.channels) / constants.BYTES_PER_SAMPLE)
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

func (s scroll) getPxPerSec() float32 {
	return max(s.minPxPerSec, s.pxPerSec)
}
