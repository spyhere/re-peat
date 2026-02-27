package markersview

import (
	"image"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/spyhere/re-peat/internal/common"
)

var searchable = common.Searchable{}

func (m *MarkersView) Layout(gtx layout.Context) layout.Dimensions {
	var searchDims layout.Dimensions
	common.OffsetBy(gtx, image.Pt(0, 200), func() {
		common.CenteredX(gtx, func() layout.Dimensions {
			searchDims = common.DrawSearch(gtx, m.th, common.SProps{
				DefaultText: "Название маркера...",
				Searchable:  &searchable,
			})
			return searchDims
		})
	})
	common.OffsetBy(gtx, image.Pt(0, 200+searchDims.Size.Y+20), func() {
		common.DrawDivider(gtx, m.th, common.DividerProps{
			Inset: common.DividerMiddleInset,
		})
	})

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
