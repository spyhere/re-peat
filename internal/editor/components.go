package editor

import (
	"fmt"
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

func offsetBy(gtx layout.Context, amount image.Point, w func()) {
	defer op.Offset(amount).Push(gtx.Ops).Pop()
	w()
}

func backgroundComp(gtx layout.Context, col color.NRGBA) {
	ColorBox(gtx, image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)}, col)
}

func playheadComp(gtx layout.Context, th *theme.RepeatTheme, playhead int64, audio audio, scroll scroll) {
	maxX := gtx.Constraints.Max.X
	currSamples := audio.getSamplesFromPCM(playhead) - scroll.leftB
	x := int(float32(currSamples) * float32(maxX) / float32(scroll.rightB-scroll.leftB))
	if x < 0 || x > maxX {
		return
	}
	ColorBox(gtx, image.Rect(x, 0, x+2, gtx.Constraints.Max.Y), th.Editor.Playhead)
}

func soundWavesComp(gtx layout.Context, th *theme.RepeatTheme, yBorder float32, waves [][2]float32) {
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
	paint.FillShape(gtx.Ops, th.Editor.SoundWave,
		clip.Stroke{
			Path:  path.End(),
			Width: 1,
		}.Op(),
	)
}

const (
	MIN_TIME_INTERVAL_PX = 100
	TICK_LENGTH_10_SEC   = 30
	TICK_LENGTH_5_SEC    = 20
	TICK_LENGTH          = 10
)

var TICK_COLOR_10_SEC = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
var TICK_COLOR_5_SEC = color.NRGBA{G: 0xff, B: 0xbb, A: 0xff}
var TICK_COLOR = color.NRGBA{A: 0xff}

var timeIntervals = [5]float32{1, 5, 10, 30, 60}

func secondsRulerComp(gtx layout.Context, th *theme.RepeatTheme, margin int, audio audio, scroll scroll) {
	margin -= 50
	pxPerSec := scroll.getPxPerSec()
	leftBSec := audio.getSecondsFromSamples(scroll.leftB)
	var intervalSec int
	for _, it := range timeIntervals {
		if it*pxPerSec >= MIN_TIME_INTERVAL_PX {
			intervalSec = int(it)
			break
		}
	}

	nextSec, nextSecIdx := audio.getNextSecond(leftBSec)
	curSecIdx := nextSecIdx
	curSec := int(nextSec)
	for ; curSecIdx < scroll.rightB; curSecIdx += audio.sampleRate {
		// TODO: Use theme for this
		tickLength := TICK_LENGTH
		tickColor := TICK_COLOR
		if curSec%10 == 0 {
			tickLength = TICK_LENGTH_10_SEC
			tickColor = TICK_COLOR_10_SEC
		} else if curSec%5 == 0 {
			tickLength = TICK_LENGTH_5_SEC
			tickColor = TICK_COLOR_5_SEC
		}
		x := int(float64(curSecIdx-scroll.leftB) * float64(gtx.Constraints.Max.X) / float64(scroll.rightB-scroll.leftB))
		if curSec%intervalSec == 0 {
			// TODO: center seconds properly
			off := op.Offset(image.Pt(x-20, margin-30)).Push(gtx.Ops)
			material.Body2(th.Theme, fmt.Sprintf("%d", curSec)).Layout(gtx)
			off.Pop()
		}
		ColorBox(gtx, image.Rect(x, margin, x+2, margin+tickLength), tickColor)
		curSec++
	}
}

func ColorBox(gtx layout.Context, size image.Rectangle, color color.NRGBA) layout.Dimensions {
	defer clip.Rect(size).Push(gtx.Ops).Pop()
	paint.ColorOp{Color: color}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{Size: size.Size()}
}
