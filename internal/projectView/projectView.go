package projectview

import (
	"gioui.org/widget"
	"github.com/spyhere/re-peat/internal/logging"
	"github.com/spyhere/re-peat/internal/state"
)

type Props struct {
	State *state.AppState
	Lg    logging.Logger
}

func NewProjectView(props Props) ProjectView {
	return ProjectView{
		AppState: props.State,
		lg:       props.Lg,
	}
}

type ProjectView struct {
	*state.AppState
	lg              logging.Logger
	audioLoadCl     widget.Clickable
	markersLoadCl   widget.Clickable
	markersSaveCl   widget.Clickable
	markersSaveAsCl widget.Clickable
	disabledCl      widget.Clickable
}

func (p *ProjectView) isDisabled() bool {
	return p.AppState.IsLoading() || p.AppState.IsChoosing()
}
