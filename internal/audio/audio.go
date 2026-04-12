package audio

import (
	"math"

	"github.com/spyhere/re-peat/internal/constants"
	"github.com/tosone/minimp3"
)

func NewAudio(dec *minimp3.Decoder, pcm []byte) Audio {
	pcmLen := int64(len(pcm))
	a := Audio{
		SampleRate:     dec.SampleRate,
		Channels:       dec.Channels,
		MonoSamplesLen: int(pcmLen) / dec.Channels / constants.BytesPerSample,
	}
	a.Seconds = a.GetSecondsFromSamples(a.getSamplesFromPCM(pcmLen))
	return a
}

type Audio struct {
	SampleRate     int
	Channels       int
	MonoSamplesLen int
	Seconds        float64
}

func (a Audio) getSamplesFromPCM(pcmBytes int64) int {
	return int(pcmBytes / (int64(a.Channels) * constants.BytesPerSample))
}
func (a Audio) GetSecondsFromSamples(samplesIdx int) float64 {
	return float64(samplesIdx) / float64(a.SampleRate)
}
func (a Audio) GetNextSecond(second float64) (nextSecond float64, sammplesIdx int) {
	nextSecond = math.Ceil(second)
	return nextSecond, int(nextSecond * float64(a.SampleRate))
}
func (a Audio) GetSamplesFromSeconds(seconds float64) int {
	return int(seconds * float64(a.SampleRate))
}
func (a Audio) GetSamplesAmount() int {
	return a.MonoSamplesLen * a.Channels
}
