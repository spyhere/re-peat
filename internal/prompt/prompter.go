package prompt

import (
	"image"
	"time"

	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/common"
	micons "github.com/spyhere/re-peat/internal/mIcons"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

func NewPrompter(th *theme.RepeatTheme) Prompter {
	return Prompter{
		th:      th,
		Dialog:  common.Dialog{},
		ch:      make(chan bool),
		working: make(chan struct{}, 1),
	}
}

type Prompter struct {
	th      *theme.RepeatTheme
	Dialog  common.Dialog
	ch      chan bool
	working chan struct{}
}

// This blocks goroutine
func (p *Prompter) Ask(title, question string) bool {
	p.working <- struct{}{}
	p.Dialog.Basic(p.th, title, func(gtx layout.Context) layout.Dimensions {
		return material.Body2(p.th.Theme, question).Layout(gtx)
	})
	p.Dialog.SetIcon(micons.Warning)
	p.Dialog.Show()
	return <-p.ch
}

// This blocks goroutine
func (p *Prompter) Tell(title, msg, ok string) bool {
	p.working <- struct{}{}
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

type UpdateInfo struct {
	HtmlUrl     string
	TagName     string
	Name        string
	PublishedAt time.Time
	Body        string
	Size        int64
}

var hyperl = common.Hyperlinkable{}

func (p *Prompter) AskUpdate(upd UpdateInfo) bool {
	p.working <- struct{}{}
	tagName := upd.TagName
	pubAt := upd.PublishedAt.Format("(02/01/2006)")
	p.Dialog.Basic(p.th, "New version released - "+tagName+" "+pubAt, func(gtx layout.Context) layout.Dimensions {
		if hyperl.IsPressed() {
			common.OpenBrowserLink(upd.HtmlUrl)
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				title := material.H4(p.th.Theme, upd.Name)
				title.Alignment = text.Middle
				gtx.Constraints.Min = image.Pt(gtx.Constraints.Max.X, 0)
				return title.Layout(gtx)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = image.Pt(gtx.Constraints.Max.X, 0)
				flex := layout.Flex{Spacing: layout.SpaceSides}.Layout(gtx,
					layout.Rigid(material.Body2(p.th.Theme, "Read in browser").Layout),
					layout.Rigid(layout.Spacer{Width: 10}.Layout),
					layout.Rigid(common.Hyperlink(p.th, &hyperl, upd.HtmlUrl).Layout),
				)
				if cursor, ok := hyperl.GetCursorType(); ok {
					common.SetCursor(gtx, cursor)
				}
				return flex
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				body := material.Body2(p.th.Theme, "\n\n"+upd.Body)
				return body.Layout(gtx)
			}),
		)
	})
	p.Dialog.OkProps.Text = "Download from browser (" + common.ParseSize(upd.Size) + ")"
	p.Dialog.CancelProps.Text = "Remind me later"
	p.Dialog.Show()
	return <-p.ch
}

// Should be at the end of the frame, since it uses dialog
func (p *Prompter) Layout(gtx layout.Context) {
	p.Dialog.Update(gtx)
	if p.Dialog.IsCanceled() {
		p.Dialog.Hide()
		p.ch <- false
		<-p.working
	}
	if p.Dialog.IsConfirmed() {
		p.Dialog.Hide()
		p.ch <- true
		<-p.working
	}

	p.Dialog.Layout(gtx)

	if cursor, ok := p.Dialog.GetCursorType(); ok {
		cursor.Add(gtx.Ops)
	}
}
