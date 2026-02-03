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
}

type marker struct {
	pcm    int64
	name   string
	tags   *markerTags
	isDead bool
}
type markerTags struct {
	flag  *struct{}
	pole  *struct{}
	label *struct{}
}
type mInteraction struct {
	flag  bool
	pole  bool
	label bool
}

func (m *markers) newMarker(pcm int64) {
	if len(m.arr)+1 > markersLimit {
		// TODO: display error
		return
	}
	// TODO: Remove placeholder name
	name := "Chorus"
	m.arr = append(m.arr, &marker{
		pcm:  pcm,
		name: name,
		tags: &markerTags{
			flag:  &struct{}{},
			pole:  &struct{}{},
			label: &struct{}{},
		},
	})
}

func (m *marker) markDead() {
	m.isDead = true
}

func (m *markers) deleteDead() {
	m.arr = slices.DeleteFunc(m.arr, func(it *marker) bool {
		return it.isDead
	})
}

func (m *markers) sortCb(a, b *marker) int {
	return int(b.pcm - a.pcm)
}

func (m *markers) getSortedMarkers() []*marker {
	if slices.IsSortedFunc(m.arr, m.sortCb) {
		return m.arr
	}
	seq := slices.Values(m.arr)
	return slices.SortedStableFunc(seq, m.sortCb)
}
