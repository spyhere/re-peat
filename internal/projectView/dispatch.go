package projectview

import (
	"gioui.org/layout"
	"gioui.org/widget"
)

func (pv *ProjectView) dispatch(gtx layout.Context) {
	if pv.audioLoadCl.Clicked(gtx) {
		pv.audioLoadCl = widget.Clickable{}
		pv.AudioLoad()
	}

	if pv.markersLoadCl.Clicked(gtx) {
		pv.markersLoadCl = widget.Clickable{}
		pv.MarkersLoad()
	}

	if pv.markersSaveCl.Clicked(gtx) {
		pv.MarkersSave()
	}

	if pv.markersSaveAsCl.Clicked(gtx) {
		pv.markersSaveAsCl = widget.Clickable{}
		pv.MarkersSaveAs()
	}
}
