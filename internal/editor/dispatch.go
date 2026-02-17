package editor

import (
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/spyhere/re-peat/internal/common"
)

func (ed *Editor) dispatch(gtx layout.Context) {
	ed.dispatchMEditorEvent(gtx)
	ed.dispatchKeyEvents(gtx)

	ed.dispatchMLifeEvent(gtx)
	ed.dispatchSoundWaveEvent(gtx)
	ed.dispatchNoneEvent(gtx)

	ed.dispatchMCreateButtonEvent(gtx)
	ed.dispatchMarkerEvent(gtx)

	ed.dispatchBackdropEvent(gtx)
}

func (ed *Editor) dispatchKeyEvents(gtx layout.Context) {
	common.HandleKeyEvents(gtx, ed.handleKeyEvents,
		key.Filter{
			Name: key.NameSpace,
		},
		key.Filter{
			Name: key.NameEscape,
		},
		key.Filter{
			Name: key.NameLeftArrow,
		},
		key.Filter{
			Name: key.NameRightArrow,
		},
	)
}

func (ed *Editor) dispatchMLifeEvent(gtx layout.Context) {
	common.HandlePointerEvents(
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
	common.HandlePointerEvents(
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
	common.HandlePointerEvents(
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
		common.HandlePointerEvents(
			gtx,
			&marker.tags.flag,
			pointer.Press|pointer.Move,
			func(e pointer.Event) {
				ed.handlePointer(pointerEvent{
					Event: e,
					Target: hitTarget{
						Kind:   hitMDeleteArea,
						Marker: marker,
					},
				})
			},
		)
		common.HandlePointerEvents(
			gtx,
			&marker.tags.pole,
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
		common.HandlePointerEvents(
			gtx,
			&marker.tags.label,
			pointer.Move|pointer.Press,
			func(e pointer.Event) {
				ed.handlePointer(pointerEvent{
					Event: e,
					Target: hitTarget{
						Kind:   hitMName,
						Marker: marker,
					},
				})
			},
		)
	}
}

func (ed *Editor) dispatchMCreateButtonEvent(gtx layout.Context) {
	common.HandlePointerEvents(
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

func (ed *Editor) dispatchMEditorEvent(gtx layout.Context) {
	for {
		we, ok := ed.mEditor.Update(gtx)
		if !ok {
			break
		}
		ed.handleMEditor(we)
	}
}

func (ed *Editor) dispatchBackdropEvent(gtx layout.Context) {
	common.HandlePointerEvents(
		gtx,
		&ed.tags.backdrop,
		pointer.Enter|pointer.Move|pointer.Press,
		func(e pointer.Event) {
			ed.handlePointer(pointerEvent{
				Event: e,
				Target: hitTarget{
					Kind: hitBackdrop,
				},
			})
		},
	)
}
