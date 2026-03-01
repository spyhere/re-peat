package markersview

import (
	"fmt"
	"image"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	micons "github.com/spyhere/re-peat/internal/mIcons"
)

var searchable = common.Searchable{}
var topM = 140

var table = common.NewTable(common.TableProps{
	Axis:      layout.Vertical,
	ColumsNum: 7,
	HeaderCellsAlignment: []layout.Direction{
		layout.Center,
		layout.Center,
		layout.W,
		layout.Center,
		layout.W,
		layout.Center,
		layout.Center,
	},
	RowCellsAlignment: []layout.Direction{
		layout.Center,
		layout.Center,
		layout.W,
		layout.Center,
		layout.Center,
		layout.Center,
		layout.Center,
	},
})

func (m *MarkersView) Layout(gtx layout.Context) layout.Dimensions {
	common.DrawBackground(gtx, m.th.Palette.MarkersViewBg)

	var searchDims layout.Dimensions
	common.OffsetBy(gtx, image.Pt(0, topM), func() {
		common.CenteredX(gtx, func() layout.Dimensions {
			searchDims = common.DrawSearch(gtx, m.th, common.SProps{
				DefaultText: "Название маркера...",
				Searchable:  &searchable,
			})
			return searchDims
		})
	})
	common.OffsetBy(gtx, image.Pt(0, topM+searchDims.Size.Y+20), func() {
		common.DrawDivider(gtx, m.th, common.DividerProps{
			Inset: common.DividerMiddleInset,
		})
	})

	marginX := gtx.Dp(20)
	common.OffsetBy(gtx, image.Pt(marginX, topM+searchDims.Size.Y+50), func() {
		gtx.Constraints.Max.X -= marginX * 2
		gtx.Constraints.Max.Y -= topM + searchDims.Size.Y + 50
		table.Rows = len(*m.timeMarkers)
		table.HeadCells(
			func(gtx layout.Context) layout.Dimensions {
				txt := material.Body2(m.th.Theme, "№")
				txt.Font.Weight = font.Bold
				return txt.Layout(gtx)
			},
			func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{}
			},
			func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				txt := material.Body2(m.th.Theme, "Name")
				txt.Font.Weight = font.Bold
				return txt.Layout(gtx)
			},
			func(gtx layout.Context) layout.Dimensions {
				txt := material.Body2(m.th.Theme, "Time")
				txt.Font.Weight = font.Bold
				return txt.Layout(gtx)
			},
			func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				txt := material.Body2(m.th.Theme, "Tags")
				txt.Font.Weight = font.Bold
				return txt.Layout(gtx)
			},
			func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{}
			},
			func(gtx layout.Context) layout.Dimensions {
				// Delete button
				return layout.Dimensions{}
			},
		)
		table.RowCells(
			func(gtx layout.Context, rowIdx, colIdx int) layout.Dimensions {
				txt := material.Body2(m.th.Theme, fmt.Sprintf("%02d", rowIdx+1))
				return txt.Layout(gtx)
			},
			func(gtx layout.Context, rowIdx, colIdx int) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Dp(24)
				return micons.Play.Layout(gtx, m.th.Palette.Backdrop)
			},
			func(gtx layout.Context, rowIdx, colIdx int) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				txt := material.Body2(m.th.Theme, (*m.timeMarkers).GetAsc(rowIdx).Name)
				return txt.Layout(gtx)
			},
			func(gtx layout.Context, rowIdx, colIdx int) layout.Dimensions {
				txt := material.Body2(m.th.Theme, fmt.Sprint((*m.timeMarkers).GetAsc(rowIdx).Pcm))
				return txt.Layout(gtx)
			},
			func(gtx layout.Context, rowIdx, colIdx int) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				return common.DrawBox(gtx, common.Box{
					Size:  image.Rectangle(gtx.Constraints),
					Color: m.th.Palette.SegButtons.Disabled.Selected,
				})
			},
			func(gtx layout.Context, rowIdx, colIdx int) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				return common.DrawBox(gtx, common.Box{
					Size:  image.Rectangle(gtx.Constraints),
					Color: m.th.Palette.Editor.Bg,
				})
			},
			func(gtx layout.Context, rowIdx, colIdx int) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				return common.DrawBox(gtx, common.Box{
					Size:  image.Rectangle(gtx.Constraints),
					Color: m.th.Palette.Editor.SoundWave,
				})
			},
		)
		table.Layout(gtx, m.th, []int{4, 4, 31, 6, 46, 3, 6})
	})

	// TODO: Move this to Searchable
	if searchable.IsHovered() {
		if searchable.IsFocused() {
			common.SetCursor(gtx, pointer.CursorText)
		} else {
			common.SetCursor(gtx, pointer.CursorPointer)
		}
	}
	if searchable.Cancel.Hovered() {
		common.SetCursor(gtx, pointer.CursorPointer)
	}
	return layout.Dimensions{}
}
