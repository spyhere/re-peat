package main

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/unit"
)

const audioFilePath = "./assets/test_song.mp3"

func main() {
	repeatApp := newApp()
	go func() {
		window := new(app.Window)
		window.Option(app.Title("re-peat"))
		window.Option(app.Size(unit.Dp(900), unit.Dp(700)))
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
