package projectview

import (
	"gioui.org/layout"
	"gioui.org/widget"
)

func (pv *ProjectView) dispatch(gtx layout.Context) {
	if pv.AudioLoadCl.Clicked(gtx) {
		pv.AudioLoadCl = widget.Clickable{}
		pv.AudioLoad()
	}
	if pv.MarkersLoadCl.Clicked(gtx) {
		pv.MarkersLoadCl = widget.Clickable{}
		pv.MarkersLoad()
	}
	if pv.MarkersSaveCl.Clicked(gtx) {
		pv.MarkersSaveCl = widget.Clickable{}
		pv.MarkersSave()
	}
	if pv.MarkersSaveAsCl.Clicked(gtx) {
		pv.MarkersSaveAsCl = widget.Clickable{}
		pv.MarkersSaveAs()
	}
}
