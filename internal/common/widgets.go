package common

import (
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"
)

type Searchable struct {
	isHovered bool
	isFocused bool
	widget.Editor
	widget.Clickable
	Cancel widget.Clickable
}

func (s *Searchable) Update(gtx layout.Context) {
	if s.Cancel.Clicked(gtx) {
		s.Editor.SetText("")
		s.Blur()
	}
	HandlePointerEvents(gtx, &s.Editor, pointer.Press|pointer.Move|pointer.Leave, func(e pointer.Event) {
		switch e.Kind {
		case pointer.Move:
			s.isHovered = true
		case pointer.Leave:
			s.isHovered = false
		case pointer.Press:
			// Even if user missed the input field and pressed container the caret will be set anyway
			if s.Focus() {
				if e.Position.X < 100 {
					s.Editor.SetCaret(0, 0)
				} else {
					txtLen := len(s.GetInput())
					s.Editor.SetCaret(txtLen, txtLen)
				}
			}
		}
	})

	HandleKeyEvents(gtx, func(e key.Event) {
		switch e.Name {
		case key.NameEscape:
			s.Blur()
		}
	},
		key.Filter{Name: key.NameEscape},
	)
}

func (s *Searchable) Subscribe(gtx layout.Context) {
	defer pointer.PassOp{}.Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, &s.Editor)
}

func (s *Searchable) Blur() {
	s.isFocused = false
}

func (s *Searchable) Focus() (wasFocusedBefore bool) {
	if s.isFocused {
		return true
	}
	txtLen := len(s.Editor.Text())
	if txtLen > 0 {
		s.Editor.SetCaret(txtLen, 0)
	}
	s.isFocused = true
	return false
}

func (s *Searchable) GetInput() string {
	return s.Editor.Text()
}

func (s *Searchable) IsHovered() bool {
	return s.isHovered
}

func (s *Searchable) IsFocused() bool {
	return s.isFocused
}
