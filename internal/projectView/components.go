package projectview

import (
	"image"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

func drawInfoRow(th *theme.RepeatTheme, left, right string) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		flexDims := layout.Flex{Spacing: layout.SpaceBetween}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				leftStyle := material.Label(th.Theme, 14, left)
				leftStyle.Font.Weight = font.Bold
				return leftStyle.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Width: 5}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				rightStyle := material.Label(th.Theme, 14, right)
				rightStyle.WrapPolicy = text.WrapWords
				return rightStyle.Layout(gtx)
			}),
		)
		common.OffsetBy(gtx, image.Pt(0, flexDims.Size.Y), func(gtx layout.Context) {
			gtx.Constraints.Max.X = gtx.Constraints.Min.X
			common.DrawDivider(gtx, th, common.DividerProps{})
		})
		return flexDims
	})
}

type infoListStyle struct {
	Inset   layout.Inset
	Title   string
	HRowGap unit.Dp
	th      *theme.RepeatTheme
}

func infoList(th *theme.RepeatTheme, title string) infoListStyle {
	return infoListStyle{
		th:      th,
		Title:   title,
		Inset:   layout.UniformInset(15),
		HRowGap: 10,
	}
}

func (i infoListStyle) layout(gtx layout.Context, rows ...layout.FlexChild) layout.Dimensions {
	maxW := gtx.Constraints.Min.X
	dims := common.DrawBox(gtx, common.Box{
		Size:  image.Rect(0, 0, maxW, gtx.Constraints.Min.Y),
		Color: i.th.Bg,
		R:     theme.CornerR(15, 15, 15, 15),
	})
	i.Inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				titleStyle := material.Label(i.th.Theme, 16, i.Title)
				titleStyle.Alignment = text.Middle
				gtx.Constraints.Min.X = maxW
				return titleStyle.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: i.HRowGap}.Layout),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				xInset := i.Inset.Left + i.Inset.Right
				gtx.Constraints.Min.X -= gtx.Dp(xInset)
				gtx.Constraints.Max.X = gtx.Constraints.Min.X
				return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceBetween}.Layout(gtx, rows...)
			}),
		)
	})
	return dims
}
