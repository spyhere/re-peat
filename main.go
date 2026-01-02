package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
)

const audioFilePath = "./assets/test_song.mp3"

func main() {
	decoder, rawAudio, err := decodeFile(audioFilePath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Decoded successfully!")
	normalisedSamples, err := getNormalisedSamples(rawAudio)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Audio data is normalised")
	renderableWaves := RenderableWaves{
		SampleRate: decoder.SampleRate,
		Samples:    normalisedSamples,
	}
	go func() {
		window := new(app.Window)
		window.Option(app.Title("re-peat"))
		err := run(window, renderableWaves)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

const margin = 400

func run(window *app.Window, rWaves RenderableWaves) error {
	theme := material.NewTheme()
	var ops op.Ops
	list := layout.List{}
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			_ = theme
			yMid := (e.Size.Y / 2) - margin
			rWaves.SetMaxX(e.Size.X)
			waves := rWaves.GetRenderableWaves()
			offset := op.Offset(image.Pt(0, margin)).Push(gtx.Ops)
			list.Layout(gtx, len(waves), func(gtx layout.Context, idx int) layout.Dimensions {
				offset := op.Offset(image.Pt(idx+rWaves.PxPerSec, 0)).Push(gtx.Ops)
				high := yMid - int((waves[idx][1] * float32(yMid)))
				low := yMid - int(waves[idx][0]*float32(yMid))
				box := ColorBox(gtx, image.Rect(0, high, rWaves.PxPerSec-2, low), color.NRGBA{R: 0x99, B: 0xcc, A: 0xff})
				offset.Pop()
				return box
			})
			offset.Pop()
			e.Frame(gtx.Ops)
		}
	}
}

func ColorBox(gtx layout.Context, size image.Rectangle, color color.NRGBA) layout.Dimensions {
	defer clip.Rect(size).Push(gtx.Ops).Pop()
	paint.ColorOp{Color: color}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{Size: size.Size()}
}
