package projectview

import (
	"image"

	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

const (
	columnWidth unit.Dp = 20
	columnW             = "30%"
	columnWMax  unit.Dp = 400
	titleCtaGap unit.Dp = 30
	CtaListGap  unit.Dp = 20
	ListCtaGap  unit.Dp = 20
	CtaGap      unit.Dp = 20
)

// TODO: Pass theme to layout only
func (pv *ProjectView) Layout(gtx layout.Context) layout.Dimensions {
	common.DrawBackground(gtx, pv.th.Palette.ProjectViewBg)
	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				return common.DrawBox(gtx, common.Box{
					Size:    image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Min.Y),
					Color:   pv.th.Palette.CardBg,
					R:       theme.CornerR(25, 25, 25, 25),
					StrokeC: pv.th.Palette.ProjectCardS,
					StrokeW: 4,
				})
			}),

			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				tableW := min(common.PrcToPx(gtx.Constraints.Max.X, columnW), gtx.Dp(columnWMax))
				var audioDims layout.Dimensions
				audioDims = layout.UniformInset(columnWidth).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							titleSt := material.H4(pv.th.Theme, "Audio")
							titleSt.Alignment = text.Middle
							gtx.Constraints.Min.X = tableW
							return titleSt.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: titleCtaGap}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = tableW
							return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return common.DrawBox(gtx, common.Box{
									Size:  image.Rect(0, 0, gtx.Dp(100), gtx.Dp(56)),
									Color: pv.th.Palette.Backdrop,
								})
							})
						}),
						layout.Rigid(layout.Spacer{Height: CtaListGap}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return common.DrawBox(gtx, common.Box{
								Size:  image.Rect(0, 0, tableW, 500),
								Color: pv.th.Bg,
							})
						}),
					)
				})

				common.OffsetBy(gtx, image.Pt(audioDims.Size.X, 0), func(gtx layout.Context) {
					gtx.Constraints.Max.Y = audioDims.Size.Y
					common.DrawDivider(gtx, pv.th, common.DividerProps{
						Axis:  common.Vertical,
						Inset: common.DividerMiddleInset,
					})
				})

				gtx.Constraints.Max.Y = audioDims.Size.Y
				common.OffsetBy(gtx, image.Pt(audioDims.Size.X, 0), func(gtx layout.Context) {
					layout.UniformInset(columnWidth).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								titleSt := material.H4(pv.th.Theme, "Markers")
								titleSt.Alignment = text.Middle
								titleSt.Alignment = text.Middle
								gtx.Constraints.Min.X = tableW
								return titleSt.Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Height: titleCtaGap}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = tableW
								return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return common.DrawBox(gtx, common.Box{
										Size:  image.Rect(0, 0, gtx.Dp(100), gtx.Dp(56)),
										Color: pv.th.Palette.Backdrop,
									})
								})
							}),
							layout.Rigid(layout.Spacer{Height: CtaListGap}.Layout),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return common.DrawBox(gtx, common.Box{
									Size:  image.Rect(0, 0, tableW, gtx.Constraints.Min.Y),
									Color: pv.th.Bg,
								})
							}),
							layout.Rigid(layout.Spacer{Height: ListCtaGap}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Max.X = tableW
								return layout.Flex{}.Layout(gtx,
									layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
										return common.DrawBox(gtx, common.Box{
											Size:  image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Dp(56)),
											Color: pv.th.Palette.Backdrop,
										})
									}),
									layout.Rigid(layout.Spacer{Width: CtaGap}.Layout),
									layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
										return common.DrawBox(gtx, common.Box{
											Size:  image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Dp(56)),
											Color: pv.th.Palette.Backdrop,
										})
									}),
								)
							}),
						)
					})
				})
				return layout.Dimensions{Size: image.Pt(audioDims.Size.X*2, audioDims.Size.Y)}
			}),
		)
	})
	return layout.Dimensions{}
}
