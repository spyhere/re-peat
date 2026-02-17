package main

import (
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
	common.CenteredX(gtx, func() layout.Dimensions {
		return groupedButtons(gtx, a.th, a.selectedTab, a.buttons)
	})
	if a.buttons.isPointerHitting {
		common.SetCursor(gtx, pointer.CursorPointer)
	}
	return layout.Dimensions{}
}
