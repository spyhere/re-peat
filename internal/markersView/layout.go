package markersview

import (
	"fmt"
	"image"
	"strings"
	"time"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	micons "github.com/spyhere/re-peat/internal/mIcons"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
)

var topM = 140

var interval = 250 * time.Millisecond

func (m *MarkersView) Layout(gtx layout.Context) layout.Dimensions {
	m.dispatch(gtx)
	m.dialogUpdate(gtx)
	isPlaying := m.p.IsPlaying()
	if isPlaying {
		m.listenToPlayerUpdates()
		gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(interval)})
	} else {
		m.pausePlaying()
	}
	common.DrawBackground(gtx, m.th.Palette.MarkersViewBg)

	var searchDims layout.Dimensions
	common.OffsetBy(gtx, image.Pt(0, topM), func(gtx layout.Context) {
		common.CenteredX(gtx, func() layout.Dimensions {
			searchDims = common.DrawSearch(gtx, m.th, common.SProps{
				DefaultText: "Название маркера...",
				Inputable:   m.searchbar,
			})
			return searchDims
		})
	})

	drawAddMarkerButton(gtx, m.th, m.createButton, gtx.Constraints.Max.X/4, topM+searchDims.Size.Y/2)

	common.OffsetBy(gtx, image.Pt(0, topM+searchDims.Size.Y+20), func(gtx layout.Context) {
		common.DrawDivider(gtx, m.th, common.DividerProps{
			Inset: common.DividerMiddleInset,
		})
	})

	marginX := gtx.Dp(20)
	common.OffsetBy(gtx, image.Pt(marginX, topM+searchDims.Size.Y+50), func(gtx layout.Context) {
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
				return drawClickableIcon(gtx, m.th, clickableIconProps{
					icon:     icon,
					iconSize: 24,
					cl:       m.replayButton,
				})
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
					m.openTagsFilterDialog()
				}
				if m.tagButton.Hovered() {
					common.SetCursor(gtx, pointer.CursorPointer)
				}
				var gap unit.Dp = 5
				gtx.Constraints.Min = image.Point{}
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return drawClickableIcon(gtx, m.th, clickableIconProps{
							icon:     micons.Filter,
							iconSize: 24,
							cl:       m.tagButton,
							disabled: len(m.chipsFilter.all) == 0,
						})
					}),
					layout.Rigid(layout.Spacer{Width: gap}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min = image.Point{}
						txt := material.Body2(m.th.Theme, "Tags")
						txt.Font.Weight = font.Bold
						return txt.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: gap}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						enabledChips := m.chipsFilter.getEnabledChips()
						inset := layout.Inset{Left: 2, Right: 2}
						return m.enabledTagsLs.Layout(gtx, len(enabledChips), func(gtx layout.Context, index int) layout.Dimensions {
							return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return common.DrawChip(gtx, m.th, common.ChipProps{
									Text:     enabledChips[index],
									Selected: true,
									HideIcon: true,
								})
							})
						})
					}),
				)
			},
			func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{}
			},
			func(gtx layout.Context) layout.Dimensions {
				if m.deleteButton.Clicked(gtx) {
					m.openDeleteAllDialog()
				}
				if m.deleteButton.Hovered() {
					common.SetCursor(gtx, pointer.CursorPointer)
				}
				return common.DrawIconButton(gtx, common.IconButtonProps{
					Icon:  micons.Delete,
					Th:    m.th,
					Cl:    m.deleteButton,
					IsOff: len(*m.timeMarkers) == 0,
				})
			},
		)

		m.table.RowCells(
			func(gtx layout.Context, rowIdx int, curMarker *tm.TimeMarker) layout.Dimensions {
				rowNum := fmt.Sprintf("%02d", rowIdx+1)
				curInput := string(m.hotKeyBuf)
				txt := material.Body2(m.th.Theme, rowNum)
				dims := txt.Layout(gtx)
				if strings.HasPrefix(rowNum, curInput) {
					var highlightTDim layout.Dimensions
					macro, highlightTDim := common.MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
						highlightT := material.Body2(m.th.Theme, curInput)
						highlightT.Color = m.th.Palette.Selection.Fg
						highlightT.TextSize += 2
						return highlightT.Layout(gtx)
					})
					common.DrawBox(gtx, common.Box{
						Size:  image.Rect(0, 0, highlightTDim.Size.X, highlightTDim.Size.Y),
						Color: m.th.Palette.Selection.Bg,
					})
					macro.Add(gtx.Ops)
				}
				return dims
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
				return drawClickableIcon(gtx, m.th, clickableIconProps{
					icon:     icon,
					iconSize: 26,
					cl:       curMarker.Play,
				})
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
					m.openMarkerDialog(curMarker, edit, "Marker Edit")
				}
				if curMarker.Edit.Hovered() {
					common.SetCursor(gtx, pointer.CursorPointer)
				}
				return drawClickableIcon(gtx, m.th, clickableIconProps{
					icon:     micons.Edit,
					iconSize: 24,
					cl:       curMarker.Edit,
				})
			},
			func(gtx layout.Context, rowIdx int, curMarker *tm.TimeMarker) layout.Dimensions {
				if curMarker.Delete.Clicked(gtx) {
					curMarker.MarkDead()
				}
				if curMarker.Delete.Hovered() {
					common.SetCursor(gtx, pointer.CursorPointer)
				}
				return drawClickableIcon(gtx, m.th, clickableIconProps{
					icon:     micons.Delete,
					iconSize: 24,
					cl:       curMarker.Delete,
				})
			},
		)
		m.table.Layout(gtx, m.th, []int{4, 4, 30, 6, 46, 4, 6})
	})

	m.fm.PlaceScrim(gtx)
	if cursor, ok := m.searchbar.GetCursorType(); ok {
		common.SetCursor(gtx, cursor)
	}
	m.updateDefferedState()
	return layout.Dimensions{}
}
