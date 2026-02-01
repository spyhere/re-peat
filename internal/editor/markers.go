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
	arr []*marker
	idx int8
}

type marker struct {
	Pcm  int64
	Name string
	Tag  *struct{}
}

func (m *markers) newMarker(pcm int64) {
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
	m.arr = append(m.arr, &marker{
		Pcm:  pcm,
		Name: name,
		Tag:  &struct{}{},
	})
	m.idx++
}

func (m *markers) sortCb(a, b *marker) int {
	return int(b.Pcm - a.Pcm)
}

func (m *markers) getSortedMarkers() []*marker {
	if slices.IsSortedFunc(m.arr, m.sortCb) {
		return m.arr
	}
	seq := slices.Values(m.arr)
	return slices.SortedStableFunc(seq, m.sortCb)
}
