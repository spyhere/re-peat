package markersview

import "github.com/spyhere/re-peat/internal/ui/theme"

type Props struct {
	Th *theme.RepeatTheme
}

func NewMarkersView(props Props) *MarkersView {
	return &MarkersView{
		th: props.Th,
	}
}

type MarkersView struct {
	th *theme.RepeatTheme
}
