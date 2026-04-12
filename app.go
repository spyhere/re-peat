package main

import (
	"image"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/spyhere/re-peat/internal/audio"
	"github.com/spyhere/re-peat/internal/common"
	editorview "github.com/spyhere/re-peat/internal/editorView"
	markersview "github.com/spyhere/re-peat/internal/markersView"
	p "github.com/spyhere/re-peat/internal/player"
	projectview "github.com/spyhere/re-peat/internal/projectView"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

func newApp() *App {
	th, err := theme.New()
	if err != nil {
		log.Fatal(err)
	}
	decoder, pcm, err := decodeFile(audioFilePath)
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.Open(audioFilePath)
	if err != nil {
		log.Fatal(err)
	}
	player := p.NewPlayer()
	err = player.SetAudio(file)
	if err != nil {
		log.Fatal(err)
	}
	player.SetVolume(0.4)
	// TODO: Create app state and put it there
	timeMarkers := tm.NewTimeMarkers()
	a := audio.NewAudio(decoder, pcm)
	d := common.Dialog{}
	d.CancelProps.Text = "Отмена"
	appInstance := &App{
		th:      th,
		dialog:  &d,
		buttons: newButtons(),
		projectView: projectview.NewProjectView(projectview.Props{
			Th: th,
		}),
		markersView: markersview.NewMarkersView(markersview.Props{
			Audio:       a,
			Th:          th,
			TimeMarkers: &timeMarkers,
			Player:      player,
			Dialog:      &d,
		}),
		timeMarkers: timeMarkers,
	}
	ed, err := editorview.NewEditor(editorview.EditorProps{
		Audio:         a,
		Dec:           decoder,
		Player:        player,
		Th:            th,
		Pcm:           pcm,
		TimeMarkers:   &timeMarkers,
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

type App struct {
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
	return layout.Dimensions{}
}
