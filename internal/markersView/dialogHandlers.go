package markersview

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	micons "github.com/spyhere/re-peat/internal/mIcons"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
)

func (m *MarkersView) dialogUpdate(gtx layout.Context) {
	if m.dialog.Cancel.Clicked(gtx) || m.dialog.Scrim.Clicked(gtx) {
		m.dialog.Hide()
		// TODO: Interface needed
		switch m.dialogOwner {
		case edit:
			m.markerDialog.blur(gtx)
		}
	}
	if m.dialog.Body.Clicked(gtx) {
		m.markerDialog.blur(gtx)
	}
	if m.dialog.Ok.Clicked(gtx) {
		switch m.dialogOwner {
		case edit:
			m.confirmEdit(gtx)
		case tagFilter:
			m.confirmTagFilter()
		case deleteAll:
			m.confirmDeleteAll()
		}
	}

	if cursor, ok := m.markerDialog.getCursorType(); ok {
		common.SetCursor(gtx, cursor)
		gtx.Execute(op.InvalidateCmd{})
	}
	m.markerDialog.handleFieldsEvents(gtx)

	if cursor, ok := m.tagsDialog.getCursorAndHandleEvents(gtx); ok {
		common.SetCursor(gtx, cursor)
	}
}

func (m *MarkersView) confirmEdit(gtx layout.Context) {
	m.chipsFilter.updateAll(m.markerDialog.tags)
	m.markerDialog.executeConfirm(m.audio)
	m.blur(gtx)
	m.dialog.Hide()
}
func (m *MarkersView) confirmTagFilter() {
	m.chipsFilter.updateEnabled(m.tagsDialog.filterChips)
	m.dialog.Hide()
}
func (m *MarkersView) confirmDeleteAll() {
	m.deleteMarkers()
	m.chipsFilter.purge()
	m.dialog.Hide()
}

// Move this to markerDialog.Layout
func (m *MarkersView) openEditDialog(curMarker *tm.TimeMarker) {
	if curMarker == nil {
		return
	}
	m.dialogOwner = edit
	m.markerDialog.prepareForOpening(curMarker, m.audio, m.chipsFilter.all)

	m.dialog.Basic(m.th, "Marker Edit", func(gtx layout.Context) layout.Dimensions {
		return drawMarkerDialogFields(gtx, m.th, markerDialogFieldsProps{
			name:         m.markerDialog.nameField,
			time:         m.markerDialog.timeField,
			tags:         m.markerDialog.tagsField,
			chips:        m.markerDialog.tags,
			totalSeconds: m.audio.Seconds,
			tagOptions:   m.markerDialog.getTagOptions(),
		})
	})
	m.dialog.Show()
}

const maxFilterW unit.Dp = 350

func (m *MarkersView) openTagsFilterDialog() {
	m.dialogOwner = tagFilter
	m.chipsFilter.recreate(*m.timeMarkers)
	filterChips := m.tagsDialog.createFreshChips(m.chipsFilter.all, m.chipsFilter.enabled)
	m.dialog.Basic(m.th, "Tags Filter", func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X = gtx.Dp(maxFilterW)
		return common.DrawChipsFilter(gtx, m.th, filterChips)
	})
	m.dialog.Show()
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
