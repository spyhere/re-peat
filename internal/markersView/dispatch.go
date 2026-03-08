package markersview

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"github.com/spyhere/re-peat/internal/common"
)

func (m *MarkersView) dispatch(gtx layout.Context) {
	isModalOpen := m.dialogOwner != none
	if !m.searchbar.IsFocused() && !isModalOpen {
		m.dispatchKeyEvents(gtx)
	}
}

func (m *MarkersView) dispatchKeyEvents(gtx layout.Context) {
	common.HandleKeyEvents(gtx, m.handleKeyEvents,
		key.Filter{Name: key.NameEscape},
		key.Filter{Name: key.NameSpace},
		key.Filter{Name: key.NameDeleteBackward},
		key.Filter{Name: "1"},
		key.Filter{Name: "2"},
		key.Filter{Name: "3"},
		key.Filter{Name: "4"},
		key.Filter{Name: "5"},
		key.Filter{Name: "6"},
		key.Filter{Name: "7"},
		key.Filter{Name: "8"},
		key.Filter{Name: "9"},
		key.Filter{Name: "0"},
	)
}
