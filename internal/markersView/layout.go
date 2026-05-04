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
	"github.com/spyhere/re-peat/fonts"
	"github.com/spyhere/re-peat/internal/common"
	micons "github.com/spyhere/re-peat/internal/mIcons"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
)

var (
	topM           = 140
	redrawInterval = 50 * time.Millisecond
)

func (m *MarkersView) Layout(gtx layout.Context) layout.Dimensions {
	if m.isDisabled() {
		gtx = gtx.Disabled()
	}
	m.dispatch(gtx)
	m.dialogUpdate()
	isPlaying := m.HasAudioLoaded() && m.Player.IsPlaying()
	if isPlaying {
		m.listenToPlayerUpdates()
		gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(redrawInterval)})
	} else if m.HasAudioLoaded() {
		m.pausePlaying()
	}
	common.DrawBackground(gtx, m.Th.Palette.MarkersViewBg)

	var searchDims layout.Dimensions
	common.OffsetBy(gtx, image.Pt(0, topM), func(gtx layout.Context) {
		common.CenteredX(gtx, func() layout.Dimensions {
			searchDims = common.DrawSearch(gtx, m.Th, common.SProps{
				DefaultText: m.I18n.Markers.SearchBPlaceholder,
				Inputable:   m.searchbar,
				Disabled:    m.TimeMarkers.IsEmpty(),
			})
			return searchDims
		})
	})

	cl := &m.createCl
	if !gtx.Enabled() {
		cl = &m.disabledCl
	}
	if m.isCreateButtonEnabled() {
		drawAddMarkerButton(gtx, m.Th, cl, gtx.Constraints.Max.X/4, topM+searchDims.Size.Y/2)
	}

	common.OffsetBy(gtx, image.Pt(0, topM+searchDims.Size.Y+20), func(gtx layout.Context) {
		common.DrawDivider(gtx, m.Th, common.DividerProps{
			Inset: common.DividerMiddleInset,
		})
	})

	marginX := gtx.Dp(20)
	common.OffsetBy(gtx, image.Pt(marginX, topM+searchDims.Size.Y+50), func(gtx layout.Context) {
		gtx.Constraints.Max.X -= marginX * 2
		gtx.Constraints.Max.Y -= topM + searchDims.Size.Y + 50

		m.table.HeadCells(
			func(gtx layout.Context) layout.Dimensions {
				txt := material.Body2(m.Th.Theme, "№")
				txt.Font.Weight = font.Bold
				return txt.Layout(gtx)
			},
			func(gtx layout.Context) layout.Dimensions {
				if m.replayCl.Clicked(gtx) {
					m.replayMarkers()
				}
				if m.replayCl.Hovered() {
					common.SetCursor(gtx, pointer.CursorPointer)
				}
				icon := micons.Replay
				if isPlaying {
					icon = micons.Pause
				}
				return drawClickableIcon(gtx, m.Th, clickableIconProps{
					icon:     icon,
					iconSize: 24,
					cl:       &m.replayCl,
				})
			},
			func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				txt := material.Body2(m.Th.Theme, m.I18n.Generic.Name)
				txt.Font.Weight = font.Bold
				return txt.Layout(gtx)
			},
			func(gtx layout.Context) layout.Dimensions {
				txt := material.Body2(m.Th.Theme, m.I18n.Generic.Time)
				txt.Font.Weight = font.Bold
				return txt.Layout(gtx)
			},
			func(gtx layout.Context) layout.Dimensions {
				if m.tagCl.Clicked(gtx) {
					m.openTagsFilterDialog()
				}
				if m.tagClearCl.Clicked(gtx) {
					m.clearTagFilter()
				}
				if m.tagCl.Hovered() || m.tagClearCl.Hovered() {
					common.SetCursor(gtx, pointer.CursorPointer)
				}
				var gap unit.Dp = 5
				gtx.Constraints.Min = image.Point{}
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return drawClickableIcon(gtx, m.Th, clickableIconProps{
							icon:     micons.Filter,
							iconSize: 24,
							cl:       &m.tagCl,
							disabled: len(m.ChipsFilter.All) == 0,
						})
					}),
					layout.Rigid(layout.Spacer{Width: gap}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min = image.Point{}
						txt := material.Body2(m.Th.Theme, m.I18n.Generic.Tags)
						txt.Font.Weight = font.Bold
						return txt.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: gap}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if len(m.ChipsFilter.Enabled) == 0 {
							return layout.Dimensions{}
						}
						return drawClickableIcon(gtx, m.Th, clickableIconProps{
							icon:     micons.Cancel,
							iconSize: 24,
							cl:       &m.tagClearCl,
						})
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						enabledChips := m.ChipsFilter.GetEnabledChips()
						inset := layout.Inset{Left: 2, Right: 2}
						return m.enabledTagsLs.Layout(gtx, len(enabledChips), func(gtx layout.Context, index int) layout.Dimensions {
							return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return common.DrawChip(gtx, m.Th, common.ChipProps{
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
				return layout.Dimensions{}
			},
			func(gtx layout.Context) layout.Dimensions {
				if m.deleteCl.Clicked(gtx) {
					m.openDeleteAllDialog()
				}
				if m.deleteCl.Hovered() {
					common.SetCursor(gtx, pointer.CursorPointer)
				}
				return common.DrawIconButton(gtx, common.IconButtonProps{
					Icon:  micons.Delete,
					Th:    m.Th,
					Cl:    &m.deleteCl,
					IsOff: m.TimeMarkers.IsEmpty(),
				})
			},
		)

		m.table.RowCells(
			func(gtx layout.Context, rowIdx int, curMarker *tm.TimeMarker) layout.Dimensions {
				rowNum := fmt.Sprintf("%02d", rowIdx+1)
				curInput := string(m.hotKeyBuf)
				txt := material.Body2(m.Th.Theme, rowNum)
				txt.Font = fonts.GoMedium(font.Medium, font.Regular)
				dims := txt.Layout(gtx)
				if strings.HasPrefix(rowNum, curInput) {
					var highlightTDim layout.Dimensions
					macro, highlightTDim := common.MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
						highlightT := material.Body2(m.Th.Theme, curInput)
						highlightT.Color = m.Th.Palette.Selection.Fg
						highlightT.TextSize += 2
						return highlightT.Layout(gtx)
					})
					common.DrawBox(gtx, common.Box{
						Size:  image.Rect(0, 0, highlightTDim.Size.X, highlightTDim.Size.Y),
						Color: m.Th.Palette.Selection.Bg,
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
				return drawClickableIcon(gtx, m.Th, clickableIconProps{
					icon:     icon,
					iconSize: 26,
					cl:       &curMarker.Play,
				})
			},
			func(gtx layout.Context, rowIdx int, curMarker *tm.TimeMarker) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				txt := material.Body2(m.Th.Theme, m.TimeMarkers.Get(rowIdx, true).Name)
				txt.Font = fonts.GoMedium(font.Medium, font.Regular)
				return txt.Layout(gtx)
			},
			func(gtx layout.Context, rowIdx int, curMarker *tm.TimeMarker) layout.Dimensions {
				currSamples := m.TimeMarkers.Get(rowIdx, true).Samples
				formattedSeconds := common.FormatSeconds(m.AudioMeta.GetSecondsFromSamples(currSamples))
				txt := material.Body2(m.Th.Theme, formattedSeconds)
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
					dim := common.DrawChip(gtx, m.Th, common.ChipProps{
						Text: tagsArr[index/2],
					})
					return dim
				})
			},
			func(gtx layout.Context, rowIdx int, curMarker *tm.TimeMarker) layout.Dimensions {
				if curMarker.Comment.Clicked(gtx) {
					m.openCommentDialog(curMarker)
				}
				if curMarker.Comment.Hovered() {
					common.SetCursor(gtx, pointer.CursorPointer)
				}
				hasNotes := curMarker.Notes != ""
				icon := micons.Comment
				if hasNotes {
					icon = micons.CommentInsert
				}
				return drawClickableIcon(gtx, m.Th, clickableIconProps{
					icon:     icon,
					iconSize: 24,
					cl:       &curMarker.Comment,
				})
			},
			func(gtx layout.Context, rowIdx int, curMarker *tm.TimeMarker) layout.Dimensions {
				if curMarker.Edit.Clicked(gtx) {
					m.openMarkerDialog(curMarker, edit, m.I18n.Markers.MEdit)
				}
				if curMarker.Edit.Hovered() {
					common.SetCursor(gtx, pointer.CursorPointer)
				}
				return drawClickableIcon(gtx, m.Th, clickableIconProps{
					icon:     micons.Edit,
					iconSize: 24,
					cl:       &curMarker.Edit,
				})
			},
			func(gtx layout.Context, rowIdx int, curMarker *tm.TimeMarker) layout.Dimensions {
				if curMarker.Delete.Clicked(gtx) {
					curMarker.MarkDead()
				}
				if curMarker.Delete.Hovered() {
					common.SetCursor(gtx, pointer.CursorPointer)
				}
				return drawClickableIcon(gtx, m.Th, clickableIconProps{
					icon:     micons.Delete,
					iconSize: 24,
					cl:       &curMarker.Delete,
				})
			},
		)
		m.table.Layout(gtx, m.Th, len(m.TimeMarkers), []int{4, 4, 30, 6, 44, 4, 4, 4})
	})

	if isPlaying {
		m.pc.totalS = m.AudioMeta.Seconds
		m.pc.setVolume(m.Player.GetVolume())
		m.pc.currentSec = m.Player.GetCurrentSecond()
		playerState(m.Th, &m.pc).Layout(gtx)
		if cursor, ok := m.pc.getCursorType(); ok {
			common.SetCursor(gtx, cursor)
		}
	}

	m.fm.PlaceScrim(gtx)
	if cursor, ok := m.searchbar.GetCursorType(); ok {
		common.SetCursor(gtx, cursor)
	}
	m.updateDefferedState()
	return layout.Dimensions{}
}
