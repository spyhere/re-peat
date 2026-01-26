package editor

import (
	"image"

	"gioui.org/app"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
)

func (ed *Editor) Layout(gtx layout.Context, e app.FrameEvent) layout.Dimensions {
	player := ed.p
	isPlaying := player.IsPlaying()
	ed.handlePointerEvents(gtx)
	ed.handleKey(gtx, isPlaying)

	backgroundComp(gtx, ed.th.Palette.Editor.Bg)

	yCenter := gtx.Constraints.Max.Y / 2
	offsetBy(gtx, image.Pt(0, ed.margin), func() {
		soundWavesComp(gtx, ed.th, float32(yCenter-ed.margin), ed.getRenderableWaves(), ed.scroll, ed.cache)
	})
	wavesArea := clip.Rect(image.Rect(0, ed.margin, gtx.Constraints.Max.X, gtx.Constraints.Max.Y-ed.margin)).Push(gtx.Ops)
	setCursor(gtx, pointer.CursorCrosshair)
	wavesArea.Pop()

	playheadComp(gtx, ed.th, ed.playhead, ed.audio, ed.scroll)
	if isPlaying {
		if ed.playhead < ed.audio.pcmLen {
			gtx.Source.Execute(op.InvalidateCmd{At: gtx.Now.Add(ed.playheadUpdate)})
		}
		ed.listenToPlayerUpdates()
	}
	markersComp(gtx, ed.th, ed.margin, ed.scroll, ed.markers)
	secondsRulerComp(gtx, ed.th, ed.margin-50, ed.audio, ed.scroll)
	newMarkerComp(gtx, ed.th, ed.margin, &ed.markers)
	return layout.Dimensions{}
}
