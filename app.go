package main

import (
	"image"
	"log"

	"gioui.org/app"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	"github.com/spyhere/re-peat/internal/editor"
	p "github.com/spyhere/re-peat/internal/player"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

func newApp() *App {
	decoder, pcm, err := decodeFile(audioFilePath)
	if err != nil {
		log.Fatal(err)
	}
	player, err := p.NewPlayer(decoder, pcm)
	if err != nil {
		log.Fatal(err)
	}
	player.SetVolume(0.7)
	th := theme.New()
	ed, err := editor.NewEditor(th, decoder, pcm, player)
	if err != nil {
		log.Fatal(err)
	}
	return &App{
		editor:  ed,
		th:      th,
		buttons: newButtons(),
	}
}

type tab int

const (
	Project tab = iota
	Markers
	Editor
)

type App struct {
	editor      *editor.Editor
	selectedTab tab
	th          *theme.RepeatTheme
	*buttons
}

func (a *App) Layout(gtx layout.Context, e app.FrameEvent) layout.Dimensions {
	a.dispatch(gtx)
	switch a.selectedTab {
	case Project:
		material.H1(a.th.Theme, "Project").Layout(gtx)
	case Markers:
		material.H1(a.th.Theme, "Markers").Layout(gtx)
	case Editor:
		a.editor.SetSize(e.Size)
		a.editor.MakePeakMap()
		a.editor.Layout(gtx, e)
	}
	xCenter := gtx.Constraints.Max.X / 2
	common.OffsetBy(gtx, image.Pt(xCenter, 20), func() {
		groupedButtons(gtx, a.th, a.buttons)
	})
	if a.buttons.isPointerHitting {
		common.SetCursor(gtx, pointer.CursorPointer)
	}
	return layout.Dimensions{}
}
