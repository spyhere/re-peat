package main

import (
	"fmt"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/widget/material"
	p "github.com/spyhere/re-peat/internal/player"
	wRenderer "github.com/spyhere/re-peat/internal/waveRenderer"
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
	wavesR, err := wRenderer.NewWavesRenderer(decoder, pcm, player)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		window := new(app.Window)
		window.Option(app.Title("re-peat"))
		err := run(window, wavesR)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(window *app.Window, wavesR *wRenderer.WavesRenderer) error {
	theme := material.NewTheme()
	var ops op.Ops
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			wavesR.SetSize(e.Size)
			wavesR.Layout(gtx, theme, e)

			e.Frame(gtx.Ops)
		}
	}
}
