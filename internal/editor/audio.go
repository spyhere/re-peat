package editor

import (
	"math"

	"github.com/spyhere/re-peat/internal/constants"
)

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
