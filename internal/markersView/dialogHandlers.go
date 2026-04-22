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
	if m.Dialog.IsCanceled() {
		m.cancelDialog()
	}
	if m.Dialog.IsConfirmed() {
		m.confirmDialog()
	}
	m.Dialog.OkProps.Text = m.I18n.Generic.Ok
	m.Dialog.CancelProps.Text = m.I18n.Generic.Cancel
}

func (m *MarkersView) confirmCreate() {
	m.markerDialog.executeConfirm(m.AudioMeta)
	if !m.TimeMarkers.AttachNewMarker(m.draftMarker) {
		return
	}
	m.ChipsFilter.UpdateAll(m.draftMarker.CategoryTags)
	m.draftMarker = tm.TimeMarker{}
}
func (m *MarkersView) confirmEdit() {
	curMarker := m.markerDialog.TimeMarker
	m.markerDialog.executeConfirm(m.AudioMeta)
	m.ChipsFilter.UpdateAll(curMarker.CategoryTags)
}
func (m *MarkersView) confirmTagFilter() {
	m.ChipsFilter.UpdateEnabled(m.tagsDialog.filterChips)
}
func (m *MarkersView) confirmDeleteAll() {
	m.deleteMarkers()
	m.ChipsFilter.Purge()
}

func (m *MarkersView) openMarkerDialog(curMarker *tm.TimeMarker, owner dialogOwner, title string) {
	if curMarker == nil {
		return
	}
	m.Lg.Info("Markers: open marker dialog")
	m.dialogOwner = owner
	if owner == create {
		m.markerDialog.focuser.RequestFocus(m.markerDialog.nameField)
	}
	m.markerDialog.prepareForOpening(m.I18n, m.AudioMeta, curMarker, m.ChipsFilter.All)

	m.Dialog.Basic(m.Th, title, func(gtx layout.Context) layout.Dimensions {
		return m.markerDialog.Layout(gtx, m.AudioMeta.Seconds)
	})
	m.Dialog.Show()
}

func (m *MarkersView) openCommentDialog(curMarker *tm.TimeMarker) {
	m.Lg.Info("Markers: open comment dialog")
	m.dialogOwner = comment
	m.commentDialog.prepareForOpening(curMarker, m.I18n)
	m.Dialog.Basic(m.Th, curMarker.Name, func(gtx layout.Context) layout.Dimensions {
		return m.commentDialog.Layout(gtx)
	})
	m.Dialog.Show()
}

func (m *MarkersView) confirmComment() {
	m.commentDialog.executeConfirm()
}

const maxFilterW unit.Dp = 350

func (m *MarkersView) openTagsFilterDialog() {
	m.Lg.Info("Markers: open tags filter dialog")
	m.dialogOwner = tagFilter
	m.ChipsFilter.Recreate(m.TimeMarkers)
	filterChips := m.tagsDialog.createFreshChips(m.ChipsFilter.All, m.ChipsFilter.EnabledMap)
	m.Dialog.Basic(m.Th, m.I18n.Markers.TagsFilter, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X = gtx.Dp(maxFilterW)
		if cursor, ok := m.tagsDialog.getCursorAndHandleEvents(gtx); ok {
			common.SetCursor(gtx, cursor)
		}
		return common.DrawChipsFilter(gtx, m.Th, filterChips)
	})
	m.Dialog.Show()
}

func (m *MarkersView) clearTagFilter() {
	m.ChipsFilter.UpdateEnabled(nil)
}

func (m *MarkersView) openDeleteAllDialog() {
	m.Lg.Info("Markers: open delete all dialog")
	m.dialogOwner = deleteAll
	m.Dialog.SetIcon(micons.Warning)
	m.Dialog.Basic(m.Th, m.I18n.Markers.MDeleteALlTitle, func(gtx layout.Context) layout.Dimensions {
		txt := material.Body2(m.Th.Theme, m.I18n.Markers.MDeleteALlBody)
		txt.WrapPolicy = text.WrapWords
		txt.Alignment = text.Middle
		return txt.Layout(gtx)
	})
	m.Dialog.Show()
}
