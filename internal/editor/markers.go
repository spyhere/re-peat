package editor

import (
	"slices"
)

const markersLimit = 100

func newMarkers() *markers {
	return &markers{
		arr: make([]*marker, 0, markersLimit),
	}
}

type markers struct {
	arr   []*marker
	idx   int8
	draft struct {
		isVisible bool
		pointerX  float32
	}
}

type marker struct {
	Samples int
	Name    string
	Tag     struct{}
}

func (m *markers) enableDraft(x float32) {
	m.draft.isVisible = true
	m.draft.pointerX = x
}

func (m *markers) disableDraft() {
	m.draft.isVisible = false
}

func (m *markers) isDraftVisible() bool {
	return m.draft.isVisible
}

func (m *markers) newMarker(samples int) {
	if m.idx+1 > markersLimit {
		// TODO: display error
		return
	}
	// TODO: Remove placeholder name
	name := "Chorus"
	if m.idx == 7 {
		name = "Eugene"
	} else {
		name = "Chorus"
	}
	m.arr = append(m.arr, &marker{Samples: samples, Name: name})
	m.idx++
}

func (m *markers) sortCb(a, b *marker) int {
	return b.Samples - a.Samples
}

func (m *markers) getSortedMarkers() []*marker {
	if slices.IsSortedFunc(m.arr, m.sortCb) {
		return m.arr
	}
	seq := slices.Values(m.arr)
	return slices.SortedStableFunc(seq, m.sortCb)
}
