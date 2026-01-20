package editor

import (
	"cmp"
)

// Lock "this" between "from" and "to"
func clamp[T cmp.Ordered](from T, this T, to T) T {
	return min(max(from, this), to)
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
