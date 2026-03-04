package markersview

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/spyhere/re-peat/internal/common"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

func drawClickableIcon(gtx layout.Context, icon *widget.Icon, iconSize unit.Dp, iconC color.NRGBA, cl *widget.Clickable) layout.Dimensions {
	iconS := gtx.Dp(iconSize)
	gtx.Constraints.Min.X = iconS
	iconSizeHalf := iconS / 2
	common.DrawBox(gtx, common.Box{
		Size:      image.Rect(0, 0, iconS, iconS),
		R:         theme.CornerR(iconSizeHalf, iconSizeHalf, iconSizeHalf, iconSizeHalf),
		Clickable: cl,
	})
	return icon.Layout(gtx, iconC)
}
