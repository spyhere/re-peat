package editor

import (
	"slices"
)

const markersLimit = 100

type markers struct {
	arr   []marker
	idx   int8
	draft struct {
		isPointerInside bool
		isVisible       bool
		pointerX        float32
	}
}

type marker struct {
	Samples int
	Name    string
}

func (m *markers) NewMarker(samples int) {
	if m.idx+1 > markersLimit {
		// TODO: display error
		return
	}
	// TODO: Remove placeholder name
	m.arr = append(m.arr, marker{Samples: samples, Name: "Chorus"})
	m.idx++
	slices.SortStableFunc(m.arr, func(a, b marker) int {
		return a.Samples - b.Samples
	})
}
