package markersview

import (
	"github.com/spyhere/re-peat/internal/audio"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

type Props struct {
	Audio       audio.Audio
	Th          *theme.RepeatTheme
	TimeMarkers *tm.TimeMarkers
}

func NewMarkersView(props Props) *MarkersView {
	return &MarkersView{
		audio:       props.Audio,
		th:          props.Th,
		timeMarkers: props.TimeMarkers,
	}
}

type MarkersView struct {
	audio       audio.Audio
	th          *theme.RepeatTheme
	timeMarkers *tm.TimeMarkers
}
