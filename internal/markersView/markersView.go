package markersview

import (
	"slices"
	"strings"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/spyhere/re-peat/internal/common"
	"github.com/spyhere/re-peat/internal/state"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
)

const (
	selectionRuneLimit = 3
	globalChipsLimit   = 100
)

type Props struct {
	TimeMarkers *tm.TimeMarkers
	State       *state.AppState
}

func NewMarkersView(props Props) MarkersView {
	fm := &common.FocusManager{}
	mView := MarkersView{
		AppState:      props.State,
		hotKeyBuf:     make([]rune, 0, selectionRuneLimit),
		searchbar:     &common.Inputable{Focuser: fm},
		fm:            fm,
		enabledTagsLs: &widget.List{},
		markerDialog:  newMarkerDialog(globalChipsLimit, props.State.Th, props.State.AudioMeta),
		tagsDialog:    newTagsDialog(globalChipsLimit),
		commentDialog: newCommentDialog(props.State.Th),
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
	*state.AppState
	draftMarker   tm.TimeMarker
	markerInPlay  *tm.TimeMarker
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
	dialogOwner   dialogOwner
	markerDialog  markerDialog
	tagsDialog    tagsDialog
	commentDialog commentDialog
	hotKeyBuf     []rune
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
	m.Player.Set(curMarker.Samples)
	m.Playhead.Set(curMarker.Samples)
	m.Player.Play()
}

func (m *MarkersView) pausePlaying() {
	m.markerInPlay = nil
	m.Player.Pause()
}

func (m *MarkersView) isThisMarkerPlaying(curMarker *tm.TimeMarker) bool {
	return m.markerInPlay == curMarker
}

func (m *MarkersView) updateDefferedState() {
	if m.TimeMarkers.DeleteDead() {
		m.ChipsFilter.ReconcileEnabled(m.TimeMarkers)
	}
	if m.TimeMarkers.IsEmpty() && m.searchbar.GetInput() != "" {
		m.searchbar.SetText("")
		m.SearchbarV = ""
	}
}

func (m *MarkersView) getTableRowValue(rowIdx int) *tm.TimeMarker {
	return m.AppState.TimeMarkers.Get(rowIdx, true)
}

func (m *MarkersView) tableRowFilter(curMarker *tm.TimeMarker) bool {
	hasChipsMatch := true
	for _, chip := range m.ChipsFilter.GetEnabledChips() {
		hasChipsMatch = false
		if slices.Contains(curMarker.CategoryTags, chip) {
			hasChipsMatch = true
			break
		}
	}
	searchbarV := m.searchbar.GetInput()
	m.SearchbarV = searchbarV
	return hasChipsMatch && strings.Contains(
		strings.ToLower(curMarker.Name),
		strings.ToLower(searchbarV),
	)
}

func (m *MarkersView) replayMarkers() {
	if m.Player.IsPlaying() {
		m.Player.Pause()
	} else {
		m.Player.Set(0)
		m.Player.Play()
		m.Playhead.Set(0)
	}
}

func (m *MarkersView) deleteMarkers() {
	m.TimeMarkers.MarkAllDead()
}

func (m *MarkersView) listenToPlayerUpdates() {
	playerSamples := m.Player.GetReadAmount()
	if m.markerInPlay != nil && playerSamples < m.markerInPlay.Samples {
		// time markers were dragged in EditorView, so MarkersView should be updated as well
		var prev *tm.TimeMarker
		for _, it := range m.TimeMarkers {
			if it.Samples > playerSamples {
				m.markerInPlay = prev
				return
			}
			prev = it
		}
	}

	nextMarker := m.TimeMarkers.Get(m.TimeMarkers.GetIndex(m.markerInPlay, true)+1, true)
	// Next marker can be nil when there are no markers, or current is the last one
	if nextMarker == nil {
		return
	}
	if playerSamples >= nextMarker.Samples {
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
	m.Dialog.Hide()
	m.dialogOwner = none
}

func (m *MarkersView) confirmDialog() {
	switch m.dialogOwner {
	case create:
		m.confirmCreate()
	case comment:
		m.confirmComment()
	case edit:
		m.confirmEdit()
	case tagFilter:
		m.confirmTagFilter()
	case deleteAll:
		m.confirmDeleteAll()
	}
	m.Dialog.Hide()
	m.dialogOwner = none
}

func (m *MarkersView) handleAddMarkerButton(gtx layout.Context) {
	if m.createCl.Clicked(gtx) {
		m.openMarkerDialog(&m.draftMarker, create, "Create Marker")
	}
	if m.createCl.Hovered() {
		common.SetCursor(gtx, pointer.CursorPointer)
	}
}

func (m *MarkersView) isDisabled() bool {
	return !m.HasAudioLoaded() || m.AppState.IsLoading()
}
