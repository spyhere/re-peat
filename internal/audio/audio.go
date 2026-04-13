package audio

import (
	"fmt"
	"math"

	"github.com/spyhere/re-peat/internal/common"
)

func NewAudioMeta(sampleRate, channels int, maxSamples int) AudioMeta {
	a := AudioMeta{
		SampleRate:     sampleRate,
		Channels:       channels,
		MonoSamplesLen: maxSamples,
	}
	a.Seconds = a.GetSecondsFromSamples(maxSamples)
	return a
}

type AudioMeta struct {
	SampleRate     int
	Channels       int
	MonoSamplesLen int
	Seconds        float64
}

func (a AudioMeta) GetSecondsFromSamples(samplesIdx int) float64 {
	return float64(samplesIdx) / float64(a.SampleRate)
}
func (a AudioMeta) GetNextSecond(second float64) (nextSecond float64, samplesIdx int) {
	nextSecond = math.Ceil(second)
	return nextSecond, int(nextSecond * float64(a.SampleRate))
}
func (a AudioMeta) GetSamplesFromSeconds(seconds float64) int {
	return int(seconds * float64(a.SampleRate))
}
func (a AudioMeta) GetSamplesAmount() int {
	return a.MonoSamplesLen * a.Channels
}

func (a AudioMeta) SecondsString() string {
	if a.Seconds == 0.0 {
		return ""
	}
	return common.FormatSeconds(a.Seconds)
}

func (a AudioMeta) ChannelsString() string {
	if a.Channels == 0 {
		return ""
	}
	switch a.Channels {
	case 1:
		return "Mono"
	case 2:
		return "Stereo"
	default:
		panic("unreachable")
	}
}

func (a AudioMeta) SampleRateString() string {
	if a.SampleRate == 0 {
		return ""
	}
	return fmt.Sprintf("%.1f kHz", float64(a.SampleRate)/1000)
}
