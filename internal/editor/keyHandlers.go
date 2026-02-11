package editor

import "gioui.org/io/key"

func (ed *Editor) switchPlayerState() {
	if ed.mode == modeMEdit {
		return
	}
	if !ed.p.IsPlaying() {
		if ed.playhead.bytes >= ed.audio.pcmLen {
			return
		}
		ed.p.Play()
		ed.p.WaitUntilReady()
	} else {
		ed.p.Pause()
		ed.playhead.reset()
		ed.p.Set(ed.playhead.bytes)
	}
}

func (ed *Editor) cancelEdit() {
	if !ed.markers.isEditing() {
		return
	}
	if ed.markers.editing.name == "" {
		ed.markers.editing.markDead()
	}
	ed.markers.stopEdit()
	ed.mEditor.SetText("")
	ed.mode = modeIdle
}

func (ed *Editor) nudgePlayhead(forward bool) {
	if ed.markers.isEditing() {
		return
	}
	dPcm := ed.audio.getPcmFromSamples(int(ed.scroll.samplesPerPx))
	if !forward {
		dPcm *= -1
	}
	ed.setPlayhead(ed.playhead.bytes + dPcm*4)
}

func (ed *Editor) collapseRenamerSelection() {
	if !ed.markers.isEditing() {
		return
	}
	start, end := ed.mEditor.Selection()
	if start != end {
		ed.mEditor.SetCaret(start, start)
	}
}

func (ed *Editor) handleKeyEvents(e key.Event) {
	if e.State == key.Press {
		switch e.Name {
		case key.NameSpace:
			ed.switchPlayerState()
		case key.NameEscape:
			ed.cancelEdit()
		case key.NameLeftArrow:
			ed.collapseRenamerSelection()
			ed.nudgePlayhead(false)
		case key.NameRightArrow:
			ed.collapseRenamerSelection()
			ed.nudgePlayhead(true)
		}
	}
}
