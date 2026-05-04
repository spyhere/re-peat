package editorview

import (
	"image"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/spyhere/re-peat/internal/common"
)

func (ed *Editor) Layout(gtx layout.Context) layout.Dimensions {
	ed.dispatch(gtx)
	ed.updateDifferedState()
	if ed.HasAudioLoaded() && ed.Player.IsPlaying() {
		if !ed.Player.IsEOF() {
			gtx.Source.Execute(op.InvalidateCmd{At: gtx.Now.Add(ed.playheadUpd)})
		}
		ed.listenToPlayerUpdates()
	}

	common.DrawBackground(gtx, ed.Th.Palette.Editor.Bg)
	common.RegisterTag(gtx, &ed.tags.mLife, image.Rect(0, 0, gtx.Constraints.Max.X, ed.waveM))

	if ed.HasAudioLoaded() && len(ed.MonoSamples) == 0 {
		if ed.makeCacheCl.Clicked(gtx) {
			ed.makeCacheCl = widget.Clickable{}
			ed.DecodeAllSamples()
		}
		drawCreateCacheButton(gtx, ed.Th, &ed.makeCacheCl, ed.IsDecoding(), ed.I18n.Editor.BuildWave)
		if ed.IsDecoding() {
			common.DrawBlockingMessage(gtx, ed.Th, ed.I18n.Editor.BuildingWave)
		}
		if ed.makeCacheCl.Hovered() {
			common.SetCursor(gtx, pointer.CursorPointer)
		}
		return layout.Dimensions{}
	}

	yCenter := gtx.Constraints.Max.Y / 2
	offsetBy(gtx, image.Pt(-1, ed.waveM), func() {
		soundWavesComp(gtx, ed.Th, float32(yCenter-ed.waveM), ed.getRenderableWaves(), ed.scroll, ed.cache)
	})
	common.RegisterTag(gtx, &ed.tags.soundWave, image.Rect(0, ed.waveM, gtx.Constraints.Max.X, gtx.Constraints.Max.Y-ed.waveM))

	common.RegisterTag(gtx, &ed.tags.noneArea, image.Rect(0, gtx.Constraints.Max.Y-ed.waveM, gtx.Constraints.Max.X, gtx.Constraints.Max.Y))

	pDim := playheadComp(gtx, ed.Th, ed.Playhead.Samples, ed.scroll)
	markersComp(gtx, ed.Th, ed.mEditor, ed.mode, ed.waveM, ed.scroll, ed.markers, ed.getMI9n)
	secondsGridComp(gtx, ed.Th, ed.AudioMeta, ed.scroll, ed.waveM)
	if ed.markers.isEditing() {
		editingMarkerComp(gtx, ed.Th, &ed.tags.backdrop, ed.markers.overlayParams)
	}
	if ed.isCreateButtonVisible() {
		mCreateButtonComp(gtx, ed.Th, &ed.tags.mCreateButton, ed.waveM, pDim)
	}
	common.SetCursor(gtx, ed.cursor)
	return layout.Dimensions{}
}
