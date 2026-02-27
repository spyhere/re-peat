package main

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	micons "github.com/spyhere/re-peat/internal/mIcons"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

type segButtMaterialSpecs struct {
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

var segButtSpecs segButtMaterialSpecs = segButtMaterialSpecs{
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
	segButtonP := th.Palette.SegButtons.Enabled
	if buttons.isDisabled {
		segButtonP = th.Palette.SegButtons.Disabled
	}
	for idx, it := range buttons.arr {
		var textDim layout.Dimensions
		textOp := common.MakeMacro(gtx.Ops, func() {
			gtx.Constraints.Min = image.Point{}
			textBody2 := material.Body2(th.Theme, it.name)
			if buttons.arr[idx].tab == selectedT {
				textBody2.Color = segButtonP.SelText
			} else {
				textBody2.Color = segButtonP.UnSelText
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

	var xOffset int
	for idx, it := range buttTextOps {
		curButt := buttons.arr[idx]
		common.OffsetBy(gtx, image.Pt(xOffset, 0), func() {
			segmentedButtonComp(gtx, segmentedBProps{
				b:           curButt,
				height:      containerH,
				isFirst:     idx == 0,
				isLast:      idx == len(buttons.arr)-1,
				isSelected:  curButt.tab == selectedT,
				isDisabled:  buttons.isDisabled,
				iconSize:    iconSize,
				elementsGap: elementsGap,
				text:        it,
				textDim:     maxDim,
				th:          th,
			})
		})
		xOffset += maxDim.Size.X
	}
	return layout.Dimensions{Size: image.Pt(xOffset, containerH)}
}

type segmentedBProps struct {
	b           *button
	height      int
	isFirst     bool
	isLast      bool
	isSelected  bool
	isDisabled  bool
	iconSize    int
	elementsGap int
	text        buttonText
	textDim     layout.Dimensions
	th          *theme.RepeatTheme
}

func segmentedButtonComp(gtx layout.Context, props segmentedBProps) {
	bPalette := props.th.Palette.SegButtons.Enabled
	if props.isDisabled {
		bPalette = props.th.Palette.SegButtons.Disabled
	}
	bHoveredPalette := props.th.Palette.SegButtons.Hovered

	var corner theme.CornerRadii
	containerHHalf := props.height / 2
	if props.isFirst {
		corner = theme.CornerR(0, containerHHalf, containerHHalf, 0)
	} else if props.isLast {
		corner = theme.CornerR(containerHHalf, 0, 0, containerHHalf)
	}
	var col color.NRGBA
	if props.isSelected {
		col = bPalette.Selected
	}
	buttArea := image.Rect(0, 0, props.textDim.Size.X, props.height)
	cl := props.b.clickable
	if props.isDisabled {
		cl = nil
	}
	common.DrawBox(gtx, common.Box{
		Size:      buttArea,
		Color:     col,
		R:         corner,
		StrokeW:   segButtSpecs.outline,
		StrokeC:   bPalette.Outline,
		Clickable: cl,
	})

	if props.b.isHovered {
		col = bHoveredPalette.UnSelected
		if props.isSelected {
			col = bHoveredPalette.Selected
		}
		common.DrawBox(gtx, common.Box{
			Size:  buttArea,
			Color: col,
			R:     corner,
		})
	}

	if !props.isDisabled {
		passOp := pointer.PassOp{}.Push(gtx.Ops)
		common.RegisterTag(gtx, &props.b.tag, buttArea)
		passOp.Pop()
	}
	y := (props.height - props.textDim.Size.Y) / 2
	if props.isSelected {
		x := (props.textDim.Size.X - props.text.width - props.iconSize - props.elementsGap) / 2
		common.OffsetBy(gtx, image.Pt(x, y), func() {
			gtx.Constraints.Min.X = props.iconSize
			micons.Check.Layout(gtx, bPalette.SelText)
			common.OffsetBy(gtx, image.Pt(props.iconSize+props.elementsGap, 0), func() {
				props.text.op.Add(gtx.Ops)
			})
		})
	} else {
		x := (props.textDim.Size.X - props.text.width) / 2
		common.OffsetBy(gtx, image.Pt(x, y), func() {
			props.text.op.Add(gtx.Ops)
		})
	}
}
