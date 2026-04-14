package projectview

import (
	"gioui.org/widget"
	"github.com/spyhere/re-peat/internal/state"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

type Props struct {
	Th    *theme.RepeatTheme
	State *state.AppState
}

func NewProjectView(props Props) ProjectView {
	return ProjectView{
		th:       props.Th,
		AppState: props.State,
	}
}

type ProjectView struct {
	*state.AppState
	th              *theme.RepeatTheme
	audioLoadCl     widget.Clickable
	markersLoadCl   widget.Clickable
	markersSaveCl   widget.Clickable
	markersSaveAsCl widget.Clickable
	disabledCl      widget.Clickable
}

func (p *ProjectView) isDisabled() bool {
	return p.AppState.IsLoading() || p.AppState.IsChoosing()
}
