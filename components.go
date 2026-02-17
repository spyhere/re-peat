package main

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	micons "github.com/spyhere/re-peat/internal/mIcons"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

type segmentedButtSpecs struct {
	height        unit.Dp
	outline       unit.Dp
	minPad        unit.Dp
	iconSize      unit.Dp
	elementsGap   unit.Dp
	fontFace      font.Typeface
	fontWeight    font.Weight
	fontSize      unit.Sp
	fontLineHeigt unit.Sp
}

var segButtSpecs segmentedButtSpecs = segmentedButtSpecs{
	height:        40,
	outline:       1,
	minPad:        16,
	iconSize:      18,
	elementsGap:   8,
	fontFace:      "Roboto",
	fontWeight:    500,
	fontSize:      14,
	fontLineHeigt: 20,
}

type buttonText struct {
	width int
	op    op.CallOp
}

var buttTextOps [3]buttonText = [3]buttonText{}

func groupedButtons(gtx layout.Context, th *theme.RepeatTheme, selectedT tab, buttons *buttons) layout.Dimensions {
	var maxDim layout.Dimensions
	for idx, it := range buttons.arr {
		var textDim layout.Dimensions
		textOp := common.MakeMacro(gtx.Ops, func() {
			gtx.Constraints.Min = image.Point{}
			textBody2 := material.Body2(th.Theme, it.name)
			if buttons.arr[idx].tab == selectedT {
				textBody2.Color = th.Palette.SegButtons.SelText
			} else {
				textBody2.Color = th.Palette.SegButtons.UnSelText
			}
			textBody2.Font.Typeface = segButtSpecs.fontFace
			textBody2.Font.Weight = segButtSpecs.fontWeight
			textBody2.TextSize = segButtSpecs.fontSize
			textBody2.LineHeight = segButtSpecs.fontLineHeigt
			textDim = textBody2.Layout(gtx)
		})
		buttTextOps[idx] = buttonText{width: textDim.Size.X, op: textOp}
		if textDim.Size.X > maxDim.Size.X {
			maxDim = textDim
		}
	}
	iconSize, elementsGap, containerH :=
		gtx.Dp(segButtSpecs.iconSize), gtx.Dp(segButtSpecs.elementsGap), gtx.Dp(segButtSpecs.height)
	maxDim.Size.X += iconSize + elementsGap + gtx.Dp(segButtSpecs.minPad)*2

	containerHHalf := containerH / 2
	var xOffset int
	for idx, it := range buttTextOps {
		curButt := buttons.arr[idx]
		common.OffsetBy(gtx, image.Pt(xOffset, 0), func() {
			var corner theme.CornerRadii
			if idx == 0 {
				corner = theme.CornerR(0, containerHHalf, containerHHalf, 0)
			} else if idx == len(buttons.arr)-1 {
				corner = theme.CornerR(containerHHalf, 0, 0, containerHHalf)
			}
			var col color.NRGBA
			if curButt.tab == selectedT {
				col = th.Palette.SegButtons.Selected
			}
			buttArea := image.Rect(0, 0, maxDim.Size.X, containerH)
			common.DrawBox(gtx, common.Box{
				Size:    buttArea,
				Color:   col,
				R:       corner,
				StrokeW: segButtSpecs.outline,
				StrokeC: th.Palette.SegButtons.Outline,
			})
			common.RegisterTag(gtx, &curButt.tag, buttArea)

			y := (containerH - maxDim.Size.Y) / 2
			if curButt.tab == selectedT {
				x := (maxDim.Size.X - it.width - iconSize - elementsGap) / 2
				common.OffsetBy(gtx, image.Pt(x, y), func() {
					gtx.Constraints.Min.X = iconSize
					micons.Check.Layout(gtx, th.Palette.SegButtons.SelText)
					common.OffsetBy(gtx, image.Pt(iconSize+elementsGap, 0), func() {
						it.op.Add(gtx.Ops)
					})
				})
			} else {
				x := (maxDim.Size.X - it.width) / 2
				common.OffsetBy(gtx, image.Pt(x, y), func() {
					it.op.Add(gtx.Ops)
				})
			}
		})
		xOffset += maxDim.Size.X
	}
	return layout.Dimensions{Size: image.Pt(xOffset, containerH)}
}
