package markersview

import (
	"log"
	"strconv"

	"gioui.org/io/key"
)

func (m *MarkersView) handleKeyEvents(e key.Event) {
	if e.State == key.Release {
		return
	}

	switch e.Name {
	case key.NameSpace:
		if len(m.hotKeyBuf) == 0 {
			m.replayMarkers()
			return
		}
		idx, err := strconv.Atoi(string(m.hotKeyBuf))
		if err != nil {
			log.Fatal("Unreachable", err)
		}
		marker := m.timeMarkers.Get(idx-1, true)
		if marker != nil {
			m.toggleMarker(marker)
		}
	case key.NameDeleteBackward:
		m.hotKeyBuf = m.hotKeyBuf[:0]
	default:
		if len(m.hotKeyBuf) < selectionRuneLimit {
			m.hotKeyBuf = append(m.hotKeyBuf, []rune(e.Name)[0])
		}
	}
}
