package markersview

import (
	"strings"

	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/spyhere/re-peat/internal/audio"
	"github.com/spyhere/re-peat/internal/common"
	p "github.com/spyhere/re-peat/internal/player"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

type Props struct {
	Audio       audio.Audio
	Th          *theme.RepeatTheme
	TimeMarkers *tm.TimeMarkers
	Player      *p.Player
}

func NewMarkersView(props Props) *MarkersView {
	mView := &MarkersView{
		audio:        props.Audio,
		th:           props.Th,
		timeMarkers:  props.TimeMarkers,
		p:            props.Player,
		searchable:   &common.Searchable{},
		replayButton: &widget.Clickable{},
		tagButton:    &widget.Clickable{},
		deleteButton: &widget.Clickable{},
	}
	table := common.NewTable(common.TableProps[*tm.TimeMarker]{
		Axis:      layout.Vertical,
		ColumsNum: 7,
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

type MarkersView struct {
	p            *p.Player
	timeMarkers  *tm.TimeMarkers
	markerPlayed *tm.TimeMarker
	th           *theme.RepeatTheme
	table        *common.Table[*tm.TimeMarker]
	searchable   *common.Searchable
	replayButton *widget.Clickable
	tagButton    *widget.Clickable
	deleteButton *widget.Clickable
	audio        audio.Audio
}

// TODO: Move played marker to app state
func (m *MarkersView) toggleMarker(curMarker *tm.TimeMarker) {
	if m.markerPlayed == curMarker {
		m.pausePlaying()
		return
	}
	m.startPlaying(curMarker)
}

func (m *MarkersView) startPlaying(curMarker *tm.TimeMarker) {
	m.markerPlayed = curMarker
	m.p.Set(curMarker.Pcm)
	m.p.Play()
}

func (m *MarkersView) pausePlaying() {
	m.markerPlayed = nil
	m.p.Pause()
}

func (m *MarkersView) isThisMarkerPlaying(curMarker *tm.TimeMarker) bool {
	return m.markerPlayed == curMarker
}

func (m *MarkersView) updateDefferedState() {
	m.timeMarkers.DeleteDead()
}

func (m *MarkersView) getTableRowValue(rowIdx int) *tm.TimeMarker {
	return m.timeMarkers.GetAsc(rowIdx)
}

func (m *MarkersView) tableRowFilter(curMarker *tm.TimeMarker) bool {
	return strings.Contains(
		strings.ToLower(curMarker.Name),
		strings.ToLower(m.searchable.GetInput()),
	)
}

func (m *MarkersView) replayMarkers() {
	m.p.Set(0)
	m.p.Play()
}

func (m *MarkersView) openTagsFilter() {
	//
}

func (m *MarkersView) deleteMarkers() {
	m.timeMarkers.MarkAllDead()
}
