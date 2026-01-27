package editor

import (
	"gioui.org/io/pointer"
	"gioui.org/layout"
)

func (ed *Editor) dispatch(gtx layout.Context) {
	ed.dispatchEditorEvents(gtx)
	ed.dispatchWaveEvents(gtx)
	ed.dispatchMarkerEvent(gtx)
}

func (ed *Editor) dispatchEditorEvents(gtx layout.Context) {
	handlePointerEvents(
		gtx,
		ed,
		pointer.Enter|pointer.Press|pointer.Move,
		func(e pointer.Event) {
			ed.handlePointer(pointerEvent{
				Event: e,
				Target: hitTarget{
					Kind: hitNone,
				},
			})
		},
	)
}

func (ed *Editor) dispatchWaveEvents(gtx layout.Context) {
	handlePointerEvents(
		gtx,
		ed.waveTag,
		pointer.Enter|pointer.Press|pointer.Scroll|pointer.Move,
		func(e pointer.Event) {
			ed.handlePointer(pointerEvent{
				Event: e,
				Target: hitTarget{
					Kind: hitWave,
				},
			})
		},
	)
}

func (ed *Editor) dispatchMarkerEvent(gtx layout.Context) {
	for _, marker := range ed.markers.arr {
		handlePointerEvents(
			gtx,
			marker,
			pointer.Enter|pointer.Press|pointer.Move|pointer.Drag|pointer.Release,
			func(e pointer.Event) {
				ed.handlePointer(pointerEvent{
					Event: e,
					Target: hitTarget{
						Kind:   hitMarker,
						Marker: marker,
					},
				})
			},
		)
	}
}
