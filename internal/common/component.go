package common

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

type Box struct {
	Size    image.Rectangle
	Color   color.NRGBA
	R       theme.CornerRadii
	StrokeC color.NRGBA
	StrokeW unit.Dp
}

func ColorBox(gtx layout.Context, b Box) layout.Dimensions {
	r := b.R
	rrect := clip.RRect{Rect: b.Size, SE: r.SE, SW: r.SW, NE: r.NE, NW: r.NW}
	rrectStack := rrect.Push(gtx.Ops)
	paint.ColorOp{Color: b.Color}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	rrectStack.Pop()
	if b.StrokeW != 0 {
		stroke := clip.Stroke{
			Path:  rrect.Path(gtx.Ops),
			Width: float32(gtx.Dp(b.StrokeW)),
		}.Op().Push(gtx.Ops)
		paint.ColorOp{Color: b.StrokeC}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		stroke.Pop()
	}
	return layout.Dimensions{Size: b.Size.Size()}
}
