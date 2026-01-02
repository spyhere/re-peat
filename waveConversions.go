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

type ViewState struct {
	// How many pixels are designated for 1 second of audio
	PxPerSec int
	// Left border where the decoded audio data should be read from (for zooming functionality)
	LeftB int
	// Right border where the decoded audio data should be read to (for zooming functionality)
	RightB       int
	SamplesPerPx int
}

func getRenderableWave(samples []float32, view ViewState) [][2]float32 {
	if view.RightB == 0 {
		view.RightB = len(samples)
	}
	samples = samples[view.LeftB:view.RightB]
	res := make([][2]float32, len(samples)/view.SamplesPerPx)

	var idx int
	var min float32 = 1
	var max float32 = -1
	count := view.SamplesPerPx
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
			count = view.SamplesPerPx
		}
	}

	return res
}
