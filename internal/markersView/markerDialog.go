package markersview

import (
	"fmt"
	"slices"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/spyhere/re-peat/internal/audio"
	"github.com/spyhere/re-peat/internal/common"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
)

func newMarkerDialog() *markerDialog {
	return &markerDialog{
		nameField: &common.Inputable{},
		timeField: &common.Inputable{},
		tagsField: &common.Inputable{},
	}
}

type markerDialog struct {
	*tm.TimeMarker
	tags      []string
	nameField *common.Inputable
	timeField *common.Inputable
	tagsField *common.Inputable
}

func (m *markerDialog) prepareForOpening(curMarker *tm.TimeMarker, a audio.Audio) {
	m.TimeMarker = curMarker
	m.nameField.SetText(curMarker.Name)
	formattedSeconds := common.FormatSeconds(a.GetSecondsFromPCM(curMarker.Pcm))
	m.timeField.SetText(formattedSeconds)
	m.tags = slices.Clone(curMarker.CategoryTags)
}

func (m *markerDialog) executeConfirm(a audio.Audio) {
	seconds, err := common.ParseSeconds(m.timeField.GetInput())
	if err != nil {
		fmt.Println(err)
		return
	}
	seconds = min(a.Seconds, seconds)
	m.TimeMarker.Name = m.nameField.GetInput()
	m.TimeMarker.Pcm = a.GetPcmFromSeconds(seconds)
	m.TimeMarker.CategoryTags = m.tags
	m.TimeMarker = nil
}

func (m *markerDialog) handleFieldsEvents(gtx layout.Context) {
	if m.nameField.HasSubmit() {
		m.nameField.Blur(gtx)
	}
	if m.timeField.HasSubmit() {
		m.timeField.Blur(gtx)
	}
	if m.tagsField.HasSubmit() {
		m.tags = append(m.tags, m.tagsField.GetInput())
		m.tagsField.SetText("")
	}
	if len(m.tags) > 0 && m.tagsField.HasEmptyDeleteEvent() {
		m.tags = m.tags[:len(m.tags)-1]
	}
}

func (m *markerDialog) getCursorType() (pointer.Cursor, bool) {
	if m.nameField.Hovered() || m.timeField.Hovered() || m.tagsField.Hovered() {
		return pointer.CursorPointer, true
	}
	return pointer.CursorDefault, false
}
