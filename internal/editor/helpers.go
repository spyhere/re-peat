package editor

import (
	"cmp"
	"encoding/binary"
	"fmt"
	"image"
	"log"
	"math"
	"strconv"
	"strings"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
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

func prcToPx(origin int, prc string) int {
	prc = strings.Split(prc, "%")[0]
	prcInt, err := strconv.Atoi(prc)
	if err != nil {
		log.Fatal(err)
	}
	return prcInt * origin / 100
}

func handlePointerEvents(gtx layout.Context, tag event.Tag, pKind pointer.Kind, cb func(e pointer.Event)) {
	for {
		evt, ok := gtx.Event(pointer.Filter{
			Target:  tag,
			Kinds:   pKind,
			ScrollX: pointer.ScrollRange{Min: -1e9, Max: 1e9},
			ScrollY: pointer.ScrollRange{Min: -1e9, Max: 1e9},
		})
		if !ok {
			break
		}
		e, ok := evt.(pointer.Event)
		if !ok {
			continue
		}
		cb(e)
	}
}

func handleKeyEvents(gtx layout.Context, cb func(e key.Event), filters ...event.Filter) {
	for {
		evt, ok := gtx.Event(filters...)
		if !ok {
			break
		}
		e, ok := evt.(key.Event)
		if !ok {
			continue
		}
		cb(e)
	}
}

func snap(v float32) float32 {
	return float32(math.Round(float64(v)))
}

func registerTag(gtx layout.Context, tag event.Tag, area image.Rectangle) {
	defer clip.Rect(area).Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, tag)
}

func makeMacro(ops *op.Ops, cb func()) op.CallOp {
	macro := op.Record(ops)
	cb()
	return macro.Stop()
}

func strlen(input string) int {
	return strings.Count(input, "") - 1
}

// Since non-lating letters are taking more then 1 byte `strlen` and manual idx in range is required
func truncName(name string, limit int) string {
	if limit == 0 || strlen(name) < limit {
		return name
	}
	var newName strings.Builder
	idx := 0
	for _, r := range name {
		if idx >= limit-3 {
			break
		}
		newName.WriteRune(r)
		idx++
	}
	newName.WriteString("...")
	return newName.String()
}
