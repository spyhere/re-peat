package markersview

import (
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/spyhere/re-peat/internal/common"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

func newCommentDialog(th *theme.RepeatTheme) commentDialog {
	fm := &common.FocusManager{}
	return commentDialog{
		commentField: &common.Inputable{Focuser: fm},
		focuser:      fm,
		th:           th,
	}
}

type commentDialog struct {
	*tm.TimeMarker
	commentField *common.Inputable
	focuser      *common.FocusManager
	th           *theme.RepeatTheme
}

func (c *commentDialog) prepareForOpening(curMarker *tm.TimeMarker) {
	c.TimeMarker = curMarker
	c.commentField.SetText(curMarker.Notes)
	c.focuser.RequestFocus(c.commentField)
}

func (c *commentDialog) executeConfirm() {
	c.TimeMarker.Notes = c.commentField.GetInput()
	c.TimeMarker = nil
	c.focuser.RequestBlur(nil)
}
func (c *commentDialog) cancelComment() {
	c.TimeMarker = nil
	c.focuser.RequestBlur(nil)
}

func (c *commentDialog) getCursorType() (pointer.Cursor, bool) {
	if c.commentField.IsHovered() {
		return pointer.CursorText, true
	}
	return pointer.CursorDefault, false
}

func (c *commentDialog) Layout(gtx layout.Context) layout.Dimensions {
	if cursor, ok := c.getCursorType(); ok {
		common.SetCursor(gtx, cursor)
	}
	// TODO: Keep it generic for dialogs
	s := drawMarkerDialogSpecs
	inset := layout.Inset{Top: s.fieldsYMargin, Bottom: s.fieldsYMargin, Left: s.fieldsXMargin, Right: s.fieldsXMargin}
	dims := inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return common.DrawTextField(gtx, c.th, common.TextFieldProps{
				Base: common.InputFieldBase{
					LabelText: "Notes",
				},
				Inputable:   c.commentField,
				MaxLen:      500,
				Placeholder: "Was more than absolutely perfect...",
			})
		})
	})
	c.focuser.PlaceScrim(gtx)
	return dims
}
