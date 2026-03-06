package markersview

import (
	"fmt"
	"image"

	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	micons "github.com/spyhere/re-peat/internal/mIcons"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
)

func (m *MarkersView) dialogUpdate(gtx layout.Context) {
	if m.dialog.Cancel.Clicked(gtx) || m.dialog.Scrim.Clicked(gtx) {
		m.dialog.Hide()
	}
	if m.dialog.Ok.Clicked(gtx) {
		switch m.dialogOwner {
		case edit:
			m.confirmEdit()
		case tagFilter:
			m.confirmTagFilter()
		case deleteAll:
			m.confirmDeleteAll()
		}
	}
}

func (m *MarkersView) confirmEdit() {
	fmt.Println("Edit confirmed")
	m.dialog.Hide()
}
func (m *MarkersView) confirmTagFilter() {
	fmt.Println("Tag filter conrimed")
	m.dialog.Hide()
}
func (m *MarkersView) confirmDeleteAll() {
	m.deleteMarkers()
	m.dialog.Hide()
}

func (m *MarkersView) openEditDialog(curMarker *tm.TimeMarker) {
	m.dialogOwner = edit
	m.dialog.Basic(m.th, "Tag Edit", func(gtx layout.Context) layout.Dimensions {
		return common.DrawBox(gtx, common.Box{
			Size:  image.Rect(0, 0, 200, 200),
			Color: m.th.Palette.Selection.Bg,
		})
	})
	m.dialog.Show()
}

func (m *MarkersView) openTagsFilterDialog() {
	m.dialogOwner = tagFilter
	m.dialog.Basic(m.th, "Tags Filter", func(gtx layout.Context) layout.Dimensions {
		return common.DrawBox(gtx, common.Box{
			Size:  image.Rect(0, 0, 100, 100),
			Color: m.th.Palette.Divider,
		})
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
