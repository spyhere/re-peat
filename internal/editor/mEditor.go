package editor

import (
	"gioui.org/text"
	"gioui.org/widget"
)

func newMEditor() *widget.Editor {
	e := &widget.Editor{}
	e.MaxLen = 30
	e.SingleLine = true
	e.Submit = true
	e.Alignment = text.Start
	return e
}
