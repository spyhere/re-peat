package common

import (
	"image"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
)

func SetCursor(gtx layout.Context, cursor pointer.Cursor) {
	pointer.Cursor(cursor).Add(gtx.Ops)
}

func OffsetBy(gtx layout.Context, amount image.Point, w func()) {
	defer op.Offset(amount).Push(gtx.Ops).Pop()
	w()
}

func RegisterTag(gtx layout.Context, tag event.Tag, area image.Rectangle) {
	defer clip.Rect(area).Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, tag)
}

func HandlePointerEvents(gtx layout.Context, tag event.Tag, pKind pointer.Kind, cb func(e pointer.Event)) {
	for {
		evt, ok := gtx.Event(pointer.Filter{
			Target:  tag,
			Kinds:   pKind,
			ScrollX: pointer.ScrollRange{Min: -1e9, Max: 1e9},
			ScrollY: pointer.ScrollRange{Min: -1e9, Max: 1e9},
		})
		if !ok {
			break
		}
		e, ok := evt.(pointer.Event)
		if !ok {
			continue
		}
		cb(e)
	}
}

func MakeMacro(ops *op.Ops, cb func()) op.CallOp {
	macro := op.Record(ops)
	cb()
	return macro.Stop()
}
