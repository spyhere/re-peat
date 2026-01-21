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
	ed.handlePointerEvents(gtx)
	ed.handleKey(gtx, isPlaying)

	setCrosshairCursor(gtx)
	backgroundComp(gtx, ed.th.Palette.Editor.Bg)

	yCenter := gtx.Constraints.Max.Y / 2
	offsetBy(gtx, image.Pt(0, ed.margin), func() {
		soundWavesComp(gtx, ed.th, float32(yCenter-ed.margin), ed.getRenderableWaves(), ed.scroll, ed.cache)
	})
	secondsRulerComp(gtx, ed.th, ed.margin-50, ed.audio, ed.scroll)

	playheadComp(gtx, ed.th, ed.playhead, ed.audio, ed.scroll)
	if isPlaying {
		if ed.playhead < ed.audio.pcmLen {
			gtx.Source.Execute(op.InvalidateCmd{At: gtx.Now.Add(ed.playheadUpdate)})
		}
		ed.listenToPlayerUpdates()
	}
	return layout.Dimensions{}
}
