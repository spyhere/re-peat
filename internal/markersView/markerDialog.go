package markersview

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/spyhere/re-peat/internal/audio"
	"github.com/spyhere/re-peat/internal/common"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

func newMarkerDialog(tagLimit int, th *theme.RepeatTheme, a audio.Audio) markerDialog {
	fm := &common.FocusManager{}
	return markerDialog{
		a:          a,
		nameField:  &common.Inputable{Focuser: fm},
		timeField:  &common.Inputable{Focuser: fm},
		tagsField:  new(common.Comboboxable).WithFocusManager(fm),
		allTags:    make([]string, tagLimit),
		tags:       make([]string, tagLimit),
		tagOptions: make([]string, 0, tagLimit),
		focuser:    fm,
		th:         th,
	}
}

type markerDialog struct {
	*tm.TimeMarker
	a          audio.Audio
	tags       []string
	allTags    []string
	tagOptions []string
	nameField  *common.Inputable
	timeField  *common.Inputable
	tagsField  *common.Comboboxable
	focuser    *common.FocusManager
	th         *theme.RepeatTheme
}

func (m *markerDialog) prepareForOpening(curMarker *tm.TimeMarker, allChips map[string]struct{}) {
	m.allTags = m.allTags[:0]
	for chipName := range allChips {
		m.allTags = append(m.allTags, chipName)
	}
	slices.Sort(m.allTags)

	m.TimeMarker = curMarker
	m.nameField.SetText(curMarker.Name)
	formattedSeconds := common.FormatSeconds(m.a.GetSecondsFromPCM(curMarker.Pcm))
	m.timeField.SetText(formattedSeconds)
	m.timeField.OnBlur(m.normalizeTimeInput)
	m.timeField.SetSanitizer(m.sanitizeTimeInput)
	m.tags = slices.Clone(curMarker.CategoryTags)
	m.tagsField.SetText("")
}

func (m *markerDialog) executeConfirm(a audio.Audio) {
	seconds, err := common.ParseSeconds(m.timeField.Text())
	if err != nil {
		fmt.Println(err)
		return
	}
	seconds = min(a.Seconds, seconds)
	m.TimeMarker.Name = m.nameField.Text()
	m.TimeMarker.Pcm = a.GetPcmFromSeconds(seconds)
	m.TimeMarker.CategoryTags = m.tags
	m.TimeMarker = nil
}

func (m *markerDialog) sanitizeTimeInput(input string) string {
	var b strings.Builder
	isFirstRune, colonSeen := true, false
	for _, r := range input {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else if r == ':' && !colonSeen && !isFirstRune {
			b.WriteRune(r)
			colonSeen = true
		}
		isFirstRune = false
	}
	return b.String()
}
func (m *markerDialog) normalizeTimeInput() {
	var minutes, seconds int
	var err error
	defer func() {
		if err != nil {
			// TODO: display validation error
			fmt.Println(err)
		}
	}()

	v := m.timeField.GetInput()
	if !strings.Contains(v, ":") {
		seconds, err = strconv.Atoi(v)
		if err != nil {
			return
		}
	} else {
		parts := strings.SplitN(v, ":", 2)
		if len(parts) != 2 {
			err = fmt.Errorf("Invalid time format")
			return
		}
		minutesStr, secondsStr := parts[0], parts[1]
		minutes, err = strconv.Atoi(minutesStr)
		if err != nil {
			return
		}
		seconds, err = strconv.Atoi(secondsStr)
		if err != nil {
			return
		}
	}
	maxSeconds := min(minutes*60+seconds, int(m.a.Seconds))
	minutes = maxSeconds / 60
	seconds = maxSeconds % 60
	m.timeField.SetText(fmt.Sprintf("%02d:%02d", minutes, seconds))
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
	if v, ok := m.tagsField.HasSelectedValue(); ok {
		m.tags = append(m.tags, v)
		m.tagsField.SetText("")
		m.tagOptions = m.tagOptions[:0]
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

func (m *markerDialog) getTagOptions() []string {
	isDirty := m.tagsField.IsDirty()
	if !isDirty && len(m.tagOptions) > 0 {
		return m.tagOptions
	}
	input := m.tagsField.GetInput()
	m.tagOptions = m.tagOptions[:0]
	if len(input) <= suggestionsThreshold {
		return m.tagOptions
	}

	inputLower := strings.ToLower(input)
	for _, chipName := range m.allTags {
		suchTagExists := slices.Contains(m.tags, chipName)
		if !suchTagExists && strings.HasPrefix(strings.ToLower(chipName), inputLower) {
			m.tagOptions = append(m.tagOptions, chipName)
		}
	}
	return m.tagOptions
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
						Filter:      "1234567890:",
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
						OptionsF:     m.getTagOptions,
					})
				}),
			)
		})
	})
	m.focuser.PlaceScrim(gtx)
	return dims
}
