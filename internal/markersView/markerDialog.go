package markersview

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/spyhere/re-peat/internal/audio"
	"github.com/spyhere/re-peat/internal/common"
	"github.com/spyhere/re-peat/internal/i18n"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

func newMarkerDialog(tagLimit int, th *theme.RepeatTheme, a audio.AudioMeta) markerDialog {
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
	a          audio.AudioMeta
	tags       []string
	allTags    []string
	tagOptions []string
	nameField  *common.Inputable
	timeField  *common.Inputable
	tagsField  *common.Comboboxable
	focuser    *common.FocusManager
	th         *theme.RepeatTheme
	i18n       i18n.State
}

// NOTE: Do we need to pass audioMeta here?
func (m *markerDialog) prepareForOpening(i18n i18n.State, a audio.AudioMeta, curMarker *tm.TimeMarker, allChips map[string]struct{}) {
	m.i18n = i18n
	m.a = a
	m.allTags = m.allTags[:0]
	for chipName := range allChips {
		m.allTags = append(m.allTags, chipName)
	}
	slices.Sort(m.allTags)

	m.TimeMarker = curMarker
	m.nameField.SetText(curMarker.Name)
	formattedSeconds := common.FormatSeconds(m.a.GetSecondsFromSamples(curMarker.Samples))
	m.timeField.SetText(formattedSeconds)
	m.timeField.OnBlur(m.normalizeTimeInput)
	m.timeField.SetSanitizer(m.sanitizeTimeInput)
	m.tags = slices.Clone(curMarker.CategoryTags)
	m.tagsField.SetText("")
}

func (m *markerDialog) executeConfirm(a audio.AudioMeta) {
	seconds, err := common.ParseSeconds(m.timeField.Text())
	if err != nil {
		seconds = 0
	}
	seconds = min(m.a.Seconds, seconds)
	m.TimeMarker.Name = m.nameField.Text()
	m.TimeMarker.Samples = a.GetSamplesFromSeconds(seconds)
	if m.tagsField.GetInput() != "" {
		m.handleTagsFieldNewChip()
	}
	m.TimeMarker.CategoryTags = m.tags
	m.TimeMarker = nil
	m.focuser.RequestBlur(nil)
}

func (m *markerDialog) cancelCreate() {
	m.focuser.RequestBlur(nil)
	m.TimeMarker = nil
}

func (m *markerDialog) cancelEdit() {
	m.focuser.RequestBlur(nil)
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
	if v == "" {
		return
	}
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
	maxSeconds := min(float64(minutes*60+seconds), m.a.Seconds)
	m.timeField.SetText(common.FormatSeconds(maxSeconds))
}

func (m *markerDialog) handleTagsFieldNewChip() {
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

func (m *markerDialog) handleFieldsEvents() {
	if m.nameField.HasSubmit() {
		m.focuser.RequestFocus(m.timeField)
		return
	}
	if m.timeField.HasSubmit() {
		m.focuser.RequestFocus(m.tagsField)
		return
	}
	if m.tagsField.HasSubmit() {
		m.handleTagsFieldNewChip()
		return
	}
	if v, ok := m.tagsField.HasSelectedValue(); ok {
		m.tags = append(m.tags, v)
		m.tagsField.SetText("")
		m.tagOptions = m.tagOptions[:0]
		return
	}
	if remIdx := m.tagsField.HasRemovedChip(); remIdx >= 0 {
		copy(m.tags[remIdx:], m.tags[remIdx+1:])
		m.tags = m.tags[:len(m.tags)-1]
		return
	}
	if m.tagsField.HasEmptyDeleteEvent() && len(m.tags) > 0 {
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

func (m *markerDialog) Layout(gtx layout.Context, totalSeconds float64) layout.Dimensions {
	if cursor, ok := m.getCursorType(); ok {
		common.SetCursor(gtx, cursor)
	}
	m.handleFieldsEvents()
	s := defaultFieldGroupStyle()
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
							LabelText: m.i18n.Generic.Name,
						},
						Inputable:   m.nameField,
						MaxLen:      20,
						Placeholder: m.i18n.Markers.MNamePlaceholder,
					})
					inputDims.Size.Y += gapPx
					return inputDims
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Max.X = fieldW
					inputDims := common.DrawInputField(gtx, m.th, common.InputFieldProps{
						Base: common.InputFieldBase{
							LabelText: m.i18n.Generic.Time,
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
							LabelText: m.i18n.Generic.Tags,
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
