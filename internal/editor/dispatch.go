package editor

import (
	"gioui.org/io/pointer"
	"gioui.org/layout"
)

func (ed *Editor) dispatch(gtx layout.Context) {
	ed.dispatchMLifeEvent(gtx)
	ed.dispatchSoundWaveEvent(gtx)
	ed.dispatchNoneEvent(gtx)

	ed.dispatchMCreateButtonEvent(gtx)
	ed.dispatchMarkerEvent(gtx)
}

func (ed *Editor) dispatchMLifeEvent(gtx layout.Context) {
	handlePointerEvents(
		gtx,
		&ed.tags.mLife,
		pointer.Move,
		func(e pointer.Event) {
			ed.handlePointer(pointerEvent{
				Event: e,
				Target: hitTarget{
					Kind: hitMLifeArea,
				},
			})
		},
	)
}

func (ed *Editor) dispatchSoundWaveEvent(gtx layout.Context) {
	handlePointerEvents(
		gtx,
		&ed.tags.soundWave,
		pointer.Enter|pointer.Press|pointer.Scroll|pointer.Move,
		func(e pointer.Event) {
			ed.handlePointer(pointerEvent{
				Event: e,
				Target: hitTarget{
					Kind: hitSoundWave,
				},
			})
		},
	)
}

func (ed *Editor) dispatchNoneEvent(gtx layout.Context) {
	handlePointerEvents(
		gtx,
		&ed.tags.noneArea,
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

func (ed *Editor) dispatchMarkerEvent(gtx layout.Context) {
	for _, marker := range ed.markers.arr {
		handlePointerEvents(
			gtx,
			&marker.Tag,
			pointer.Enter|pointer.Press|pointer.Move|pointer.Drag|pointer.Release,
			func(e pointer.Event) {
				ed.handlePointer(pointerEvent{
					Event: e,
					Target: hitTarget{
						Kind:   hitM,
						Marker: marker,
					},
				})
			},
		)
	}
}

func (ed *Editor) dispatchMCreateButtonEvent(gtx layout.Context) {
	handlePointerEvents(
		gtx,
		&ed.tags.mCreateButton,
		pointer.Move|pointer.Press,
		func(e pointer.Event) {
			ed.handlePointer(pointerEvent{
				Event: e,
				Target: hitTarget{
					Kind: hitMCreateArea,
				},
			})
		},
	)
}
