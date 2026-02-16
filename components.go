package main

import (
	"image"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

const outlineWidth unit.Dp = 1
const minPadDP unit.Dp = 12

// TODO: make helper to make it 50% according to the height of the button
var cornerL = theme.CornerR(0, 35, 35, 0)
var cornerR = theme.CornerR(35, 0, 0, 35)

func groupedButtons(gtx layout.Context, th *theme.RepeatTheme, buttons *buttons) {
	var xOffset int
	minPad := gtx.Dp(minPadDP)
	buttonsLayout := common.MakeMacro(gtx.Ops, func() {
		for idx, it := range buttons.arr {
			var textDim layout.Dimensions
			textOp := common.MakeMacro(gtx.Ops, func() {
				gtx.Constraints.Min = image.Point{}
				textBody2 := material.Body2(th.Theme, it.name)
				textDim = textBody2.Layout(gtx)
			})

			common.OffsetBy(gtx, image.Pt(xOffset, 0), func() {
				buttArea := image.Rect(0, 0, textDim.Size.X+minPad*2, textDim.Size.Y+minPad*2)
				var corner theme.CornerRadii
				if idx == 0 {
					corner = cornerL
				} else if idx == len(buttons.arr)-1 {
					corner = cornerR
				}
				common.ColorBox(gtx, common.Box{
					Size:    buttArea,
					Color:   th.Palette.Editor.Bg,
					R:       corner,
					StrokeW: outlineWidth,
					StrokeC: th.Palette.GrButtons.Outline,
				})
				common.OffsetBy(gtx, image.Pt(int(minPad), (buttArea.Dy()-textDim.Size.Y)/2), func() {
					textOp.Add(gtx.Ops)
				})
				common.RegisterTag(gtx, &it.tag, buttArea)
			})
			xOffset += textDim.Size.X + int(minPad)*2
		}
	})
	common.OffsetBy(gtx, image.Pt(-xOffset/2, 0), func() {
		buttonsLayout.Add(gtx.Ops)
	})
}
