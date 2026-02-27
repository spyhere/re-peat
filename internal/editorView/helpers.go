package editorview

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
		// Normalize to -1..1
		normalised[i/2] = float32(sample) / maxUin16
	}
	return normalised, nil
}

func makeSamplesMono(samples []float32, chanNum int) []float32 {
	if chanNum == 1 {
		return samples
	}
	if chanNum > 2 {
		fmt.Println("Not supported more than 2 channels")
		return []float32{}
	}
	res := make([]float32, len(samples)/chanNum)

	for i := 0; i < len(samples); i += 2 {
		lSample := samples[i]
		rSample := samples[i+1]
		res[i/2] = (lSample + rSample) * 0.5
	}
	return res
}

func populateCache(cache map[int][][2]float32, samples []float32, workers []*cacheWorker) {
	for _, it := range samples {
		for _, w := range workers {
			if it < w.min {
				w.min = it
			}
			if it > w.max {
				w.max = it
			}
			w.count--
			if w.count == 0 {
				cache[w.samplesPerPx][w.sliceIdx] = [2]float32{w.min, w.max}
				w.sliceIdx++
				w.min = 1
				w.max = -1
				w.count = w.samplesPerPx
			}
		}
	}
}

func reducePeaks(data [][2]float32) (low float32, high float32) {
	if len(data) == 0 {
		return 0, 0
	}
	low = 1
	high = -1
	for _, it := range data {
		if it[0] < low {
			low = it[0]
		}
		if it[1] > high {
			high = it[1]
		}
	}
	return low, high
}
