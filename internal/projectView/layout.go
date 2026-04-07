package projectview

import (
	"image"

	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	micons "github.com/spyhere/re-peat/internal/mIcons"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

const (
	columnMar   unit.Dp = 40
	columnW             = "30%"
	columnWMax  unit.Dp = 400
	columnH             = "38%"
	columnHMax  unit.Dp = 270
	titleCtaGap unit.Dp = 30
	CtaListGap  unit.Dp = 20
	ListCtaGap  unit.Dp = 20
	CtaGap      unit.Dp = 20
)

var cl = widget.Clickable{}

// TODO: Pass theme to layout only
func (pv *ProjectView) Layout(gtx layout.Context) layout.Dimensions {
	common.DrawBackground(gtx, pv.th.Palette.Project.Bg)
	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				return common.DrawBox(gtx, common.Box{
					Size:    image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Min.Y),
					Color:   pv.th.Palette.Project.CargBg,
					R:       theme.CornerR(25, 25, 25, 25),
					StrokeC: pv.th.Palette.Project.CardStroke,
					StrokeW: 4,
				})
			}),

			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				tableW := min(common.PrcToPx(gtx.Constraints.Max.X, columnW), gtx.Dp(columnWMax))
				tableH := min(common.PrcToPx(gtx.Constraints.Max.Y, columnH), gtx.Dp(columnHMax))
				var audioDims layout.Dimensions
				audioDims = layout.UniformInset(columnMar).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
								btn := material.IconButton(pv.th.Theme, &cl, micons.Folder, "Load")
								btn.Background = pv.th.Palette.Project.LoadButtonBg
								return btn.Layout(gtx)
							})
						}),
						layout.Rigid(layout.Spacer{Height: CtaListGap}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = tableW
							gtx.Constraints.Min.Y = tableH
							gtx.Constraints.Max.Y = tableH
							return infoList(pv.th, "Test_audio.mp3").layout(gtx,
								drawInfoRow(pv.th, "Length:", "3:42"),
								drawInfoRow(pv.th, "Size:", "3.4Mb"),
								drawInfoRow(pv.th, "Audio Channels:", "Stereo"),
								drawInfoRow(pv.th, "Sample rate:", "44.1 Khz"),
								drawInfoRow(pv.th, "Created:", "Wednesday, 5 April 2017 at 18:39"),
								drawInfoRow(pv.th, "Edited:", "Wednesday, 3 March 2018 at 11:09"),
							)
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
					layout.UniformInset(columnMar).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
									btn := material.IconButton(pv.th.Theme, &cl, micons.Folder, "Load")
									btn.Background = pv.th.Palette.Project.LoadButtonBg
									return btn.Layout(gtx)
								})
							}),
							layout.Rigid(layout.Spacer{Height: CtaListGap}.Layout),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = tableW
								gtx.Constraints.Max.Y = gtx.Constraints.Min.Y
								return infoList(pv.th, "my_markers").layout(gtx,
									drawInfoRow(pv.th, "Amount:", "13"),
									drawInfoRow(pv.th, "Created:", "Wednesday, 5 April 2017 at 18:39"),
									drawInfoRow(pv.th, "Modified:", "Wednesday, 3 March 2018 at 11:09"),
								)
							}),
							layout.Rigid(layout.Spacer{Height: ListCtaGap}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Max.X = tableW
								btnBg := pv.th.Palette.Project.SaveButtonBg
								btnFg := pv.th.Palette.Project.SaveButtonFg
								return layout.Flex{}.Layout(gtx,
									layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
										btnStyle := common.Button(pv.th, &cl, micons.Save, "Save")
										btnStyle.WExpanded = true
										btnStyle.Bg = btnBg
										btnStyle.Fg = btnFg
										return btnStyle.Layout(gtx)
									}),
									layout.Rigid(layout.Spacer{Width: CtaGap}.Layout),
									layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
										btnStyle := common.Button(pv.th, &cl, micons.Save, "Save as")
										btnStyle.WExpanded = true
										btnStyle.Bg = btnBg
										btnStyle.Fg = btnFg
										return btnStyle.Layout(gtx)
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
