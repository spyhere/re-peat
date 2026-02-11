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
	hitBackdrop
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
	isEditingMarker := ed.mode == modeMEdit
	switch p.Target.Kind {
	case hitNone:
		if isDraggingMarker {
			return
		}
		ed.setCursor(pointer.CursorDefault)
		ed.mode = modeIdle
		ed.markers.stopHover()
	case hitSoundWave:
		if isDraggingMarker {
			return
		}
		ed.setCursor(pointer.CursorCrosshair)
		ed.mode = modeHitWave
		ed.markers.stopHover()
	case hitMLifeArea:
		if isDraggingMarker {
			return
		}
		ed.setCursor(pointer.CursorDefault)
		ed.mode = modeMLife
		ed.markers.stopHover()
	case hitMCreateArea:
		ed.setCursor(pointer.CursorPointer)
		ed.mode = modeMCreateIntent
		ed.markers.stopHover()
	case hitMDeleteArea:
		ed.setCursor(pointer.CursorPointer)
		ed.mode = modeMDeleteIntent
		ed.markers.startHover(p.Target.Marker)
	case hitM:
		if isDraggingMarker {
			return
		}
		ed.setCursor(pointer.CursorGrab)
		ed.markers.startHover(p.Target.Marker)
		ed.mode = modeMHit
	case hitMName:
		if isDraggingMarker || isEditingMarker {
			return
		}
		ed.setCursor(pointer.CursorText)
		ed.mode = modeMEditIntent
	case hitBackdrop:
		ed.setCursor(pointer.CursorDefault)
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
		if p.Target.Marker == nil {
			ed.handleWaveClick(p.Event.Position, p.Event.Buttons)
		}
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
		ed.setCursor(pointer.CursorText)
		ed.mode = modeMEdit
		return
	}
	ed.transition(p)
}

func (ed *Editor) handleMDeleteIntent(p pointerEvent) {
	switch p.Event.Kind {
	case pointer.Press:
		p.Target.Marker.markDead()
	}
	ed.transition(p)
}

func (ed *Editor) handleMHit(p pointerEvent) {
	switch p.Event.Kind {
	case pointer.Release:
		ed.setPlayhead(p.Target.Marker.pcm)
	case pointer.Drag:
		ed.mode = modeMDrag
		ed.setCursor(pointer.CursorGrabbing)
	}
	ed.transition(p)
}

func (ed *Editor) handleMEditIntent(p pointerEvent) {
	switch p.Event.Kind {
	case pointer.Press:
		ed.mode = modeMEdit
		m := p.Target.Marker
		ed.mEditor.SetText(m.name)
		ed.mEditor.SetCaret(len(m.name), 0)
		ed.markers.startEdit(m)
	}
	ed.transition(p)
}

func (ed *Editor) handleMEdit(p pointerEvent) {
	switch p.Event.Kind {
	case pointer.Press:
		if p.Target.Kind == hitBackdrop {
			ed.mode = modeIdle
			ed.cancelEdit()
		}
	}
	ed.transition(p)
}

func (ed *Editor) handleDragMarker(p pointerEvent) {
	switch p.Event.Kind {
	case pointer.Drag:
		dSamples := int(ed.scroll.samplesPerPx * p.Event.Position.X)
		m := p.Target.Marker
		m.pcm = ed.audio.getPcmFromSamples(ed.scroll.leftB + int(dSamples))
		m.pcm = clamp(0, m.pcm, ed.audio.pcmLen)
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
		ed.handleMDeleteIntent(p)
	case modeMHit:
		ed.handleMHit(p)
	case modeMEditIntent:
		ed.handleMEditIntent(p)
	case modeMEdit:
		ed.handleMEdit(p)
	case modeMDrag:
		ed.handleDragMarker(p)
	}
}
