package editor

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	micons "github.com/spyhere/re-peat/internal/mIcons"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

func offsetBy(gtx layout.Context, amount image.Point, w func()) {
	defer op.Offset(amount).Push(gtx.Ops).Pop()
	w()
}

func backgroundComp(gtx layout.Context, col color.NRGBA) {
	ColorBox(gtx, image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)}, col)
}

func playheadComp(gtx layout.Context, th *theme.RepeatTheme, playhead int64, audio audio, scroll scroll) layout.Dimensions {
	maxX := gtx.Constraints.Max.X
	currSamples := audio.getSamplesFromPCM(playhead) - scroll.leftB
	x := int(float32(currSamples) * float32(maxX) / float32(scroll.rightB-scroll.leftB))
	if x < 0 || x > maxX {
		return layout.Dimensions{Size: image.Pt(x, 0)}
	}
	ColorBox(gtx, image.Rect(x, 0, x+th.Sizing.Editor.PlayheadW, gtx.Constraints.Max.Y), th.Palette.Editor.Playhead)
	return layout.Dimensions{Size: image.Pt(x, gtx.Constraints.Max.Y)}
}

func mCreateButtonComp(gtx layout.Context, th *theme.RepeatTheme, tag event.Tag, waveMT int, playheadDim layout.Dimensions) {
	mrkSz := th.Sizing.Editor.Markers
	c := th.Palette.Editor.AddMarker
	iconSize := th.Sizing.Editor.Markers.Lbl.IconW
	x := playheadDim.Size.X

	if x < 0 || x > gtx.Constraints.Max.X {
		return
	}

	y := waveMT - prcToPx(waveMT, th.Sizing.Editor.PlayheadButtMB)
	lblW := mrkSz.Lbl.MinW + iconSize
	labelArea := image.Rect(x, y, x+lblW, y+mrkSz.Lbl.H)
	ColorBoxR(gtx, labelArea, c, mrkSz.Lbl.CRound)
	offsetBy(gtx, image.Pt(x+(lblW/2-iconSize/2), y+(th.Sizing.Editor.Markers.Lbl.H-iconSize)/2), func() {
		gtx.Constraints.Min.X = iconSize
		micons.ContentAddCircle.Layout(gtx, th.Palette.Editor.SoundWave)
	})
	registerTag(gtx, tag, labelArea)
}

func soundWavesComp(gtx layout.Context, th *theme.RepeatTheme, yCenter float32, waves [][2]float32, s scroll, c cache) {
	yCenter = snap(yCenter)
	width := gtx.Constraints.Max.X + waveEdgePadding

	var path clip.Path
	path.Begin(gtx.Ops)

	started := false

	// --- TOP ---
	lastI0, lastI1 := -1, -1
	var prevY float32
	for px := range width {
		sample0 := s.leftB + int(float32(px)*s.samplesPerPx)
		sample1 := s.leftB + int(float32(px+1)*s.samplesPerPx)
		i0 := (sample0 / c.curLvl) - c.leftB
		i1 := (sample1 / c.curLvl) - c.leftB
		i1 = clamp(i0+1, i1, len(waves))
		if i0 == lastI0 && i1 == lastI1 {
			continue
		}
		lastI0, lastI1 = i0, i1

		_, high := reducePeaks(waves[i0:i1])
		y := snap(yCenter - high*yCenter)
		x := float32(px)

		if !started {
			path.MoveTo(f32.Pt(x, y))
			started = true
		} else {
			if y > prevY {
				path.LineTo(f32.Pt(x, prevY))
			}
			prevY = y
			path.LineTo(f32.Pt(x, y))
		}
	}

	// --- BOTTOM ---
	lastI0, lastI1 = -1, -1
	prevY = 0
	for px := width - 1; px >= 0; px-- {
		sample0 := s.leftB + int(float32(px)*s.samplesPerPx)
		sample1 := s.leftB + int(float32(px+1)*s.samplesPerPx)
		i0 := (sample0 / c.curLvl) - c.leftB
		i1 := (sample1 / c.curLvl) - c.leftB
		i1 = clamp(i0+1, i1, len(waves))
		if i0 == lastI0 && i1 == lastI1 {
			continue
		}
		lastI0, lastI1 = i0, i1

		low, _ := reducePeaks(waves[i0:i1])
		y := snap(yCenter - low*yCenter)
		x := float32(px)

		if y < prevY {
			path.LineTo(f32.Pt(x, prevY))
		}
		prevY = y
		// Add silence line
		if y == yCenter {
			y += 1
		}
		path.LineTo(f32.Pt(x, y))
	}

	path.Close()
	paint.FillShape(gtx.Ops, th.Palette.Editor.SoundWave,
		clip.Outline{Path: path.End()}.Op(),
	)
}

