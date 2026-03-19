package markersview

import (
	"image"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	micons "github.com/spyhere/re-peat/internal/mIcons"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

type clickableIconProps struct {
	icon     *widget.Icon
	iconSize unit.Dp
	cl       *widget.Clickable
	disabled bool
}

// TODO: Looks like this should be a part of common DrawIconButton
func drawClickableIcon(gtx layout.Context, th *theme.RepeatTheme, props clickableIconProps) layout.Dimensions {
	iconS := gtx.Dp(props.iconSize)
	gtx.Constraints.Min.X = iconS
	iconSizeHalf := iconS / 2
	cl := props.cl
	color := th.Palette.Backdrop
	if props.disabled {
		color = th.Palette.IconButton.Disabled.Bg
		cl = nil
	}
	common.DrawBox(gtx, common.Box{
		Size:      image.Rect(0, 0, iconS, iconS),
		R:         theme.CornerR(iconSizeHalf, iconSizeHalf, iconSizeHalf, iconSizeHalf),
		Clickable: cl,
	})
	return props.icon.Layout(gtx, color)
}

func drawAddMarkerButton(gtx layout.Context, th *theme.RepeatTheme, cl *widget.Clickable, x, y int) {
	addIconM, addIconDims := common.MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
		iconStyle := material.IconButton(th.Theme, cl, micons.ContentAddCircle, "")
		iconStyle.Size = 20
		iconStyle.Inset = layout.UniformInset(7)
		gtx.Constraints.Min = image.Point{}
		return iconStyle.Layout(gtx)
	})
	common.OffsetBy(gtx, image.Pt(x, y-addIconDims.Size.Y/2), func(gtx layout.Context) {
		addIconM.Add(gtx.Ops)
	})
}
