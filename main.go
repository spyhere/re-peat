package main

import (
	"fmt"
	"log"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/widget/material"
)

const audioFilePath = "./assets/test_song.mp3"

func main() {
	decoder, rawAudio, err := decodeFile(audioFilePath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Decoded successfully!")
	if err = player(decoder, rawAudio); err != nil {
		log.Fatal(err)
	}
	// go func() {
	// 	window := new(app.Window)
	// 	window.Option(app.Title("re-peat"))
	// 	err := run(window)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	os.Exit(0)
	// }()
	// app.Main()
}

func run(window *app.Window) error {
	theme := material.NewTheme()
	var ops op.Ops
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			_ = theme
			e.Frame(gtx.Ops)
		}
	}
}
