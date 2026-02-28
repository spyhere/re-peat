package editorview

import (
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
)

func newMarkers(tmArray *tm.TimeMarkers) *markers {
	return &markers{
		arr: tmArray,
	}
}

type markers struct {
	arr           *tm.TimeMarkers
	editing       *tm.TimeMarker
	hovering      *tm.TimeMarker
	overlayParams markerProps
}

type mInteraction struct {
	flag    bool
	pole    bool
	label   bool
	hovered bool
}

func (m *markers) newMarker(pcm int64) {
	newM := m.arr.NewMarker(pcm)
	if newM == nil {
		return
	}
	m.editing = newM
}

func (m *markers) deleteDead() {
	m.arr.DeleteDead()
}

func (m *markers) getSortedMarkers() tm.TimeMarkers {
	return m.arr.Sorted()
}

func (m *markers) startEdit(curMarker *tm.TimeMarker) {
	m.editing = curMarker
}

func (m *markers) stopEdit() {
	m.editing = nil
}

func (m *markers) isEditing() bool {
	return m.editing != nil
}

func (m *markers) startHover(curMarker *tm.TimeMarker) {
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
