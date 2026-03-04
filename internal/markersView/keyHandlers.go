package markersview

import (
	"fmt"
	"log"
	"strconv"
	"strings"

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
		buf := string(m.hotKeyBuf)
		if len(buf) < 2 {
			return
		}
		idx, err := strconv.Atoi(buf)
		if err != nil {
			log.Fatal("Unreachable", err)
		}
		marker := m.timeMarkers.Get(idx-1, true)
		if marker == nil {
			log.Fatal("Unreachable", err)
		}
		m.togglePlayer(marker)
	case key.NameEscape:
		m.clearHotKeyBuf()
	case key.NameDeleteBackward:
		m.clearHotKeyBuf()
	default:
		if len(m.hotKeyBuf) < selectionRuneLimit {
			m.hotKeyBuf = append(m.hotKeyBuf, []rune(e.Name)[0])
			buf := string(m.hotKeyBuf)
			if buf == "00" {
				m.clearHotKeyBuf()
				return
			}
			num, err := strconv.Atoi(buf)
			if err != nil {
				log.Fatal("Unreachable", err)
			}
			if num > len(*m.timeMarkers) {
				m.clearHotKeyBuf()
				return
			}
			if len(*m.timeMarkers) < 10 {
				mLenStr := fmt.Sprintf("%02d", len(*m.timeMarkers))
				if len(buf) < len(mLenStr) && !strings.HasPrefix(mLenStr, buf) {
					m.clearHotKeyBuf()
				}
			}
		}
	}
}
