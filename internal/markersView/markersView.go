package markersview

import (
	"slices"
	"strings"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/spyhere/re-peat/internal/audio"
	"github.com/spyhere/re-peat/internal/common"
	p "github.com/spyhere/re-peat/internal/player"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

const (
	selectionRuneLimit = 3
	globalChipsLimit   = 100
)

type Props struct {
	Audio       audio.Audio
	Th          *theme.RepeatTheme
	TimeMarkers *tm.TimeMarkers
	Player      *p.Player
	Dialog      *common.Dialog
}

func NewMarkersView(props Props) *MarkersView {
	fm := &common.FocusManager{}
	mView := &MarkersView{
		audio:         props.Audio,
		th:            props.Th,
		timeMarkers:   props.TimeMarkers,
		p:             props.Player,
		hotKeyBuf:     make([]rune, 0, selectionRuneLimit),
		dialog:        props.Dialog,
		searchbar:     &common.Inputable{Focuser: fm},
		fm:            fm,
		enabledTagsLs: &widget.List{},
		markerDialog:  newMarkerDialog(globalChipsLimit, props.Th, props.Audio),
		tagsDialog:    newTagsDialog(globalChipsLimit),
		chipsFilter:   newChipsFilter(globalChipsLimit),
		commentDialog: newCommentDialog(props.Th),
	}
	table := common.NewTable(common.TableProps[*tm.TimeMarker]{
		Axis: layout.Vertical,
		HeaderCellsAlignment: []layout.Direction{
			layout.Center,
			layout.Center,
			layout.W,
			layout.Center,
			layout.W,
			layout.Center,
			layout.Center,
		},
		RowCellsAlignment: []layout.Direction{
			layout.Center,
			layout.Center,
			layout.W,
			layout.Center,
			layout.W,
			layout.Center,
			layout.Center,
		},
		RowValueCb:  mView.getTableRowValue,
		RowFilterCb: mView.tableRowFilter,
	})
	mView.table = table
	return mView
}

type dialogOwner uint8

const (
	none dialogOwner = iota
	create
	comment
	edit
	tagFilter
	deleteAll
)

type MarkersView struct {
	p             *p.Player
	timeMarkers   *tm.TimeMarkers
	markerInPlay  *tm.TimeMarker
	th            *theme.RepeatTheme
	table         *common.Table[*tm.TimeMarker]
	searchbar     *common.Inputable
	fm            *common.FocusManager
	replayCl      widget.Clickable
	tagCl         widget.Clickable
	tagClearCl    widget.Clickable
	enabledTagsLs *widget.List
	createCl      widget.Clickable
	disabledCl    widget.Clickable
	deleteCl      widget.Clickable
	dialog        *common.Dialog
	dialogOwner
	markerDialog
	tagsDialog
	commentDialog commentDialog
	chipsFilter
	hotKeyBuf []rune
	audio     audio.Audio
}

func (m *MarkersView) togglePlayer(curMarker *tm.TimeMarker) {
	if m.markerInPlay == nil {
		m.startPlaying(curMarker)
	} else {
		m.pausePlaying()
	}
}

func (m *MarkersView) toggleMarker(curMarker *tm.TimeMarker) {
	if m.markerInPlay == curMarker {
		m.pausePlaying()
		return
	}
	m.startPlaying(curMarker)
}

func (m *MarkersView) startPlaying(curMarker *tm.TimeMarker) {
	m.markerInPlay = curMarker
	m.p.Set(curMarker.Pcm)
	m.p.Play()
}

func (m *MarkersView) pausePlaying() {
	m.markerInPlay = nil
	m.p.Pause()
}

func (m *MarkersView) isThisMarkerPlaying(curMarker *tm.TimeMarker) bool {
	return m.markerInPlay == curMarker
}

func (m *MarkersView) updateDefferedState() {
	m.timeMarkers.DeleteDead()
}

func (m *MarkersView) getTableRowValue(rowIdx int) *tm.TimeMarker {
	return m.timeMarkers.Get(rowIdx, true)
}

func (m *MarkersView) tableRowFilter(curMarker *tm.TimeMarker) bool {
	hasChipsMatch := true
	for _, chip := range m.chipsFilter.getEnabledChips() {
		hasChipsMatch = false
		if slices.Contains(curMarker.CategoryTags, chip) {
			hasChipsMatch = true
			break
		}
	}
	return hasChipsMatch && strings.Contains(
		strings.ToLower(curMarker.Name),
		strings.ToLower(m.searchbar.GetInput()),
	)
}

func (m *MarkersView) replayMarkers() {
	if m.p.IsPlaying() {
		m.p.Pause()
	} else {
		m.p.Set(0)
		m.p.Play()
	}
}

func (m *MarkersView) deleteMarkers() {
	m.timeMarkers.MarkAllDead()
}

func (m *MarkersView) listenToPlayerUpdates() {
	playerPos := m.p.GetReadAmount()
	if m.markerInPlay != nil && playerPos < m.markerInPlay.Pcm {
		// time markers were dragged in EditorView, so MarkersView should be updated as well
		var prev *tm.TimeMarker
		for _, it := range *m.timeMarkers {
			if it.Pcm > playerPos {
				m.markerInPlay = prev
				return
			}
			prev = it
		}
	}

	nextMarker := m.timeMarkers.Get(m.timeMarkers.GetIndex(m.markerInPlay, true)+1, true)
	// Next marker can be nil when there are no markers, or current is the last one
	if nextMarker == nil {
		return
	}
	if playerPos >= nextMarker.Pcm {
		m.markerInPlay = nextMarker
	}
}

func (m *MarkersView) clearHotKeyBuf() {
	m.hotKeyBuf = m.hotKeyBuf[:0]
}

func (m *MarkersView) cancelDialog() {
	switch m.dialogOwner {
	case create:
		m.markerDialog.cancelCreate()
	case comment:
		m.commentDialog.cancelComment()
	case edit:
		m.markerDialog.cancelEdit()
	}
	m.dialog.Hide()
	m.dialogOwner = none
}

func (m *MarkersView) confirmDialog() {
	switch m.dialogOwner {
	case create:
		m.confirmEdit()
	case comment:
		m.confirmComment()
	case edit:
		m.confirmEdit()
	case tagFilter:
		m.confirmTagFilter()
	case deleteAll:
		m.confirmDeleteAll()
	}
	m.dialog.Hide()
	m.dialogOwner = none
}

func (m *MarkersView) handleAddMarkerButton(gtx layout.Context) {
	if m.createCl.Clicked(gtx) {
		m.openMarkerDialog(m.timeMarkers.NewMarker(0), create, "Create Marker")
	}
	if m.createCl.Hovered() {
		common.SetCursor(gtx, pointer.CursorPointer)
	}
}
