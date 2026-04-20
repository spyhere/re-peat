package prompt

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	micons "github.com/spyhere/re-peat/internal/mIcons"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

func NewPrompter(th *theme.RepeatTheme) Prompter {
	return Prompter{
		th:     th,
		Dialog: common.Dialog{},
		ch:     make(chan bool),
	}
}

type Prompter struct {
	th     *theme.RepeatTheme
	Dialog common.Dialog
	ch     chan bool
}

// This blocks goroutine
func (p *Prompter) Ask(title, question string) bool {
	p.Dialog.Basic(p.th, title, func(gtx layout.Context) layout.Dimensions {
		return material.Body2(p.th.Theme, question).Layout(gtx)
	})
	p.Dialog.SetIcon(micons.Warning)
	p.Dialog.Show()
	return <-p.ch
}

// This blocks goroutine
func (p *Prompter) Tell(title, msg, ok string) bool {
	p.Dialog.Info(p.th, title, func(gtx layout.Context) layout.Dimensions {
		txtStyles := material.Body2(p.th.Theme, msg)
		txtStyles.TextSize = 16
		return txtStyles.Layout(gtx)
	})
	if ok != "" {
		p.Dialog.OkProps.Text = ok
	}
	p.Dialog.Show()
	return <-p.ch
}

// Should be at the end of the frame, since it uses dialog
func (p *Prompter) Layout(gtx layout.Context) {
	p.Dialog.Update(gtx)
	if p.Dialog.IsCanceled() {
		p.Dialog.Hide()
		p.ch <- false
	}
	if p.Dialog.IsConfirmed() {
		p.Dialog.Hide()
		p.ch <- true
	}

	p.Dialog.Layout(gtx)

	if cursor, ok := p.Dialog.GetCursorType(); ok {
		cursor.Add(gtx.Ops)
	}
}
