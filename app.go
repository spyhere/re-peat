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
)

func newApp(appState *state.AppState) *App {
	appInstance := &App{
		AppState: appState,
		buttons:  newButtons(),
		projectView: projectview.NewProjectView(projectview.Props{
			State: appState,
		}),
		markersView: markersview.NewMarkersView(markersview.Props{
			State: appState,
		}),
	}
	ed := editorview.NewEditor(editorview.EditorProps{
		State:         appState,
		OnStartEditCb: appInstance.onStartMarkerEdit,
		OnStopEditCb:  appInstance.onStopMarkerEdit,
	})
	appInstance.editorView = ed
	return appInstance
}

type tab int

const (
	Project tab = iota
	Markers
	Editor
)

type App struct {
	*state.AppState
	projectView projectview.ProjectView
	markersView markersview.MarkersView
	editorView  editorview.Editor
	selectedTab tab
	buttons
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
	a.Dialog.Update(gtx)
	gtxEnabled := gtx
	if a.Dialog.ShouldDisableGtx(gtx) {
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

	common.OffsetBy(gtx, image.Pt(0, a.Th.Sizing.SegButtonsTopM), func(gtx layout.Context) {
		common.CenteredX(gtx, func() layout.Dimensions {
			return groupedButtons(gtx, a.Th, a.selectedTab, a.buttons)
		})
	})
	if a.buttons.isPointerHitting {
		common.SetCursor(gtx, pointer.CursorPointer)
	}
	a.Dialog.Layout(gtxEnabled)
	if cursor, ok := a.Dialog.GetCursorType(); ok {
		common.SetCursor(gtx, cursor)
	}

	a.Prompter.Layout(gtx)

	if a.AppState.IsLoading() {
		common.DrawBlockingMessage(gtx, a.Th, "Loading file...")
	}
	return layout.Dimensions{}
}
