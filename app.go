package main

import (
	"image"
	"log"

	"gioui.org/app"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	"github.com/spyhere/re-peat/internal/editorView"
	"github.com/spyhere/re-peat/internal/markersView"
	p "github.com/spyhere/re-peat/internal/player"
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
	player, err := p.NewPlayer(decoder, pcm)
	if err != nil {
		log.Fatal(err)
	}
	player.SetVolume(0.7)
	timeMarkers := tm.NewTimeMarkers()
	appInstance := &App{
		th:      th,
		buttons: newButtons(),
		markersView: markersview.NewMarkersView(markersview.Props{
			Th: th,
		}),
		timeMarkers: timeMarkers,
	}
	ed, err := editorview.NewEditor(editorview.EditorProps{
		Dec:           decoder,
		Player:        player,
		Th:            th,
		Pcm:           pcm,
		TimeMarkers:   timeMarkers,
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
	a.dispatch(gtx)
	switch a.selectedTab {
	case Project:
		material.H1(a.th.Theme, "Project").Layout(gtx)
	case Markers:
		a.markersView.Layout(gtx)
	case Editor:
		a.editorView.SetSize(e.Size)
		a.editorView.MakePeakMap()
		a.editorView.Layout(gtx, e)
	}
	common.OffsetBy(gtx, image.Pt(0, a.th.Sizing.SegButtonsTopM), func() {
		common.CenteredX(gtx, func() layout.Dimensions {
			return groupedButtons(gtx, a.th, a.selectedTab, a.buttons)
		})
	})
	if a.buttons.isPointerHitting {
		common.SetCursor(gtx, pointer.CursorPointer)
	}
	return layout.Dimensions{}
}
