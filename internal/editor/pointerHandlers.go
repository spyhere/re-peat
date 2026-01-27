package editor

import "gioui.org/io/pointer"

type hitKind int

const (
	hitNone hitKind = iota
	hitWave
	hitMarker
)

type hitTarget struct {
	Kind   hitKind
	Marker *marker
}

type pointerEvent struct {
	Event  pointer.Event
	Target hitTarget
}

func (ed *Editor) transition(p pointerEvent) {
	isDragMarker := ed.mode == modeDragMarker
	isSetMarker := ed.mode == modeSetMarker
	switch p.Target.Kind {
	case hitNone:
		if isDragMarker || isSetMarker {
			return
		}
		ed.setCursor(pointer.CursorDefault)
		ed.mode = modeIdle
	case hitWave:
		if isDragMarker || isSetMarker {
			return
		}
		ed.setCursor(pointer.CursorCrosshair)
		ed.mode = modeHitWave
	case hitMarker:
		if isDragMarker || isSetMarker {
			return
		}
		ed.mode = modeHitMarker
		ed.setCursor(pointer.CursorGrab)
	}
}

func (ed *Editor) handleIdle(p pointerEvent) {
	ed.transition(p)
}

func (ed *Editor) handleWave(p pointerEvent) {
	switch p.Event.Kind {
	case pointer.Scroll:
		ed.handleWaveScroll(p.Event.Scroll, p.Event.Position)
	case pointer.Press:
		ed.handleWaveClick(p.Event.Position, p.Event.Buttons)
	case pointer.Move:
		ed.handleWaveMove(p.Target.Marker, p.Event.Position)
	}
	ed.transition(p)
}

func (ed *Editor) handleHitMarker(p pointerEvent) {
	switch p.Event.Kind {
	case pointer.Drag:
		ed.mode = modeDragMarker
		ed.setCursor(pointer.CursorGrabbing)
	}
	ed.transition(p)
}

func (ed *Editor) handleDragMarker(p pointerEvent) {
	switch p.Event.Kind {
	case pointer.Drag:
		dSamples := ed.scroll.samplesPerPx * p.Event.Position.X
		m := p.Target.Marker
		m.Samples = int(dSamples)
		m.Samples = clamp(0, m.Samples, ed.scroll.rightB)
	case pointer.Release:
		ed.mode = modeHitWave
	}
	ed.transition(p)
}

func (ed *Editor) handlePointer(p pointerEvent) {
	switch ed.mode {
	case modeIdle:
		ed.handleIdle(p)
	case modeHitWave:
		ed.handleWave(p)
	case modeHitMarker:
		ed.handleHitMarker(p)
	case modeSetMarker:
		ed.handleWave(p)
	case modeDragMarker:
		ed.handleDragMarker(p)
	}
}
