package main

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/spyhere/re-peat/internal/state"
)

func main() {
	window := new(app.Window)
	appState := state.NewAppState(window)
	repeatApp := newApp(&appState)
	go func() {
		window.Option(app.Title("re-peat"))
		window.Option(app.Size(unit.Dp(1000), unit.Dp(700)))
		err := run(window, repeatApp)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(window *app.Window, repeatApp *App) error {
	var ops op.Ops
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			repeatApp.Layout(gtx, e)
			e.Frame(gtx.Ops)
		}
	}
}
