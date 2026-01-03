package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
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
	rWaves := &RenderableWaves{
		SampleRate: decoder.SampleRate,
		Frames:     len(normalisedSamples) / decoder.Channels,
		Samples:    normalisedSamples,
	}
	rWaves.MakeSamplesMono(decoder.Channels)
	go func() {
		window := new(app.Window)
		window.Option(app.Title("re-peat"))
		err := run(window, rWaves)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

const MARGIN = 400
const PADDING = 90

func run(window *app.Window, rWaves *RenderableWaves) error {
	theme := material.NewTheme()
	clickable := widget.Clickable{}
	var ops op.Ops
	list := layout.List{}
	carret := image.Pt(0, 0)
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			_ = theme
			yMid := e.Size.Y / 2
			wavesYBorder := yMid - MARGIN

			rWaves.SetMaxX(e.Size.X)
			waves := rWaves.GetRenderableWaves()

			bgArea := image.Rect(0, MARGIN-PADDING, e.Size.X, e.Size.Y-MARGIN+PADDING)
			ColorBox(gtx, bgArea, color.NRGBA{R: 0x11, G: 0x77, B: 0x66, A: 0xff})

			listOffset := op.Offset(image.Pt(0, MARGIN)).Push(gtx.Ops)
			list.Layout(gtx, len(waves), func(gtx layout.Context, idx int) layout.Dimensions {
				high := wavesYBorder - int((waves[idx][1] * float32(wavesYBorder)))
				low := wavesYBorder - int(waves[idx][0]*float32(wavesYBorder))
				return ColorBox(gtx, image.Rect(0, high, 1, low), color.NRGBA{G: 0x32, B: 0x55, A: 0xff})
			})
			listOffset.Pop()

			if clickable.Clicked(gtx) {
				clickHistory := clickable.History()
				pressX := clickHistory[len(clickHistory)-1].Position.X
				carret.X = pressX
			}
			activeA := op.Offset(bgArea.Min).Push(gtx.Ops)
			clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				semantic.Button.Add(gtx.Ops)
				return layout.Dimensions{Size: image.Pt(bgArea.Dx(), bgArea.Dy())}
			})
			activeA.Pop()

			ColorBox(gtx, image.Rect(carret.X, 0, carret.X+1, e.Size.Y), color.NRGBA{R: 0xff, G: 0xdd, B: 0xdd, A: 0xff})
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
