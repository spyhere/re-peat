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
	arr      []*marker
	editing  *marker
	hovering *marker
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
	flag    bool
	pole    bool
	label   bool
	hovered bool
}

func (m *markers) newMarker(pcm int64) {
	if len(m.arr)+1 > markersLimit {
		// TODO: display error
		return
	}
	// TODO: Remove placeholder name
	newM := &marker{
		pcm: pcm,
		tags: &markerTags{
			flag:  &struct{}{},
			pole:  &struct{}{},
			label: &struct{}{},
		},
	}
	m.arr = append(m.arr, newM)
	m.editing = newM
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

func (m *markers) startEdit(curMarker *marker) {
	m.editing = curMarker
}

func (m *markers) stopEdit() {
	m.editing = nil
}

func (m *markers) isEditing() bool {
	return m.editing != nil
}

func (m *markers) startHover(curMarker *marker) {
	if m.hovering == nil {
		m.hovering = curMarker
	}
}

func (m *markers) stopHover() {
	m.hovering = nil
}

func (m *markers) isHovering() bool {
	return m.hovering != nil
}
