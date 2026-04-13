package main

import (
	"image"
	"log"

	"gioui.org/app"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/spyhere/re-peat/internal/common"
	editorview "github.com/spyhere/re-peat/internal/editorView"
	markersview "github.com/spyhere/re-peat/internal/markersView"
	projectview "github.com/spyhere/re-peat/internal/projectView"
	"github.com/spyhere/re-peat/internal/state"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

func newApp(appState *state.AppState) *App {
	th, err := theme.New()
	if err != nil {
		log.Fatal(err)
	}
	d := common.Dialog{}
	d.CancelProps.Text = "Отмена"

	appInstance := &App{
		AppState: appState,
		th:       th,
		dialog:   &d,
		buttons:  newButtons(),
		projectView: projectview.NewProjectView(projectview.Props{
			Th:    th,
			State: appState,
		}),
		markersView: markersview.NewMarkersView(markersview.Props{
			Th:     th,
			State:  appState,
			Dialog: &d,
		}),
	}
	ed, err := editorview.NewEditor(editorview.EditorProps{
		Th:            th,
		State:         appState,
		OnStartEditCb: appInstance.onStartMarkerEdit,
		OnStopEditCb:  appInstance.onStopMarkerEdit,
	})
	if err != nil {
		log.Fatal(err)
	}
	appInstance.editorView = ed
	return appInstance
}

type tab int

const (
	Project tab = iota
	Markers
	Editor
)

// TODO: Get rid of redundant pointers
type App struct {
	*state.AppState
	dialog      *common.Dialog
	projectView projectview.ProjectView
	markersView *markersview.MarkersView
	editorView  *editorview.Editor
	selectedTab tab
	th          *theme.RepeatTheme
	timeMarkers tm.TimeMarkers
	*buttons
}

func (a *App) onStartMarkerEdit() {
	a.buttons.disable()
}

func (a *App) onStopMarkerEdit() {
	a.buttons.enable()
}

func (a *App) Layout(gtx layout.Context, e app.FrameEvent) layout.Dimensions {
	if err := a.AppState.GetError(); err != nil {
		log.Println(err)
	}
	a.dialog.Update(gtx)
	gtxEnabled := gtx
	if a.dialog.ShouldDisableGtx(gtx) {
		gtx = gtx.Disabled()
	}

	switch a.selectedTab {
	case Project:
		a.projectView.Layout(gtx)
	case Markers:
		a.markersView.Layout(gtx)
	case Editor:
		a.editorView.SetSize(e.Size)
		a.editorView.MakePeakMap()
		a.editorView.Layout(gtx)
	}
	a.dispatch(gtx)

	common.OffsetBy(gtx, image.Pt(0, a.th.Sizing.SegButtonsTopM), func(gtx layout.Context) {
		common.CenteredX(gtx, func() layout.Dimensions {
			return groupedButtons(gtx, a.th, a.selectedTab, a.buttons)
		})
	})
	if a.buttons.isPointerHitting {
		common.SetCursor(gtx, pointer.CursorPointer)
	}
	a.dialog.Layout(gtxEnabled)
	if cursor, ok := a.dialog.GetCursorType(); ok {
		common.SetCursor(gtx, cursor)
	}

	if a.AppState.IsLoading() {
		common.DrawBlockingMessage(gtx, a.th, "Loading file...")
	}
	return layout.Dimensions{}
}
