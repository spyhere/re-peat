package markersview

import (
	"fmt"
	"slices"
	"unicode"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/spyhere/re-peat/internal/audio"
	"github.com/spyhere/re-peat/internal/common"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
)

func newMarkerDialog(tagLimit int) markerDialog {
	return markerDialog{
		nameField:     &common.Inputable{},
		timeField:     &common.Inputable{},
		tagsField:     &common.Inputable{},
		tagOptionsMap: make(map[string]common.ComboboxOption, tagLimit),
		tagOptions:    make([]common.ComboboxOption, 0, tagLimit),
	}
}

type markerDialog struct {
	*tm.TimeMarker
	tags          []string
	tagOptionsMap map[string]common.ComboboxOption
	tagOptions    []common.ComboboxOption
	nameField     *common.Inputable
	timeField     *common.Inputable
	tagsField     *common.Inputable
}

func (m *markerDialog) prepareForOpening(curMarker *tm.TimeMarker, a audio.Audio, allChips map[string]struct{}) {
	for chipName := range allChips {
		if _, ok := m.tagOptionsMap[chipName]; !ok {
			m.tagOptionsMap[chipName] = common.ComboboxOption{
				Text: chipName,
				Cl:   &widget.Clickable{},
			}
		}
	}

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
		newTag := m.tagsField.GetInput()
		runes := []rune(newTag)
		runes[0] = unicode.ToUpper(runes[0])
		newTag = string(runes)
		m.tags = append(m.tags, newTag)
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

const suggestionsThreshold = 2

func (m *markerDialog) getTagOptions() []common.ComboboxOption {
	// Makes no sense unless dropdown situation is resolved
	return m.tagOptions
	// input := m.tagsField.GetInput()
	// m.tagOptions = m.tagOptions[:0]
	// if len(input) <= suggestionsThreshold {
	// 	return m.tagOptions
	// }
	// for chipName, chipOption := range m.tagOptionsMap {
	// 	if strings.Contains(strings.ToLower(chipName), strings.ToLower(input)) {
	// 		m.tagOptions = append(m.tagOptions, chipOption)
	// 	}
	// }
	// return m.tagOptions
}
