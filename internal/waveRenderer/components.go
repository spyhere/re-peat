package waverenderer

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

func offsetBy(gtx layout.Context, amount image.Point, w func()) {
	defer op.Offset(amount).Push(gtx.Ops).Pop()
	w()
}

func backgroundComp(gtx layout.Context, col color.NRGBA) {
	ColorBox(gtx, image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)}, col)
}

func playheadComp(gtx layout.Context, playhead int, audio audio, scroll scroll) {
	currSec := float32(playhead) * audio.secsPerByte
	// Since leftB is an index of monoPLC we need to divide it only by sampleRate
	// Optimisation: seconds value for left border can be saved when building waves
	leftBSec := float32(scroll.leftB) / float32(audio.sampleRate)
	xCoord := int((currSec - leftBSec) * max(scroll.minPxPerSec, scroll.deltaY))
	if xCoord < 0 || xCoord > gtx.Constraints.Max.X {
		return
	}
	ColorBox(gtx, image.Rect(xCoord, 0, xCoord+1, gtx.Constraints.Max.Y), color.NRGBA{R: 0xff, G: 0xdd, B: 0xdd, A: 0xff})
}

func soundWavesComp(gtx layout.Context, yBorder float32, waves [][2]float32) {
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(0, yBorder))
	for idx, it := range waves {
		high := yBorder - it[1]*yBorder
		low := yBorder - it[0]*yBorder
		path.LineTo(f32.Pt(float32(idx+1), high))
		path.LineTo(f32.Pt(float32(idx+1), low))
	}
	path.MoveTo(f32.Pt(0, yBorder))
	path.Close()
	paint.FillShape(gtx.Ops, color.NRGBA{G: 0x32, B: 0x55, A: 0xff},
		clip.Stroke{
			Path:  path.End(),
			Width: 1,
		}.Op(),
	)
}

func ColorBox(gtx layout.Context, size image.Rectangle, color color.NRGBA) layout.Dimensions {
	defer clip.Rect(size).Push(gtx.Ops).Pop()
	paint.ColorOp{Color: color}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{Size: size.Size()}
}
