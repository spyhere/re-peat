package prompt

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

func NewPrompter(th *theme.RepeatTheme) Prompter {
	return Prompter{
		th:     th,
		dialog: common.Dialog{},
		ch:     make(chan bool),
	}
}

type Prompter struct {
	th     *theme.RepeatTheme
	dialog common.Dialog
	ch     chan bool
}

// This blocks goroutine
func (p *Prompter) Ask(title, question string) bool {
	p.dialog.Basic(p.th, title, func(gtx layout.Context) layout.Dimensions {
		return material.Body2(p.th.Theme, question).Layout(gtx)
	})
	p.dialog.Show()
	return <-p.ch
}

// Should be at the end of the frame, since it uses dialog
func (p *Prompter) Layout(gtx layout.Context) {
	p.dialog.Update(gtx)
	if p.dialog.IsCanceled() {
		p.dialog.Hide()
		p.ch <- false
	}
	if p.dialog.IsConfirmed() {
		p.dialog.Hide()
		p.ch <- true
	}

	p.dialog.Layout(gtx)

	if cursor, ok := p.dialog.GetCursorType(); ok {
		cursor.Add(gtx.Ops)
	}
}
