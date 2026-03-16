package markersview

import (
	"fmt"
	"slices"
	"unicode"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/spyhere/re-peat/internal/audio"
	"github.com/spyhere/re-peat/internal/common"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

func newMarkerDialog(tagLimit int, th *theme.RepeatTheme) markerDialog {
	fm := &common.FocusManager{}
	return markerDialog{
		nameField:     &common.Inputable{Focuser: fm},
		timeField:     &common.Inputable{Focuser: fm},
		tagsField:     new(common.Comboboxable).WithFocusManager(fm),
		tagOptionsMap: make(map[string]common.ComboboxOption, tagLimit),
		tagOptions:    make([]common.ComboboxOption, 0, tagLimit),
		focuser:       fm,
		th:            th,
	}
}

type markerDialog struct {
	*tm.TimeMarker
	tags          []string
	tagOptionsMap map[string]common.ComboboxOption
	tagOptions    []common.ComboboxOption
	nameField     *common.Inputable
	timeField     *common.Inputable
	tagsField     *common.Comboboxable
	focuser       *common.FocusManager
	th            *theme.RepeatTheme
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

func (m *markerDialog) blur() {
	m.focuser.RequestBlur()
}

func (m *markerDialog) handleFieldsEvents() {
	if m.nameField.HasSubmit() {
		m.focuser.RequestFocus(m.timeField)
	}
	if m.timeField.HasSubmit() {
		m.focuser.RequestFocus(m.tagsField)
	}
	if m.tagsField.HasSubmit() {
		newTag := m.tagsField.GetInput()
		if newTag == "" {
			return
		}
		runes := []rune(newTag)
		runes[0] = unicode.ToUpper(runes[0])
		newTag = string(runes)
		suchTagExists := slices.ContainsFunc(m.tags, func(it string) bool {
			return it == newTag
		})
		if suchTagExists {
			return
		}
		m.tags = append(m.tags, newTag)
		m.tagsField.SetText("")
	}
	if len(m.tags) > 0 && m.tagsField.HasEmptyDeleteEvent() {
		m.tags = m.tags[:len(m.tags)-1]
	}
}

func (m *markerDialog) getCursorType() (pointer.Cursor, bool) {
	if m.nameField.IsHovered() || m.timeField.IsHovered() || m.tagsField.IsHovered() {
		return pointer.CursorText, true
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

type drawMarkerDialogSizeSpecs struct {
	fieldsYMargin unit.Dp
	fieldsXMargin unit.Dp
	fieldW        unit.Dp
	gap           unit.Dp
}

var drawMarkerDialogSpecs = drawMarkerDialogSizeSpecs{
	fieldsYMargin: 10,
	fieldsXMargin: 10,
	fieldW:        270,
	gap:           20,
}

func (m *markerDialog) Layout(gtx layout.Context, totalSeconds float64) layout.Dimensions {
	s := drawMarkerDialogSpecs
	inset := layout.Inset{Top: s.fieldsYMargin, Bottom: s.fieldsYMargin, Left: s.fieldsXMargin, Right: s.fieldsXMargin}
	dims := inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gapPx := gtx.Dp(s.gap)
		gtx.Constraints.Max.X = gtx.Constraints.Min.X
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		fieldW := gtx.Dp(s.fieldW)
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Max.X = fieldW
					inputDims := common.DrawInputField(gtx, m.th, common.InputFieldProps{
						Base: common.InputFieldBase{
							LabelText: "Имя",
						},
						Inputable:   m.nameField,
						MaxLen:      20,
						Placeholder: "Новый маркер...",
					})
					inputDims.Size.Y += gapPx
					return inputDims
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Max.X = fieldW
					inputDims := common.DrawInputField(gtx, m.th, common.InputFieldProps{
						Base: common.InputFieldBase{
							LabelText: "Время",
						},
						Inputable:   m.timeField,
						MaxLen:      7,
						Placeholder: common.FormatSeconds(totalSeconds),
					})
					inputDims.Size.Y += gapPx
					return inputDims
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Max.X = fieldW
					return common.DrawCombobox(gtx, m.th, common.ComboboxProps{
						Base: common.InputFieldBase{
							LabelText: "Категории",
						},
						Comboboxable: m.tagsField,
						Chips:        m.tags,
						MaxLen:       20,
						Options:      m.getTagOptions(),
					})
				}),
			)
		})
	})
	m.focuser.PlaceScrim(gtx)
	return dims
}
