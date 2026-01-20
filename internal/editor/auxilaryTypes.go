package editor

import (
	"math"

	"github.com/spyhere/re-peat/internal/constants"
)

// TODO: Split this file

type audio struct {
	sampleRate int
	channels   int
	pcmLen     int64
	pcmMonoLen int
	seconds    float32
}

func (a audio) getSecondsFromSamples(samplesIdx int) float64 {
	return float64(samplesIdx) / float64(a.sampleRate)
}
func (a audio) getNextSecond(second float64) (nextSecond float64, sammplesIdx int) {
	nextSecond = math.Ceil(second)
	return nextSecond, int(nextSecond * float64(a.sampleRate))
}
func (a audio) getSamplesFromPCM(pcmBytes int64) int {
	return int(pcmBytes / int64(a.channels) / constants.BytesPerSample)
}

type scroll struct {
	leftB           int // left border of samples to skip what's outside of visible range
	rightB          int // right border of samples
	maxLvl          int
	samplesPerPx    float32
	minSamplesPerPx float32
	maxSamplesPerPx float32
}

// Stores peak map where "samplesPerPx" is key (level)
type cache struct {
	peakMap     map[int][][2]float32
	curSlice    [][2]float32
	workers     []*cacheWorker
	isPopulated bool
	levels      []int // Stores possible "samplesPerPx" values
	curLvl      int
	leftB       int
}

// Used to build one level (samplesPerPx) of cache
type cacheWorker struct {
	min          float32
	max          float32
	samplesPerPx int
	count        int
	sliceIdx     int
}

func (c cache) getLevel(spp float32) int {
	for i := len(c.levels) - 1; i >= 0; i-- {
		if float32(c.levels[i]) >= spp {
			return c.levels[i]
		}
	}
	return c.levels[0]
}
