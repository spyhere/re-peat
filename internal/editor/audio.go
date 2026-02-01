package editor

import (
	"math"

	"github.com/spyhere/re-peat/internal/constants"
	"github.com/tosone/minimp3"
)

func newAudio(dec *minimp3.Decoder, pcm []byte, monoSamples []float32, frames int) audio {
	return audio{
		sampleRate: dec.SampleRate,
		channels:   dec.Channels,
		pcmLen:     int64(len(pcm)),
		pcmMonoLen: len(monoSamples),
		seconds:    float32(frames) / float32(dec.SampleRate),
	}
}

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
	return int(pcmBytes / (int64(a.channels) * constants.BytesPerSample))
}
func (a audio) getPcmFromSamples(samples int) int64 {
	return int64(samples * a.channels * constants.BytesPerSample)
}
