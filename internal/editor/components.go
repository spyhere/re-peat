package editor

import (
	"fmt"
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/io/pointer"
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
	ColorBox(gtx, image.Rect(x, 0, x+2, gtx.Constraints.Max.Y), th.Palette.Editor.Playhead)
}

func soundWavesComp(gtx layout.Context, th *theme.RepeatTheme, yCenter float32, waves [][2]float32, s scroll, c cache) {
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(0, yCenter))
	lastI0 := -1
	lastI1 := -1
	for px := range gtx.Constraints.Max.X + waveEdgePadding {
		sample0 := s.leftB + int(float32(px)*s.samplesPerPx)
		sample1 := s.leftB + int(float32(px+1)*s.samplesPerPx)
		i0 := (sample0 / c.curLvl) - c.leftB
		i1 := (sample1 / c.curLvl) - c.leftB
		i1 = min(i1+1, len(waves))
		if i0 == lastI0 && i1 == lastI1 {
			continue
		}
		lastI0 = i0
		lastI1 = i1
		low, high := reducePeaks(waves[i0:i1])
		high = yCenter - high*yCenter
		low = yCenter - low*yCenter
		path.LineTo(f32.Pt(float32(px), high))
		path.LineTo(f32.Pt(float32(px), low))
	}
	paint.FillShape(gtx.Ops, th.Palette.Editor.SoundWave,
		clip.Stroke{
			Path:  path.End(),
			Width: 1,
		}.Op(),
	)
}

var timeIntervals = [5]float32{1, 5, 10, 30, 60}

func secondsRulerComp(gtx layout.Context, th *theme.RepeatTheme, y int, audio audio, scroll scroll) {
	pxPerSec := float32(audio.sampleRate) / scroll.samplesPerPx
	leftBSec := audio.getSecondsFromSamples(scroll.leftB)
	var intervalSec int
	for _, it := range timeIntervals {
		if it*pxPerSec >= float32(th.Sizing.Editor.Grid.MinTimeInterval) {
			intervalSec = int(it)
			break
		}
	}

	nextSec, nextSecIdx := audio.getNextSecond(leftBSec)
	curSecIdx := nextSecIdx
	curSec := int(nextSec)
	gridPalette := th.Palette.Editor.Grid
	gridSizing := th.Sizing.Editor.Grid
	for ; curSecIdx < scroll.rightB; curSecIdx += audio.sampleRate {
		tickH := gridSizing.Tick
		tickC := gridPalette.Tick
		if curSec%10 == 0 {
			tickH = gridSizing.Tick10s
			tickC = gridPalette.Tick10s
		} else if curSec%5 == 0 {
			tickH = gridSizing.Tick5s
			tickC = gridPalette.Tick5s
		}
		x := int(float64(curSecIdx-scroll.leftB) * float64(gtx.Constraints.Max.X) / float64(scroll.rightB-scroll.leftB))
		if curSec%intervalSec == 0 {
			// TODO: You have to have proper text and/or widget dimension tool
			secLabel := fmt.Sprintf("%d", curSec)
			lbl := material.Body2(th.Theme, secLabel)
			secSizeX := int(lbl.TextSize) * len(secLabel)
			off := op.Offset(image.Pt(x-secSizeX/2, y-30)).Push(gtx.Ops)
			lbl.Layout(gtx)
			off.Pop()
		}
		ColorBox(gtx, image.Rect(x, y, x+2, y+tickH), tickC)
		curSec++
	}
}

type cornerRadii struct {
	SE, SW, NW, NE int
}

func cornerR(se, sw, nw, ne int) cornerRadii {
	return cornerRadii{SE: se, SW: sw, NW: nw, NE: ne}
}

func ColorBoxR(gtx layout.Context, size image.Rectangle, color color.NRGBA, r cornerRadii) layout.Dimensions {
	defer clip.RRect{Rect: size, SE: r.SE, SW: r.SW, NE: r.NE, NW: r.NW}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: color}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{Size: size.Size()}
}

func ColorBox(gtx layout.Context, size image.Rectangle, color color.NRGBA) layout.Dimensions {
	defer clip.Rect(size).Push(gtx.Ops).Pop()
	paint.ColorOp{Color: color}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{Size: size.Size()}
}

func setCursor(gtx layout.Context, cursor pointer.Cursor) {
	pointer.Cursor(cursor).Add(gtx.Ops)
}
}
