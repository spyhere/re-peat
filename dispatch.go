package main

import (
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/spyhere/re-peat/internal/common"
)

func (a *App) dispatch(gtx layout.Context) {
	a.dispatchButtonsEvents(gtx)
}

func (a *App) dispatchButtonsEvents(gtx layout.Context) {
	for _, it := range a.buttons.arr {
		common.HandlePointerEvents(
			gtx,
			&it.tag,
			pointer.Enter|pointer.Leave|pointer.Move|pointer.Press,
			func(e pointer.Event) {
				switch e.Kind {
				case pointer.Enter:
					a.buttons.setHover()
				case pointer.Move:
					a.buttons.setHover()
				case pointer.Leave:
					a.buttons.stopHover()
				case pointer.Press:
					a.selectedTab = it.tab
				}
			},
		)
	}
}
