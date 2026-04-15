package editorview

import "gioui.org/io/key"

func (ed *Editor) switchPlayerState() {
	if ed.mode == modeMEdit {
		return
	}
	if !ed.Player.IsPlaying() {
		if ed.playhead.samples >= ed.AudioMeta.MaxMonoSamples() {
			return
		}
		ed.startPlay()
	} else {
		ed.pausePlay()
	}
}

const nudgeMultiplier = 4

func (ed *Editor) nudgePlayhead(forward bool) {
	if ed.markers.isEditing() {
		return
	}
	dSamples := ed.scroll.samplesPerPx
	if !forward {
		dSamples = -dSamples
	}
	ed.setPlayhead(ed.playhead.samples + int(dSamples*nudgeMultiplier))
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
