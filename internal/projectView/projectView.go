package projectview

import (
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
}
