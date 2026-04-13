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
	AudioLoadCl     widget.Clickable
	MarkersLoadCl   widget.Clickable
	MarkersSaveCl   widget.Clickable
	MarkersSaveAsCl widget.Clickable
}

func (p *ProjectView) isDisabled() bool {
	return p.AppState.IsLoading() || p.AppState.IsChoosing()
}
