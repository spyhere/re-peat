package markersview

import (
	"github.com/spyhere/re-peat/internal/audio"
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
	return &MarkersView{
		audio:       props.Audio,
		th:          props.Th,
		timeMarkers: props.TimeMarkers,
		p:           props.Player,
	}
}

type MarkersView struct {
	p            *p.Player
	timeMarkers  *tm.TimeMarkers
	markerPlayed *tm.TimeMarker
	th           *theme.RepeatTheme
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
	m.p.Pause()
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
