package markersview

import (
	"image"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/spyhere/re-peat/internal/common"
)

var searchable = common.Searchable{}

func (m *MarkersView) Layout(gtx layout.Context) layout.Dimensions {
	common.OffsetBy(gtx, image.Pt(0, 400), func() {
		common.CenteredX(gtx, func() layout.Dimensions {
			return common.DrawSearch(gtx, m.th, common.SProps{
				DefaultText: "Название маркера...",
				Searchable:  &searchable,
			})
		})
	})
	if searchable.IsHovered() {
		if searchable.IsFocused() {
			common.SetCursor(gtx, pointer.CursorText)
		} else {
			common.SetCursor(gtx, pointer.CursorPointer)
		}
	}
	return layout.Dimensions{}
}
