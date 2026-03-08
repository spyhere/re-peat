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
		PcmLen:         pcmLen,
		MonoSamplesLen: int(pcmLen) / dec.Channels / constants.BytesPerSample,
	}
	a.Seconds = a.GetSecondsFromSamples(a.GetSamplesFromPCM(pcmLen))
	return a
}

type Audio struct {
	SampleRate     int
	Channels       int
	PcmLen         int64
	MonoSamplesLen int
	Seconds        float64
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
func (a Audio) GetSamplesFromPCM(pcmBytes int64) int {
	return int(pcmBytes / (int64(a.Channels) * constants.BytesPerSample))
}
func (a Audio) GetSecondsFromPCM(pcmBytes int64) float64 {
	return a.GetSecondsFromSamples(a.GetSamplesFromPCM(pcmBytes))
}
func (a Audio) GetPcmFromSeconds(seconds float64) int64 {
	return a.GetPcmFromSamples(a.GetSamplesFromSeconds(seconds))
}
func (a Audio) GetPcmFromSamples(samples int) int64 {
	return int64(samples * a.Channels * constants.BytesPerSample)
}
