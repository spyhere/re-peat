package markersview

import (
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	micons "github.com/spyhere/re-peat/internal/mIcons"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
)

func (m *MarkersView) dialogUpdate() {
	if m.dialog.IsCanceled() {
		m.cancelDialog()
	}
	if m.dialog.IsConfirmed() {
		m.confirmDialog()
	}
}

func (m *MarkersView) confirmEdit() {
	m.chipsFilter.updateAll(m.markerDialog.tags)
	m.markerDialog.executeConfirm(m.audio)
}
func (m *MarkersView) confirmTagFilter() {
	m.chipsFilter.updateEnabled(m.tagsDialog.filterChips)
}
func (m *MarkersView) confirmDeleteAll() {
	m.deleteMarkers()
	m.chipsFilter.purge()
}

func (m *MarkersView) openMarkerDialog(curMarker *tm.TimeMarker, owner dialogOwner, title string) {
	if curMarker == nil {
		return
	}
	m.dialogOwner = owner
	if owner == create {
		m.markerDialog.focuser.RequestFocus(m.nameField)
	}
	m.markerDialog.prepareForOpening(curMarker, m.chipsFilter.all)

	m.dialog.Basic(m.th, title, func(gtx layout.Context) layout.Dimensions {
		return m.markerDialog.Layout(gtx, m.audio.Seconds)
	})
	m.dialog.Show()
}

func (m *MarkersView) openCommentDialog(curMarker *tm.TimeMarker) {
	m.dialogOwner = comment
	m.commentDialog.prepareForOpening(curMarker)
	m.dialog.Basic(m.th, curMarker.Name, func(gtx layout.Context) layout.Dimensions {
		return m.commentDialog.Layout(gtx)
	})
	m.dialog.Show()
}

func (m *MarkersView) confirmComment() {
	m.commentDialog.executeConfirm()
}

const maxFilterW unit.Dp = 350

func (m *MarkersView) openTagsFilterDialog() {
	m.dialogOwner = tagFilter
	m.chipsFilter.recreate(*m.timeMarkers)
	filterChips := m.tagsDialog.createFreshChips(m.chipsFilter.all, m.chipsFilter.enabledMap)
	m.dialog.Basic(m.th, "Tags Filter", func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X = gtx.Dp(maxFilterW)
		if cursor, ok := m.tagsDialog.getCursorAndHandleEvents(gtx); ok {
			common.SetCursor(gtx, cursor)
		}
		return common.DrawChipsFilter(gtx, m.th, filterChips)
	})
	m.dialog.Show()
}

func (m *MarkersView) clearTagFilter() {
	m.chipsFilter.updateEnabled(nil)
}

func (m *MarkersView) openDeleteAllDialog() {
	m.dialogOwner = deleteAll
	m.dialog.SetIcon(micons.Warning)
	m.dialog.Basic(m.th, "Удалить все маркеры?", func(gtx layout.Context) layout.Dimensions {
		txt := material.Body2(m.th.Theme, "Это действие удалит все существующие маркеры для этой звуковой дорожки!")
		txt.WrapPolicy = text.WrapWords
		txt.Alignment = text.Middle
		return txt.Layout(gtx)
	})
	m.dialog.Show()
}
