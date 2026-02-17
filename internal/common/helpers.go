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

func CenteredX(gtx layout.Context, w func() layout.Dimensions) {
	var dimensions layout.Dimensions
	macro := MakeMacro(gtx.Ops, func() {
		dimensions = w()
	})
	xCenter := gtx.Constraints.Max.X / 2
	wCenter := dimensions.Size.X / 2
	OffsetBy(gtx, image.Pt(xCenter-wCenter, 0), func() {
		macro.Add(gtx.Ops)
	})
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
