package projectview

import "github.com/spyhere/re-peat/internal/ui/theme"

type Props struct {
	Th *theme.RepeatTheme
}

func NewProjectView(props Props) ProjectView {
	return ProjectView{
		th: props.Th,
	}
}

type ProjectView struct {
	th *theme.RepeatTheme
}
