package editor

import (
	"image"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
)

func (ed *Editor) Layout(gtx layout.Context, e app.FrameEvent) layout.Dimensions {
	ed.dispatch(gtx)
	ed.handleKey(gtx)
	ed.updateDifferedState()

	backgroundComp(gtx, ed.th.Palette.Editor.Bg)
	registerTag(gtx, &ed.tags.mLife, image.Rect(0, 0, gtx.Constraints.Max.X, ed.waveM))

	yCenter := gtx.Constraints.Max.Y / 2
	offsetBy(gtx, image.Pt(-1, ed.waveM), func() {
		soundWavesComp(gtx, ed.th, float32(yCenter-ed.waveM), ed.getRenderableWaves(), ed.scroll, ed.cache)
	})
	registerTag(gtx, &ed.tags.soundWave, image.Rect(0, ed.waveM, gtx.Constraints.Max.X, gtx.Constraints.Max.Y-ed.waveM))

	registerTag(gtx, &ed.tags.noneArea, image.Rect(0, gtx.Constraints.Max.Y-ed.waveM, gtx.Constraints.Max.X, gtx.Constraints.Max.Y))

	pDim := playheadComp(gtx, ed.th, ed.playhead.bytes, ed.audio, ed.scroll)
	if ed.p.IsPlaying() {
		if ed.playhead.bytes < ed.audio.pcmLen {
			gtx.Source.Execute(op.InvalidateCmd{At: gtx.Now.Add(ed.playhead.update)})
		}
		ed.listenToPlayerUpdates()
	}
	markersComp(gtx, ed.th, ed.renamer, ed.mode, ed.waveM, ed.scroll, ed.audio, ed.markers, ed.getMI9n())
	offsetBy(gtx, image.Pt(0, ed.waveM+prcToPx(ed.waveM, ed.th.Sizing.Editor.Grid.MargT)), func() {
		secondsRulerComp(gtx, ed.th, ed.audio, ed.scroll)
	})
	if ed.isCreateButtonVisible() {
		mCreateButtonComp(gtx, ed.th, &ed.tags.mCreateButton, ed.waveM, pDim)
	}
	setCursor(gtx, ed.cursor)
	return layout.Dimensions{}
}
