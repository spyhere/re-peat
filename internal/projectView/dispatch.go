package projectview

import (
	"gioui.org/layout"
	"gioui.org/widget"
)

func (pv *ProjectView) dispatch(gtx layout.Context) {
	if pv.audioLoadCl.Clicked(gtx) {
		pv.audioLoadCl = widget.Clickable{}
		if err := pv.AudioLoad(); err != nil {
			pv.lg.Error("Project: audio load", err)
		}
	}
	if pv.markersLoadCl.Clicked(gtx) {
		pv.markersLoadCl = widget.Clickable{}
		if err := pv.MarkersLoad(); err != nil {
			pv.lg.Error("Project: markers load", err)
		}
	}
	if pv.markersSaveCl.Clicked(gtx) {
		if err := pv.MarkersSave(); err != nil {
			pv.lg.Error("Project: markers save", err)
		}
	}
	if pv.markersSaveAsCl.Clicked(gtx) {
		pv.markersSaveAsCl = widget.Clickable{}
		if err := pv.MarkersSaveAs(); err != nil {
			pv.lg.Error("Project: markers save as", err)
		}
	}
}
