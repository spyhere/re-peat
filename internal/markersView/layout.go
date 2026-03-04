package markersview

import (
	"fmt"
	"image"
	"time"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	micons "github.com/spyhere/re-peat/internal/mIcons"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
)

var topM = 140

var interval = 250 * time.Millisecond

func (m *MarkersView) Layout(gtx layout.Context) layout.Dimensions {
	m.dispatch(gtx)
	isPlaying := m.p.IsPlaying()
	if isPlaying {
		m.listenToPlayerUpdates()
		gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(interval)})
	} else {
		m.pausePlaying()
	}
	common.DrawBackground(gtx, m.th.Palette.MarkersViewBg)

	var searchDims layout.Dimensions
	common.OffsetBy(gtx, image.Pt(0, topM), func() {
		common.CenteredX(gtx, func() layout.Dimensions {
			searchDims = common.DrawSearch(gtx, m.th, common.SProps{
				DefaultText: "Название маркера...",
				Searchable:  m.searchable,
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
		m.table.Rows = len(*m.timeMarkers)

		m.table.HeadCells(
			func(gtx layout.Context) layout.Dimensions {
				txt := material.Body2(m.th.Theme, "№")
				txt.Font.Weight = font.Bold
				return txt.Layout(gtx)
			},
			func(gtx layout.Context) layout.Dimensions {
				if m.replayButton.Clicked(gtx) {
					m.replayMarkers()
				}
				if m.replayButton.Hovered() {
					common.SetCursor(gtx, pointer.CursorPointer)
				}
				icon := micons.Replay
				if isPlaying {
					icon = micons.Pause
				}
				return drawClickableIcon(gtx, icon, 24, m.th.Palette.Backdrop, m.replayButton)
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
				if m.tagButton.Clicked(gtx) {
					m.openTagsFilter()
				}
				if m.tagButton.Hovered() {
					common.SetCursor(gtx, pointer.CursorPointer)
				}
				iconDims := drawClickableIcon(gtx, micons.Filter, 24, m.th.Palette.Backdrop, m.tagButton)

				gtx.Constraints.Min = image.Point{}
				txt := material.Body2(m.th.Theme, "Tags")
				txt.Font.Weight = font.Bold
				var textDim layout.Dimensions
				common.OffsetBy(gtx, image.Pt(iconDims.Size.X, 0), func() {
					textDim = txt.Layout(gtx)
				})
				return layout.Dimensions{Size: image.Pt(iconDims.Size.X+textDim.Size.X, textDim.Size.Y)}
			},
			func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{}
			},
			func(gtx layout.Context) layout.Dimensions {
				if m.deleteButton.Clicked(gtx) {
					m.deleteMarkers()
				}
				if m.deleteButton.Hovered() {
					common.SetCursor(gtx, pointer.CursorPointer)
				}
				return common.DrawIconButton(gtx, common.IconButtonProps{
					Icon: micons.Delete,
					Bg:   m.th.Palette.IconButton.Enabled.Bg,
					Fg:   m.th.Palette.IconButton.Enabled.Icon,
					Cl:   m.deleteButton,
				})
			},
		)

		m.table.RowCells(
			func(gtx layout.Context, rowIdx int, curMarker *tm.TimeMarker) layout.Dimensions {
				txt := material.Body2(m.th.Theme, fmt.Sprintf("%02d", rowIdx+1))
				return txt.Layout(gtx)
			},
			func(gtx layout.Context, rowIdx int, curMarker *tm.TimeMarker) layout.Dimensions {
				if curMarker.Play.Clicked(gtx) {
					m.toggleMarker(curMarker)
				}
				if curMarker.Play.Hovered() {
					common.SetCursor(gtx, pointer.CursorPointer)
				}
				icon := micons.Play
				if m.isThisMarkerPlaying(curMarker) {
					icon = micons.Pause
				}
				return drawClickableIcon(gtx, icon, 26, m.th.Palette.Backdrop, curMarker.Play)
			},
			func(gtx layout.Context, rowIdx int, curMarker *tm.TimeMarker) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				txt := material.Body2(m.th.Theme, (*m.timeMarkers).Get(rowIdx, true).Name)
				return txt.Layout(gtx)
			},
			func(gtx layout.Context, rowIdx int, curMarker *tm.TimeMarker) layout.Dimensions {
				curPcm := (*m.timeMarkers).Get(rowIdx, true).Pcm
				formattedSeconds := common.FormatSeconds(m.audio.GetSecondsFromSamples(m.audio.GetSamplesFromPCM(curPcm)))
				txt := material.Body2(m.th.Theme, formattedSeconds)
				return txt.Layout(gtx)
			},
			func(gtx layout.Context, rowIdx int, curMarker *tm.TimeMarker) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				tagsArr := curMarker.CategoryTags
				if len(tagsArr) == 0 {
					return layout.Dimensions{}
				}
				return curMarker.List.Layout(gtx, len(tagsArr)+len(tagsArr)-1, func(gtx layout.Context, index int) layout.Dimensions {
					if index%2 != 0 {
						return layout.Dimensions{Size: image.Pt(gtx.Dp(5), 0)}
					}
					dim := common.DrawChip(gtx, m.th, common.ChipProps{
						Text: tagsArr[index/2],
					})
					return dim
				})
			},
			func(gtx layout.Context, rowIdx int, curMarker *tm.TimeMarker) layout.Dimensions {
				if curMarker.Edit.Clicked(gtx) {
					fmt.Println("Edit", rowIdx)
				}
				if curMarker.Edit.Hovered() {
					common.SetCursor(gtx, pointer.CursorPointer)
				}
				return drawClickableIcon(gtx, micons.Edit, 24, m.th.Palette.Backdrop, curMarker.Edit)
			},
			func(gtx layout.Context, rowIdx int, curMarker *tm.TimeMarker) layout.Dimensions {
				if curMarker.Delete.Clicked(gtx) {
					curMarker.MarkDead()
				}
				if curMarker.Delete.Hovered() {
					common.SetCursor(gtx, pointer.CursorPointer)
				}
				return drawClickableIcon(gtx, micons.Delete, 26, m.th.Palette.Backdrop, curMarker.Delete)
			},
		)
		m.table.Layout(gtx, m.th, []int{4, 4, 30, 6, 46, 4, 6})
	})

	if cursor, ok := m.searchable.GetCursorType(); ok {
		common.SetCursor(gtx, cursor)
	}
	m.updateDefferedState()
	return layout.Dimensions{}
}
