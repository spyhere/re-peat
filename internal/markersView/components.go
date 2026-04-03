package markersview

import (
	"fmt"
	"image"
	"strings"

	"gioui.org/layout"
	"gioui.org/text"
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
	if !gtx.Enabled() {
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
		iconStyle.Background = th.Palette.Editor.SoundWave
		gtx.Constraints.Min = image.Point{}
		return iconStyle.Layout(gtx)
	})
	common.OffsetBy(gtx, image.Pt(x, y-addIconDims.Size.Y/2), func(gtx layout.Context) {
		addIconM.Add(gtx.Ops)
	})
}

type fieldGroupStyle struct {
	fieldsYMargin unit.Dp
	fieldsXMargin unit.Dp
	fieldW        unit.Dp
	gap           unit.Dp
}

func defaultFieldGroupStyle() fieldGroupStyle {
	return fieldGroupStyle{
		fieldsYMargin: 10,
		fieldsXMargin: 10,
		fieldW:        270,
		gap:           20,
	}
}

type playerStateStyle struct {
	yOffset   unit.Dp
	bgW       unit.Dp
	bgH       unit.Dp
	bgShape   int
	lineH     unit.Dp
	lineShape int
	gapY      unit.Dp
	thumbDiam unit.Dp
}

func defaultPlayerStateStyle() playerStateStyle {
	return playerStateStyle{
		yOffset:   90,
		bgW:       200,
		bgH:       70,
		bgShape:   10,
		lineH:     3,
		lineShape: 5,
		gapY:      12,
		thumbDiam: 16,
	}
}

func drawPlayerState(gtx layout.Context, th *theme.RepeatTheme, curS float64, totalS float64) {
	var timeLabel strings.Builder
	fmt.Fprint(&timeLabel, common.FormatSeconds(curS))
	timeLabel.WriteString(" / ")
	fmt.Fprint(&timeLabel, common.FormatSeconds(totalS))

	s := defaultPlayerStateStyle()
	lineH := gtx.Dp(s.lineH)
	common.OffsetBy(gtx, image.Pt(0, gtx.Constraints.Max.Y-gtx.Dp(s.yOffset)), func(gtx layout.Context) {
		common.CenteredX(gtx, func() layout.Dimensions {

			// Bg
			bgDims := common.DrawBox(gtx, common.Box{
				Size:  image.Rect(0, 0, gtx.Dp(s.bgW), gtx.Dp(s.bgH)),
				Color: th.Palette.Backdrop,
				R:     theme.CornerR(s.bgShape, s.bgShape, s.bgShape, s.bgShape),
			})
			gtx.Constraints.Min = bgDims.Size
			gtx.Constraints.Max = bgDims.Size
			layout.UniformInset(10).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,

					// Time
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						txtStyle := material.H5(th.Theme, timeLabel.String())
						txtStyle.Alignment = text.Middle
						txtStyle.Color = th.Bg
						return txtStyle.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: s.gapY}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {

						// Line
						lineDims := common.DrawBox(gtx, common.Box{
							Size:  image.Rect(0, 0, gtx.Constraints.Max.X, lineH),
							Color: th.Bg,
							R:     theme.CornerR(s.lineShape, s.lineShape, s.lineShape, s.lineShape),
						})

						// Thumb
						xOffset := int(curS) * gtx.Constraints.Max.X / int(totalS)
						thumbDiam := gtx.Dp(s.thumbDiam)
						thumbRadi := thumbDiam / 2
						common.OffsetBy(gtx, image.Pt(xOffset-thumbRadi, -thumbDiam/2+lineH/2), func(gtx layout.Context) {
							common.DrawBox(gtx, common.Box{
								Size:  image.Rect(0, 0, thumbDiam, thumbDiam),
								Color: th.Bg,
								R:     theme.CornerR(thumbRadi, thumbRadi, thumbRadi, thumbRadi),
							})
						})
						return lineDims
					}),
				)
			})
			return bgDims
		})
	})
}
