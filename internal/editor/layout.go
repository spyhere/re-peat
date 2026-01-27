package editor

import (
	"image"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
)

func (ed *Editor) Layout(gtx layout.Context, e app.FrameEvent) layout.Dimensions {
	player := ed.p
	isPlaying := player.IsPlaying()
	ed.dispatch(gtx)
	ed.handleKey(gtx, isPlaying)

	backgroundComp(gtx, ed.th.Palette.Editor.Bg)
	registerTag(gtx, ed, image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Min.Y))

	yCenter := gtx.Constraints.Max.Y / 2
	offsetBy(gtx, image.Pt(0, ed.waveM), func() {
		soundWavesComp(gtx, ed.th, float32(yCenter-ed.waveM), ed.getRenderableWaves(), ed.scroll, ed.cache)
	})
	registerTag(gtx, ed.waveTag, image.Rect(0, ed.waveM, gtx.Constraints.Max.X, gtx.Constraints.Max.Y-ed.waveM))

	playheadComp(gtx, ed.th, ed.playhead, ed.audio, ed.scroll)
	if isPlaying {
		if ed.playhead < ed.audio.pcmLen {
			gtx.Source.Execute(op.InvalidateCmd{At: gtx.Now.Add(ed.playheadUpdate)})
		}
		ed.listenToPlayerUpdates()
	}
	markersComp(gtx, ed.th, ed.waveM, ed.scroll, ed.markers, ed.shouldMarkersInterest())
	offsetBy(gtx, image.Pt(0, ed.waveM+prcToPx(ed.waveM, ed.th.Sizing.Editor.Grid.MargT)), func() {
		secondsRulerComp(gtx, ed.th, ed.audio, ed.scroll)
	})
	newMarkerComp(gtx, ed.th, ed.waveM, ed.markers)
	setCursor(gtx, ed.cursor)
	return layout.Dimensions{}
}
