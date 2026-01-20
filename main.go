package main

import (
	"fmt"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/op"
	"github.com/spyhere/re-peat/internal/editor"
	p "github.com/spyhere/re-peat/internal/player"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

const audioFilePath = "./assets/test_song.mp3"

func main() {
	decoder, pcm, err := decodeFile(audioFilePath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Decoded successfully")
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
	go func() {
		window := new(app.Window)
		window.Option(app.Title("re-peat"))
		err := run(window, ed)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(window *app.Window, ed *editor.Editor) error {
	var ops op.Ops
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			ed.SetSize(e.Size)
			ed.MakePeakMap()
			ed.Layout(gtx, e)

			e.Frame(gtx.Ops)
		}
	}
}
