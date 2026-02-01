package editor

import (
	"gioui.org/io/pointer"
)

type hitKind int

const (
	hitNone hitKind = iota
	hitSoundWave
	hitMLifeArea
	hitMCreateArea
	hitMDeleteArea
	hitM
	hitMName
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
	isDraggingMarker := ed.mode == modeMDrag
	switch p.Target.Kind {
	case hitNone:
		if isDraggingMarker {
			return
		}
		ed.setCursor(pointer.CursorDefault)
		ed.mode = modeIdle
	case hitSoundWave:
		if isDraggingMarker {
			return
		}
		ed.setCursor(pointer.CursorCrosshair)
		ed.mode = modeHitWave
	case hitMLifeArea:
		if isDraggingMarker {
			return
		}
		ed.setCursor(pointer.CursorDefault)
		ed.mode = modeMLife
	case hitMCreateArea:
		ed.setCursor(pointer.CursorPointer)
		ed.mode = modeMCreateIntent
	case hitMDeleteArea:
		ed.setCursor(pointer.CursorPointer)
		ed.mode = modeMDeleteIntent
	case hitM:
		if isDraggingMarker {
			return
		}
		ed.setCursor(pointer.CursorGrab)
		ed.mode = modeMHit
	case hitMName:
		if isDraggingMarker {
			return
		}
		ed.setCursor(pointer.CursorText)
		ed.mode = modeMEditIntent
	}
}

func (ed *Editor) handleIdle(p pointerEvent) {
	ed.transition(p)
}

func (ed *Editor) handleHitWave(p pointerEvent) {
	switch p.Event.Kind {
	case pointer.Scroll:
		ed.handleWaveScroll(p.Event.Scroll, p.Event.Position)
	case pointer.Press:
		ed.handleWaveClick(p.Event.Position, p.Event.Buttons)
	}
	ed.transition(p)
}

func (ed *Editor) handleMLife(p pointerEvent) {
	ed.transition(p)
}

func (ed *Editor) handleMCreateIntent(p pointerEvent) {
	switch p.Event.Kind {
	case pointer.Press:
		ed.markers.newMarker(ed.playhead.bytes)
	}
	ed.transition(p)
}

func (ed *Editor) handleMHit(p pointerEvent) {
	switch p.Event.Kind {
	case pointer.Drag:
		ed.mode = modeMDrag
		ed.setCursor(pointer.CursorGrabbing)
	}
	ed.transition(p)
}

func (ed *Editor) handleDragMarker(p pointerEvent) {
	switch p.Event.Kind {
	case pointer.Drag:
		dSamples := int(ed.scroll.samplesPerPx * p.Event.Position.X)
		m := p.Target.Marker
		m.Pcm = ed.audio.getPcmFromSamples(ed.scroll.leftB + int(dSamples))
		m.Pcm = clamp(0, m.Pcm, ed.audio.pcmLen)
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
		ed.handleHitWave(p)
	case modeMLife:
		ed.handleMLife(p)
	case modeMCreateIntent:
		ed.handleMCreateIntent(p)
	case modeMDeleteIntent:
	//
	case modeMHit:
		ed.handleMHit(p)
	case modeMEditIntent:
	//
	case modeMEdit:
	//
	case modeMDrag:
		ed.handleDragMarker(p)
	}
}
