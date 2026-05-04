package projectview

import (
	"image"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	micons "github.com/spyhere/re-peat/internal/mIcons"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

const (
	columnMar   unit.Dp = 40
	columnW             = 30.0
	columnWMax  unit.Dp = 400
	columnH             = 38.0
	columnHMax  unit.Dp = 270
	titleCtaGap unit.Dp = 30
	CtaListGap  unit.Dp = 20
	ListCtaGap  unit.Dp = 20
	CtaGap      unit.Dp = 20
)

func (pv *ProjectView) Layout(gtx layout.Context) layout.Dimensions {
	if pv.isDisabled() {
		gtx = gtx.Disabled()
	}
	pv.dispatch(gtx)

	common.DrawBackground(gtx, pv.Th.Palette.Project.Bg)
	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				return common.DrawBox(gtx, common.Box{
					Size:    image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Min.Y),
					Color:   pv.Th.Palette.Project.CargBg,
					R:       theme.CornerR(25, 25, 25, 25),
					StrokeC: pv.Th.Palette.Project.CardStroke,
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
							titleSt := material.H4(pv.Th.Theme, pv.I18n.Generic.Audio)
							titleSt.Alignment = text.Middle
							gtx.Constraints.Min.X = tableW
							return titleSt.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: titleCtaGap}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = tableW
							return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								btn := material.IconButton(pv.Th.Theme, &pv.audioLoadCl, micons.Folder, "Load")
								btn.Background = pv.Th.Palette.Project.LoadButtonBg
								return btn.Layout(gtx)
							})
						}),
						layout.Rigid(layout.Spacer{Height: CtaListGap}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = tableW
							gtx.Constraints.Min.Y = tableH
							gtx.Constraints.Max.Y = tableH
							fMeta := pv.AFileMeta
							aMeta := pv.AudioMeta
							return infoList(pv.Th, fMeta.Name).layout(gtx,
								drawInfoRow(pv.Th, pv.I18n.Generic.Length, aMeta.SecondsString()),
								drawInfoRow(pv.Th, pv.I18n.Generic.Size, fMeta.SizeString()),
								drawInfoRow(pv.Th, pv.I18n.Generic.AudioChannels, aMeta.ChannelsString(pv.I18n)),
								drawInfoRow(pv.Th, pv.I18n.Generic.SampleRate, aMeta.SampleRateString()),
								drawInfoRow(pv.Th, pv.I18n.Generic.Modified, fMeta.UpdatedAtString()),
							)
						}),
					)
				})

				common.OffsetBy(gtx, image.Pt(audioDims.Size.X, 0), func(gtx layout.Context) {
					gtx.Constraints.Max.Y = audioDims.Size.Y
					common.DrawDivider(gtx, pv.Th, common.DividerProps{
						Axis:  common.Vertical,
						Inset: common.DividerMiddleInset,
					})
				})

				gtx.Constraints.Max.Y = audioDims.Size.Y
				common.OffsetBy(gtx, image.Pt(audioDims.Size.X, 0), func(gtx layout.Context) {
					layout.UniformInset(columnMar).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								titleSt := material.H4(pv.Th.Theme, pv.I18n.Generic.Markers)
								titleSt.Alignment = text.Middle
								titleSt.Alignment = text.Middle
								gtx.Constraints.Min.X = tableW
								return titleSt.Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Height: titleCtaGap}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = tableW
								return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									cl := &pv.markersLoadCl
									if !pv.HasAudioLoaded() {
										cl = &pv.disabledCl
										gtx = gtx.Disabled()
									}
									btn := material.IconButton(pv.Th.Theme, cl, micons.Folder, "Load")
									btn.Background = pv.Th.Palette.Project.LoadButtonBg
									return btn.Layout(gtx)
								})
							}),
							layout.Rigid(layout.Spacer{Height: CtaListGap}.Layout),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = tableW
								gtx.Constraints.Max.Y = gtx.Constraints.Min.Y
								return infoList(pv.Th, pv.MFileMeta.Name).layout(gtx,
									drawInfoRow(pv.Th, pv.I18n.Generic.Amount, pv.MarkersMeta.AmountString()),
									drawInfoRow(pv.Th, pv.I18n.Generic.WithComments, pv.MarkersMeta.WithCommentsString()),
									drawInfoRow(pv.Th, pv.I18n.Generic.Size, pv.MFileMeta.SizeString()),
									drawInfoRow(pv.Th, pv.I18n.Generic.Modified, pv.MFileMeta.UpdatedAtString()),
								)
							}),
							layout.Rigid(layout.Spacer{Height: ListCtaGap}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Max.X = tableW
								btnBg := pv.Th.Palette.Project.SaveButtonBg
								btnFg := pv.Th.Palette.Project.SaveButtonFg
								return layout.Flex{}.Layout(gtx,
									layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
										btnStyle := common.Button(pv.Th, &pv.markersSaveCl, micons.Save, pv.I18n.Generic.Save)
										btnStyle.WExpanded = true
										btnStyle.Bg = btnBg
										btnStyle.Fg = btnFg
										btnStyle.Disabled = !pv.HasMarkersLoaded() || pv.TimeMarkers.IsEmpty()
										return btnStyle.Layout(gtx)
									}),
									layout.Rigid(layout.Spacer{Width: CtaGap}.Layout),
									layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
										btnStyle := common.Button(pv.Th, &pv.markersSaveAsCl, micons.Save, pv.I18n.Generic.SaveAs)
										btnStyle.WExpanded = true
										btnStyle.Bg = btnBg
										btnStyle.Fg = btnFg
										btnStyle.Disabled = pv.TimeMarkers.IsEmpty()
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

	if pv.audioLoadCl.Hovered() || pv.markersLoadCl.Hovered() || pv.markersSaveCl.Hovered() || pv.markersSaveAsCl.Hovered() {
		common.SetCursor(gtx, pointer.CursorPointer)
	}
	return layout.Dimensions{}
}
