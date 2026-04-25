package main

import (
	"fmt"
	"image"
	"time"

	"gioui.org/app"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/spyhere/re-peat/internal/common"
	editorview "github.com/spyhere/re-peat/internal/editorView"
	"github.com/spyhere/re-peat/internal/logging"
	markersview "github.com/spyhere/re-peat/internal/markersView"
	projectview "github.com/spyhere/re-peat/internal/projectView"
	"github.com/spyhere/re-peat/internal/state"
)

const logDumpCooldown = time.Minute * 5

func newApp(appState *state.AppState) *App {
	fm := &common.FocusManager{}
	appInstance := &App{
		AppState: appState,
		buttons:  newButtons(&appState.I18n),
		projectView: projectview.NewProjectView(projectview.Props{
			State: appState,
		}),
		markersView: markersview.NewMarkersView(markersview.Props{
			State: appState,
		}),
		i18nSwitcher: common.NewI18nSwitcher(appState.I18n.Cur, fm),
		fm:           fm,
	}
	ed := editorview.NewEditor(editorview.EditorProps{
		State:         appState,
		OnStartEditCb: appInstance.onStartMarkerEdit,
		OnStopEditCb:  appInstance.onStopMarkerEdit,
	})
	appInstance.editorView = ed
	commonI18n := appInstance.I18n.Common
	go func() {
		appState.NotifyCrashReportsOnStartup()
		for range appState.Lg.DumpDoneCh {
			body := fmt.Sprintf(commonI18n.LogsDumpedBody, logging.LogReportFileName)
			appState.Prompter.Tell(commonI18n.LogsDumpedTitle, body, commonI18n.InfoDialogOk)
			// Intentionally block DumpDoneCh to stop spamming with the same error logs (dump + notification blocked)
			time.Sleep(logDumpCooldown)
		}
	}()
	return appInstance
}

type tab int

func (t tab) String() string {
	switch t {
	case Project:
		return "Project"
	case Markers:
		return "Markers"
	case Editor:
		return "Editor"
	default:
		panic("unreachable")
	}
}

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
	i18nSwitcher common.I18nSwitcher
	fm           *common.FocusManager
}

func (a *App) onStartMarkerEdit() {
	a.buttons.disable()
}

func (a *App) onStopMarkerEdit() {
	a.buttons.enable()
}

func (a *App) Layout(gtx layout.Context, e app.FrameEvent) layout.Dimensions {
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

	var groupedBtnsDims layout.Dimensions
	common.OffsetBy(gtx, image.Pt(0, a.Th.Sizing.SegButtonsTopM), func(gtx layout.Context) {
		common.CenteredX(gtx, func() layout.Dimensions {
			groupedBtnsDims = groupedButtons(gtx, a.Th, a.selectedTab, a.buttons)
			return groupedBtnsDims
		})
	})

	common.DrawVersion(gtx, a.Th, version)

	common.OffsetBy(gtx, image.Pt(gtx.Constraints.Max.X-400, a.Th.Sizing.SegButtonsTopM), func(gtx layout.Context) {
		gtx.Constraints.Min.Y = groupedBtnsDims.Size.Y
		common.I18nMenu(a.Th, &a.i18nSwitcher).Layout(gtx)
	})
	if a.buttons.isPointerHitting {
		common.SetCursor(gtx, pointer.CursorPointer)
	}
	if cursor, ok := a.i18nSwitcher.GetCursorType(); ok {
		common.SetCursor(gtx, cursor)
	}
	a.fm.PlaceScrim(gtx)

	a.Dialog.Layout(gtxEnabled)
	if cursor, ok := a.Dialog.GetCursorType(); ok {
		common.SetCursor(gtx, cursor)
	}

	a.Prompter.Layout(gtx)
	return layout.Dimensions{}
}
