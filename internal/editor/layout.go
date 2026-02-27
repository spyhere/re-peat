package editor

import (
	"image"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/spyhere/re-peat/internal/common"
)

func (ed *Editor) Layout(gtx layout.Context, e app.FrameEvent) layout.Dimensions {
	ed.dispatch(gtx)
	ed.updateDifferedState()

	common.DrawBackground(gtx, ed.th.Palette.Editor.Bg)
	common.RegisterTag(gtx, &ed.tags.mLife, image.Rect(0, 0, gtx.Constraints.Max.X, ed.waveM))

	yCenter := gtx.Constraints.Max.Y / 2
	offsetBy(gtx, image.Pt(-1, ed.waveM), func() {
		soundWavesComp(gtx, ed.th, float32(yCenter-ed.waveM), ed.getRenderableWaves(), ed.scroll, ed.cache)
	})
	common.RegisterTag(gtx, &ed.tags.soundWave, image.Rect(0, ed.waveM, gtx.Constraints.Max.X, gtx.Constraints.Max.Y-ed.waveM))

	common.RegisterTag(gtx, &ed.tags.noneArea, image.Rect(0, gtx.Constraints.Max.Y-ed.waveM, gtx.Constraints.Max.X, gtx.Constraints.Max.Y))

	pDim := playheadComp(gtx, ed.th, ed.playhead.bytes, ed.audio, ed.scroll)
	if ed.p.IsPlaying() {
		if ed.playhead.bytes < ed.audio.pcmLen {
			gtx.Source.Execute(op.InvalidateCmd{At: gtx.Now.Add(ed.playhead.update)})
		}
		ed.listenToPlayerUpdates()
	}
	markersComp(gtx, ed.th, ed.mEditor, ed.mode, ed.waveM, ed.scroll, ed.audio, ed.markers, ed.getMI9n)
	secondsGridComp(gtx, ed.th, ed.audio, ed.scroll, ed.waveM)
	if ed.markers.isEditing() {
		editingMarkerComp(gtx, ed.th, &ed.tags.backdrop, ed.markers.overlayParams)
	}
	if ed.isCreateButtonVisible() {
		mCreateButtonComp(gtx, ed.th, &ed.tags.mCreateButton, ed.waveM, pDim)
	}
	common.SetCursor(gtx, ed.cursor)
	return layout.Dimensions{}
}
