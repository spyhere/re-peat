package common

import (
	"cmp"
	"image"
	"log"
	"math"
	"strconv"
	"strings"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
)

func SetCursor(gtx layout.Context, cursor pointer.Cursor) {
	pointer.Cursor(cursor).Add(gtx.Ops)
}

// TODO: Make it return dimensions of the widget
func OffsetBy(gtx layout.Context, amount image.Point, w func()) {
	defer op.Offset(amount).Push(gtx.Ops).Pop()
	w()
}

func CenteredX(gtx layout.Context, w func() layout.Dimensions) layout.Dimensions {
	var dimensions layout.Dimensions
	macro := MakeMacro(gtx.Ops, func() {
		dimensions = w()
	})
	xCenter := gtx.Constraints.Max.X / 2
	wCenter := dimensions.Size.X / 2
	OffsetBy(gtx, image.Pt(xCenter-wCenter, 0), func() {
		macro.Add(gtx.Ops)
	})
	return dimensions
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

func HandleKeyEvents(gtx layout.Context, cb func(e key.Event), filters ...event.Filter) {
	for {
		evt, ok := gtx.Event(filters...)
		if !ok {
			break
		}
		e, ok := evt.(key.Event)
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

// Lock "this" between "from" and "to"
func Clamp[T cmp.Ordered](from T, this T, to T) T {
	return min(max(from, this), to)
}

func PrcToPx(origin int, prc string) int {
	prc = strings.Split(prc, "%")[0]
	prcInt, err := strconv.Atoi(prc)
	if err != nil {
		log.Fatal(err)
	}
	return prcInt * origin / 100
}

func Snap(v float32) float32 {
	return float32(math.Round(float64(v)))
}

func strlen(input string) int {
	return strings.Count(input, "") - 1
}

// Since non-lating letters are taking more then 1 byte `strlen` and manual idx in range is required
func StrTrunc(name string, limit int) string {
	if limit == 0 || strlen(name) < limit {
		return name
	}
	var newName strings.Builder
	idx := 0
	for _, r := range name {
		if idx >= limit-3 {
			break
		}
		newName.WriteRune(r)
		idx++
	}
	newName.WriteString("...")
	return newName.String()
}
