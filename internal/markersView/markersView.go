package markersview

import (
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

type Props struct {
	Th          *theme.RepeatTheme
	TimeMarkers *tm.TimeMarkers
}

func NewMarkersView(props Props) *MarkersView {
	return &MarkersView{
		th:          props.Th,
		timeMarkers: props.TimeMarkers,
	}
}

type MarkersView struct {
	th          *theme.RepeatTheme
	timeMarkers *tm.TimeMarkers
}