var timeIntervals = [5]float32{1, 5, 10, 30, 60}

func secondsRulerComp(gtx layout.Context, th *theme.RepeatTheme, audio audio, scroll scroll) {
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
		tickH := gridSizing.TickH
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
			off := op.Offset(image.Pt(x-secSizeX/2, -30)).Push(gtx.Ops)
			lbl.Layout(gtx)
			off.Pop()
		}
		ColorBox(gtx, image.Rect(x, 0, x+th.Sizing.Editor.Grid.TickW, tickH), tickC)
		curSec++
	}
}

func ColorBoxR(gtx layout.Context, size image.Rectangle, color color.NRGBA, r theme.CornerRadii) layout.Dimensions {
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

type renderable interface {
	Layout(gtx layout.Context) layout.Dimensions
}

func markersComp(gtx layout.Context, th *theme.RepeatTheme, r *widget.Editor, mode interactionMode, wavePadding int, s scroll, a audio, m *markers, mI9n mInteraction) {
	mrkSz := th.Sizing.Editor.Markers
	maxX := gtx.Constraints.Max.X
	soundWaveH := gtx.Constraints.Max.Y - wavePadding*2

	prevLblX, yOffset, colDeviation := maxX, 0, 0
	for _, marker := range m.getSortedMarkers() {
		// TODO: Implement proper culling
		var nameDim layout.Dimensions
		isEditing := m.editing == marker && mode == modeMEdit
		nameOp := makeMacro(gtx.Ops, func() {
			var renderable renderable
			if isEditing {
				renderable = material.Editor(th.Theme, r, "")
				gtx.Execute(key.FocusCmd{Tag: r})
			} else {
				name := truncName(marker.name, mrkSz.Lbl.MaxGlyphs)
				renderable = material.Body2(th.Theme, name)
			}
			inset := unit.Dp(mrkSz.Lbl.Margin)
			gtx.Constraints.Min = image.Point{}
			nameDim = layout.UniformInset(inset).Layout(gtx, renderable.Layout)
		})
		curSamples := a.getSamplesFromPCM(marker.pcm)
		x := int(float32(curSamples-s.leftB) / s.samplesPerPx)
		if x+nameDim.Size.X+mrkSz.Lbl.InvisPad >= prevLblX && prevLblX != maxX {
			yOffset += mrkSz.Lbl.H + mrkSz.Lbl.InvisPad
			colDeviation += th.Palette.Editor.MarkerDev
		} else {
			yOffset = 0
			colDeviation = 0
		}
		offsetBy(gtx, image.Pt(x, wavePadding), func() {
			markerComp(gtx, th,
				markerProps{
					isEditing:    isEditing,
					tags:         marker.tags,
					i9n:          mI9n,
					height:       soundWaveH,
					yOffset:      yOffset,
					nameOp:       nameOp,
					nameDim:      nameDim,
					colDeviation: uint8(colDeviation),
				},
			)
		})
		prevLblX = x
	}
}

type markerProps struct {
	isEditing    bool
	tags         *markerTags
	i9n          mInteraction
	height       int
	yOffset      int
	nameOp       op.CallOp
	nameDim      layout.Dimensions
	colDeviation uint8
}

func markerComp(gtx layout.Context, th *theme.RepeatTheme, mProps markerProps) layout.Dimensions {
	var col color.NRGBA
	col = th.Palette.Editor.Playhead
	col.R -= mProps.colDeviation
	col.G -= mProps.colDeviation
	col.B -= mProps.colDeviation
	mrkSz := th.Sizing.Editor.Markers
	// Pole
	poleYPad := prcToPx(mProps.height, mrkSz.Pole.Pad)
	poleH := poleYPad*2 + mProps.height
	y := -poleYPad
	ColorBox(gtx, image.Rect(0, y, mrkSz.Pole.W, y+mProps.yOffset+poleH), col)
	if mProps.i9n.pole {
		passOp := pointer.PassOp{}.Push(gtx.Ops)
		activePadding := th.Sizing.Editor.Markers.Pole.ActiveWPad
		activeArea := image.Rect(0, 0, mrkSz.Pole.W, poleH-poleYPad)
		activeArea.Min.X -= activePadding
		activeArea.Max.X += activePadding
		registerTag(gtx, &mProps.tags.pole, activeArea)
		passOp.Pop()
	}

	// Flag
	var path clip.Path
	path.Begin(gtx.Ops)
	flagHalfW := float32(mrkSz.Pole.FlagW) / 2
	// N
	poleCenter := float32(mrkSz.Pole.W) / 2
	yF := float32(-poleYPad)
	path.MoveTo(f32.Pt(poleCenter, yF))
	// NE
	path.Line(f32.Pt(flagHalfW, 0))
	// SE
	// tan(corner) = flagH / flagW
	notchVrtxY := int(math.Tan(mrkSz.Pole.FlagCorn) * float64(flagHalfW))
	path.Line(f32.Pt(0, float32(mrkSz.Pole.FlagH-notchVrtxY)))
	// S
	path.Line(f32.Pt(-flagHalfW, float32(notchVrtxY)))
	path.Close()
	pathSpec := path.End()
	paint.FillShape(gtx.Ops, col,
		clip.Outline{Path: pathSpec}.Op(),
	)
	// Mirror Flag
	t := f32.NewAffine2D(
		-1, 0, 2*poleCenter,
		0, 1, 0,
	)
	mir := op.Affine(t).Push(gtx.Ops)
	paint.FillShape(gtx.Ops, col,
		clip.Outline{Path: pathSpec}.Op(),
	)
	mir.Pop()
	if mProps.i9n.flag {
		iconSize := th.Sizing.Editor.Markers.Lbl.IconW
		offsetBy(gtx, image.Pt(-int(flagHalfW), int(yF)), func() {
			gtx.Constraints.Min.X = iconSize
			micons.Delete.Layout(gtx, th.Palette.Editor.SoundWave)
		})
		flagArea := image.Rect(-int(flagHalfW), int(yF), int(flagHalfW), int(yF)+mrkSz.Pole.FlagH)
		registerTag(gtx, &mProps.tags.flag, flagArea)
	}

	// Label
	lblOffset := prcToPx(poleH, mrkSz.Lbl.OffsetY)
	y = y + mProps.yOffset + poleH - lblOffset
	lblW := mProps.nameDim.Size.X + mrkSz.Lbl.Margin
	lblW = max(mrkSz.Lbl.MinW, lblW)
	lblH := max(mrkSz.Lbl.H, mProps.nameDim.Size.Y)
	lblArea := image.Rect(0, y, lblW, y+lblH)
	ColorBoxR(gtx, lblArea, col, mrkSz.Lbl.CRound)
	if mProps.i9n.label {
		registerTag(gtx, &mProps.tags.label, lblArea)
	}
	halfMargin := mrkSz.Lbl.Margin / 2
	offsetBy(gtx, image.Pt(halfMargin, y+halfMargin), func() {
		mProps.nameOp.Add(gtx.Ops)
	})
	return layout.Dimensions{Size: image.Pt(lblW, poleH)}
}
