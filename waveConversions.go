package main

import (
	"encoding/binary"
	"fmt"
)

const maxUin16 float32 = 32767.0

func getNormalisedSamples(data []byte) ([]float32, error) {
	if len(data)%2 != 0 {
		return []float32{}, fmt.Errorf("Read samples are not uint16: %d\n", len(data))
	}
	normalised := make([]float32, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		sample := int16(binary.LittleEndian.Uint16(data[i : i+2]))
		// Normalise to -1..1
		normalised[i/2] = float32(sample) / maxUin16
	}
	return normalised, nil
}

type RenderableWaves struct {
	SampleRate int
	// How many pixels are designated for 1 second of audio
	PxPerSec int
	// Left border where the decoded audio data should be read from (for zooming functionality)
	LeftB int
	// Right border where the decoded audio data should be read to (for zooming functionality)
	RightB       int
	SamplesPerPx int
	Samples      []float32
}

func (r *RenderableWaves) SetMaxX(maxX int) {
	seconds := len(r.Samples) / r.SampleRate
	r.PxPerSec = maxX / seconds
	r.SamplesPerPx = r.SampleRate / (maxX / seconds)
}

func (r RenderableWaves) GetRenderableWaves() [][2]float32 {
	samples := r.Samples
	if r.RightB == 0 {
		r.RightB = len(samples)
	}
	samples = samples[r.LeftB:r.RightB]
	res := make([][2]float32, len(samples)/r.SamplesPerPx)

	var idx int
	var min float32 = 1
	var max float32 = -1
	count := r.SamplesPerPx
	for _, it := range samples {
		if it < min {
			min = it
		}
		if it > max {
			max = it
		}
		count--
		if count == 0 {
			res[idx] = [2]float32{min, max}
			idx++
			min = 1
			max = -1
			count = r.SamplesPerPx
		}
	}

	return res
}
