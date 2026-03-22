package common

import (
	"image"
	"image/color"
	"log"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

type Focuser interface {
	RequestFocus(f Focusable)
	SetFocus(f Focusable)
	RequestBlur()
}

type Inputable struct {
	isHovered           bool
	isFocused           bool // derived from gtx.Focused during layout
	isDirty             bool
	value               string
	hasSelection        bool
	hasSubmitted        bool
	hasEmptyDeleteEvent bool // delete button has been triggered when editor is empty. Useful for combobox
	shouldResetCaret    bool
	onBlurF             func()
	sanitizer           func(string) string
	widget.Editor
	Cancel  widget.Clickable
	Focuser // To drop focus when clicking on any non-interactable widget
}

func (in *Inputable) Update(gtx layout.Context) {
	if in.shouldResetCaret {
		in.Editor.SetCaret(0, 0)
		in.shouldResetCaret = false
	}
	HandlePointerEvents(gtx, &in.Editor, pointer.Press|pointer.Move|pointer.Leave, func(e pointer.Event) {
		switch e.Kind {
		case pointer.Move:
			in.isHovered = true
		case pointer.Leave:
			in.isHovered = false
		case pointer.Press:
			in.requestFocus(gtx)
		}
	})
	in.handleKeys(gtx)
	in.processEditorEvents(gtx)

	if in.Cancel.Clicked(gtx) {
		in.Editor.SetText("")
		in.value = ""
		in.requestBlur(gtx)
	}
}

func (in *Inputable) Subscribe(gtx layout.Context) {
	defer pointer.PassOp{}.Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, &in.Editor)
}

func (in *Inputable) SetSanitizer(f func(string) string) {
	in.sanitizer = f
}

func (in *Inputable) requestBlur(gtx layout.Context) {
	if in.Focuser != nil {
		in.Focuser.RequestBlur()
	} else {
		in.Blur(gtx)
	}
}

func (in *Inputable) OnBlur(f func()) {
	in.onBlurF = f
}
func (in *Inputable) Blur(gtx layout.Context) {
	if gtx.Focused(&in.Editor) {
		gtx.Execute(key.FocusCmd{Tag: nil})
	}
	in.isFocused = false
	in.shouldResetCaret = true
	if in.onBlurF != nil {
		in.onBlurF()
	}
}

func (in *Inputable) requestFocus(gtx layout.Context) {
	if in.Focuser != nil {
		in.Focuser.RequestFocus(in)
	} else {
		in.Focus(gtx)
	}
}

func (in *Inputable) selectAllOnFocus() {
	if in.isFocused {
		return
	}
	line, col := in.Editor.CaretPos()
	if line == 0 && col == 0 {
		txtLen := strlen(in.Editor.Text())
		in.Editor.SetCaret(txtLen, 0)
	}
}

func (in *Inputable) Focus(gtx layout.Context) {
	gtx.Execute(key.FocusCmd{Tag: &in.Editor})
	in.selectAllOnFocus()
	in.isFocused = true
}

// This is designed for frequent reads, but can be inaccurate.
// If you need accuracy use Editor's Text method
func (in *Inputable) GetInput() string {
	return in.value
}

func (in *Inputable) IsDirty() bool {
	if in.isDirty {
		in.isDirty = false
		return true
	}
	return in.isDirty
}

func (in *Inputable) HasSelection() bool {
	if in.hasSelection {
		in.hasSelection = false
		return true
	}
	return false
}

func (in *Inputable) HasSubmit() bool {
	if in.hasSubmitted {
		in.hasSubmitted = false
		return true
	}
	return false
}

func (in *Inputable) HasEmptyDeleteEvent() bool {
	if in.hasEmptyDeleteEvent {
		in.hasEmptyDeleteEvent = false
		return true
	}
	return false
}

func (in *Inputable) ClearEmptyDeleteEvent() {
	in.hasEmptyDeleteEvent = false
}

func (in *Inputable) IsHovered() bool {
	return in.isHovered
}

func (in *Inputable) IsFocused(gtx layout.Context) bool {
	in.isFocused = gtx.Focused(&in.Editor)
	if in.isFocused {
		in.Focuser.SetFocus(in)
	}
	return in.isFocused
}

func (in *Inputable) GetCursorType() (cursor pointer.Cursor, ok bool) {
	if in.IsHovered() {
		return pointer.CursorText, true
	}
	if in.Cancel.Hovered() {
		return pointer.CursorPointer, true
	}
	return pointer.CursorDefault, false
}

func (in *Inputable) handleKeys(gtx layout.Context) {
	backspaceFilter := key.Filter{Name: key.NameDeleteBackward, Focus: &in.Editor}
	if in.value != "" {
		// Disable this filter when editor still has something to not block Delete button for it
		backspaceFilter.Focus = &in
	}
	escapeFilter := key.Filter{Name: key.NameEscape, Focus: &in.Editor}
	if !in.isFocused {
		// Disable escape listener when this inputable is not in focus
		escapeFilter.Focus = &in
	}
	HandleKeyEvents(gtx, func(e key.Event) {
		if e.State == key.Release {
			return
		}
		switch e.Name {
		case key.NameEscape:
			in.requestBlur(gtx)
		case key.NameDeleteBackward:
			in.hasEmptyDeleteEvent = true
		}
	},
		escapeFilter,
		backspaceFilter,
	)
}

func (in *Inputable) processEditorEvents(gtx layout.Context) {
	for {
		we, ok := in.Editor.Update(gtx)
		if !ok {
			break
		}
		switch we.(type) {
		case widget.ChangeEvent:
			in.isDirty = true
			in.value = in.Editor.Text()
			if in.sanitizer != nil {
				clean := in.sanitizer(in.value)
				if clean != in.value {
					in.value = clean
					in.Editor.SetText(clean)
					in.Editor.SetCaret(len(clean), len(clean))
				}
			}
		case widget.SelectEvent:
			in.hasSelection = true
		case widget.SubmitEvent:
			in.hasSubmitted = true
		}
	}
}

type ComboboxOption struct {
	Text string
	Cl   widget.Clickable
}
type Comboboxable struct {
	Inputable
	optionsLs widget.List
	options   []ComboboxOption
	selected  string
}

func (c *Comboboxable) HandleOptionsEvents(gtx layout.Context) {
	for idx := range c.options {
		if c.options[idx].Cl.Clicked(gtx) {
			c.setSelectedValue(c.options[idx].Text)
			// When click event is being read "selectedValue" was already checked
			// and since Gio's idiomatic way is to check for events at the start
			// of the frame, there is no other way except request a new frame
			gtx.Execute(op.InvalidateCmd{})
		}
		if c.options[idx].Cl.Hovered() {
			SetCursor(gtx, pointer.CursorPointer)
		}
	}
}

func (c *Comboboxable) SetOptions(options []string) {
	for idx, it := range options {
		if idx < len(c.options) {
			c.options[idx].Text = it
		} else {
			c.options = append(c.options, ComboboxOption{Text: it})
		}
	}
	c.options = c.options[:len(options)]
}

func (c *Comboboxable) ResetOptionScroll() {
	c.optionsLs.ScrollTo(0)
}

func (c *Comboboxable) setSelectedValue(v string) {
	c.selected = v
}

func (c *Comboboxable) WithFocusManager(f Focuser) *Comboboxable {
	c.Inputable.Focuser = f
	return c
}

func (c *Comboboxable) HasSelectedValue() (string, bool) {
	v := c.selected
	if v != "" {
		c.selected = ""
	}
	return v, v != ""
}

func (c *Comboboxable) Blur(gtx layout.Context) {
	c.optionsLs.ScrollTo(0)
	c.Inputable.Blur(gtx)
}

type TableProps[T any] struct {
	Axis                 layout.Axis
	ColumsNum            int
	HeaderCellsAlignment []layout.Direction
	RowCellsAlignment    []layout.Direction
	RowValueCb           func(int) T
	RowFilterCb          func(T) bool
}

func NewTable[T any](props TableProps[T]) *Table[T] {
	l := widget.List{}
	l.Axis = props.Axis
	return &Table[T]{
		columns:          props.ColumsNum,
		list:             &l,
		columnWidths:     make([]int, props.ColumsNum),
		hCellsAllignment: props.HeaderCellsAlignment,
		headCellFuncs:    make([]HeadCellComp, props.ColumsNum),
		rCellsAllignment: props.RowCellsAlignment,
		rowCellFuncs:     make([]CellComp[T], props.ColumsNum),
		cellsBuf:         make([]layout.FlexChild, props.ColumsNum),
		rowValueCb:       props.RowValueCb,
		rowFilterCb:      props.RowFilterCb,
	}
}

type HeadCellComp func(gtx layout.Context) layout.Dimensions
type CellComp[T any] func(gtx layout.Context, rowIdx int, rowValue T) layout.Dimensions
type Table[T any] struct {
	columns          int
	Rows             int
	rowValueCb       func(int) T
	rowFilterCb      func(T) bool
	cellsBuf         []layout.FlexChild
	list             *widget.List
	columnWidths     []int
	hCellsAllignment []layout.Direction
	rCellsAllignment []layout.Direction
	headCellFuncs    []HeadCellComp
	rowCellFuncs     []CellComp[T]
	BottomMargin     bool
}

func (t *Table[T]) HeadCells(hFuncs ...HeadCellComp) {
	if len(hFuncs) != t.columns {
		log.Fatal("Incorrect usage of table! Header: number of cell render functions are not equal to set columns amount")
	}
	t.headCellFuncs = hFuncs
}

func (t *Table[T]) RowCells(rFuncs ...CellComp[T]) {
	if len(rFuncs) != t.columns {
		log.Fatal("Incorrect usage of table! Row: number of cell render functions are not equal to set columns amount")
	}
	t.rowCellFuncs = rFuncs
}

const tableXMargin = 8
const tableYMargin = 12

func (t *Table[T]) Layout(gtx layout.Context, th *theme.RepeatTheme, colWidths []int) {
	DrawBox(gtx, Box{
		Size:  image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Max.Y),
		Color: th.Palette.CardBg,
		R:     theme.CornerR(0, 0, 20, 20),
	})

	xMargin, yMargin := gtx.Dp(tableXMargin), gtx.Dp(tableYMargin)
	gtx.Constraints.Min = image.Point{}
	gtx.Constraints.Max.X -= xMargin * 2
	gtx.Constraints.Max.Y -= yMargin
	if t.BottomMargin {
		gtx.Constraints.Max.Y -= yMargin
	}
	var cellWidthSum int
	var colSum int
	maxX := gtx.Constraints.Max.X
	for idx, it := range colWidths {
		colSum += it
		if colSum > 100 {
			log.Fatal("Learn to count! Column width percentage sum is more than 100")
		}
		v := it * maxX / 100
		cellWidthSum += v
		t.columnWidths[idx] = v
		if idx < len(colWidths)-1 {
			OffsetBy(gtx, image.Pt(cellWidthSum+xMargin, yMargin), func(gtx layout.Context) {
				DrawDivider(gtx, th, DividerProps{Axis: Vertical})
			})
		}

	}
	if cellWidthSum < gtx.Constraints.Max.X {
		t.columnWidths[len(t.columnWidths)-1] += gtx.Constraints.Max.X - cellWidthSum
	}
	OffsetBy(gtx, image.Pt(xMargin, yMargin), func(gtx layout.Context) {
		gtx.Constraints.Max.Y -= yMargin * 2
		t.layout(gtx, th, yMargin)
	})
}

const headerH unit.Dp = 42
const cellMargin unit.Dp = 5
const rowHeight unit.Dp = 50

func (t *Table[T]) layout(gtx layout.Context, th *theme.RepeatTheme, bottomMargin int) {
	headerH := gtx.Dp(headerH)
	for colIdx, it := range t.headCellFuncs {
		t.cellsBuf[colIdx] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			columnDims := layout.Dimensions{Size: image.Pt(t.columnWidths[colIdx], headerH)}
			gtx.Constraints.Max = columnDims.Size
			gtx.Constraints.Min = columnDims.Size
			cellAl := t.hCellsAllignment[colIdx]
			layout.UniformInset(cellMargin).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return cellAl.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return it(gtx)
				})
			})
			return columnDims
		})
	}
	layout.Flex{}.Layout(gtx, t.cellsBuf...)

	rowH := gtx.Dp(rowHeight)
	OffsetBy(gtx, image.Pt(0, headerH), func(gtx layout.Context) {
		DrawDivider(gtx, th, DividerProps{})
		gtx.Constraints.Max.Y -= bottomMargin
		material.List(th.Theme, t.list).Layout(gtx, t.Rows, func(gtx layout.Context, rowIdx int) layout.Dimensions {
			rowValue := t.rowValueCb(rowIdx)
			if t.rowFilterCb != nil && !t.rowFilterCb(rowValue) {
				return layout.Dimensions{}
			}
			for colIdx, it := range t.rowCellFuncs {
				t.cellsBuf[colIdx] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					columnDims := layout.Dimensions{Size: image.Pt(t.columnWidths[colIdx], rowH)}
					cellAl := t.rCellsAllignment[colIdx]
					gtx.Constraints.Max = columnDims.Size
					gtx.Constraints.Min = columnDims.Size

					layout.UniformInset(cellMargin).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return cellAl.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return it(gtx, rowIdx, rowValue)
						})
					})
					return columnDims
				})
			}
			return layout.Flex{}.Layout(gtx, t.cellsBuf...)
		})
	})
}

type dialogButton struct {
	Text     string
	Icon     widget.Icon
	Enabled  bool
	IsHidden bool
}

type dialogType int

const (
	dialogBasic dialogType = iota
	dialogError
)

type Dialog struct {
	th              *theme.RepeatTheme
	isOpen          bool
	isCanceled      bool
	isConfirmed     bool
	variant         dialogType
	Ok              widget.Clickable
	OkProps         dialogButton
	Cancel          widget.Clickable
	Scrim           widget.Clickable
	CancelProps     dialogButton
	title           string
	icon            *widget.Icon
	iconC           color.NRGBA
	hasIcon         bool
	content         func(layout.Context) layout.Dimensions
	contentLs       widget.List
	mustRedraw      bool // dialog's show or hide methods could be called via key, that is not requesting invalidation
	inputBlockArmed bool // disable gtx when modal is open giving 1 extra frame
}

func (d *Dialog) SetIcon(icon *widget.Icon) {
	d.icon = icon
}

func (d *Dialog) SetIconColor(c color.NRGBA) {
	d.iconC = c
}

func (d *Dialog) Basic(th *theme.RepeatTheme, title string, w func(gtx layout.Context) layout.Dimensions) {
	d.variant = dialogBasic
	d.th = th
	d.title = title
	d.content = w
}

func (d *Dialog) Error(th *theme.RepeatTheme, title string, w func(gtx layout.Context) layout.Dimensions) {
	d.variant = dialogError
	d.th = th
	d.title = title
	d.content = w
}

func (d *Dialog) Show() {
	d.isOpen = true
	d.mustRedraw = true
}
func (d *Dialog) Hide() {
	d.isOpen = false
	d.mustRedraw = true
	d.icon = nil
}
func (d *Dialog) IsOpen() bool {
	return d.isOpen
}

func (d *Dialog) IsCanceled() bool {
	if d.isCanceled {
		d.isCanceled = false
		return true
	}
	return false
}
func (d *Dialog) IsConfirmed() bool {
	if d.isConfirmed {
		d.isConfirmed = false
		return true
	}
	return false
}

func (d *Dialog) ShouldDisableGtx(gtx layout.Context) bool {
	if d.isOpen {
		if !d.inputBlockArmed {
			d.inputBlockArmed = true
			// We are giving 1 more frame to finish possible click events
			gtx.Execute(op.InvalidateCmd{})
		} else {
			return true
		}
	} else {
		d.inputBlockArmed = false
	}
	return false
}

func (d *Dialog) protectDialogFromScrim(gtx layout.Context) {
	event.Op(gtx.Ops, d)
}

// This should be at the beginning of the frame
func (d *Dialog) Update(gtx layout.Context) {
	if !d.isOpen {
		return
	}
	if d.Cancel.Clicked(gtx) || d.Scrim.Clicked(gtx) {
		d.isCanceled = true
	}
	if d.Ok.Clicked(gtx) {
		d.isConfirmed = true
	}
}

type dialogMaterialSpecs struct {
	shape              unit.Dp
	padd               unit.Dp
	iconSz             unit.Dp
	iconTitlePadd      unit.Dp
	titleBodyPadd      unit.Dp
	bodyActionsPadd    unit.Dp
	betweenButtonsPadd unit.Dp
	minW               unit.Dp
	maxW               unit.Dp
	maxContentHPercent int
}

var dialogSpecs = dialogMaterialSpecs{
	shape:              28,
	padd:               24,
	iconSz:             24,
	iconTitlePadd:      16,
	titleBodyPadd:      16,
	bodyActionsPadd:    24,
	betweenButtonsPadd: 8,
	minW:               280,
	maxW:               560,
	maxContentHPercent: 70,
}

func (d *Dialog) Layout(gtx layout.Context) layout.Dimensions {
	if d.mustRedraw {
		// pointer requests redraw, but key - not. So using Tab and Enter, could
		// trigger modal to show/hide, but frame won't be requested by such action
		gtx.Execute(op.InvalidateCmd{})
		d.mustRedraw = false
	}
	if !d.isOpen {
		return layout.Dimensions{}
	}
	if d.th == nil {
		panic("Dialog: no theme available")
	}
	bg := image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Max.Y)
	DrawBox(gtx, Box{
		Size:      bg,
		Color:     d.th.Palette.Backdrop,
		Clickable: &d.Scrim,
		HideInk:   true,
	})

	padd := gtx.Dp(dialogSpecs.padd)
	fullPadd := padd * 2
	minW, maxW := gtx.Dp(dialogSpecs.minW), gtx.Dp(dialogSpecs.maxW)
	maxH := dialogSpecs.maxContentHPercent * gtx.Constraints.Max.Y / 100
	shape := gtx.Dp(dialogSpecs.shape)
	betweenButtonsPad := gtx.Dp(dialogSpecs.betweenButtonsPadd)

	contentM, contentDims := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X = maxW - fullPadd
		gtx.Constraints.Min.X = minW
		gtx.Constraints.Min.Y = 0
		dims := d.content(gtx)
		return dims
	})

	innerM, innerDims := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
		currentWidth := Clamp(minW, contentDims.Size.X, maxW)
		gtx.Constraints.Max.X = currentWidth - fullPadd
		gtx.Constraints.Max.Y = maxH - fullPadd
		var incrDims layout.Dimensions
		{
			// Icon for Title
			if d.icon != nil {
				iconsSize := gtx.Dp(dialogSpecs.iconSz)
				OffsetBy(gtx, image.Pt(currentWidth/2-iconsSize/2-padd, 0), func(gtx layout.Context) {
					gtx.Constraints.Min.X = iconsSize
					iconC := d.th.Palette.Dialog.IconC
					if d.iconC.A != 0 {
						iconC = d.iconC
					}
					d.icon.Layout(gtx, iconC)
				})
				incrDims.Size.Y += iconsSize + gtx.Dp(dialogSpecs.iconTitlePadd)
			}
			// Title
			OffsetBy(gtx, image.Pt(0, incrDims.Size.Y), func(gtx layout.Context) {
				gtx.Constraints.Min = image.Point{}
				title := material.H6(d.th.Theme, d.title)
				if d.icon != nil {
					title.Alignment = text.Middle
				}
				gtx.Constraints.Min.X = currentWidth - fullPadd
				titleDims := title.Layout(gtx)
				incrDims.Size.Y += titleDims.Size.Y
			})
			incrDims.Size.Y += gtx.Dp(dialogSpecs.titleBodyPadd)
		}

		gtx.Constraints.Min = gtx.Constraints.Max

		// Content
		incrDims.Size.X = contentDims.Size.X
		OffsetBy(gtx, image.Pt(0, incrDims.Size.Y), func(gtx layout.Context) {
			gtx.Constraints.Min = image.Point{}
			gtx.Constraints.Max.X = maxW - fullPadd
			d.contentLs.Axis = layout.Vertical
			material.List(d.th.Theme, &d.contentLs).Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
				contentM.Add(gtx.Ops)
				return contentDims
			})
			incrDims.Size.Y += min(gtx.Constraints.Max.Y, contentDims.Size.Y) + gtx.Dp(dialogSpecs.bodyActionsPadd)
		})

		// Actions
		OffsetBy(gtx, image.Pt(0, incrDims.Size.Y), func(gtx layout.Context) {
			var actionsDims layout.Dimensions
			gtx.Constraints.Min = image.Pt(currentWidth-fullPadd, 0)
			actionsDims = layout.E.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				var dims layout.Dimensions
				OffsetBy(gtx, image.Pt(padd, 0), func(gtx layout.Context) {
					var button material.ButtonStyle
					if !d.CancelProps.IsHidden {
						txt := "Cancel"
						if d.CancelProps.Text != "" {
							txt = d.CancelProps.Text
						}
						button = material.Button(d.th.Theme, &d.Cancel, txt)
						button.Background.A = 0x00
						button.Color = d.th.Palette.Dialog.ButtonEnabledC
						cancelDims := button.Layout(gtx)
						dims = cancelDims
						dims.Size.X += betweenButtonsPad
					}
					if !d.OkProps.IsHidden {
						OffsetBy(gtx, image.Pt(dims.Size.X, 0), func(gtx layout.Context) {
							txt := "OK"
							if d.OkProps.Text != "" {
								txt = d.OkProps.Text
							}
							button = material.Button(d.th.Theme, &d.Ok, txt)
							button.Background.A = 0x00
							button.Color = d.th.Palette.Dialog.ButtonEnabledC
							okDims := button.Layout(gtx)
							dims.Size.X += okDims.Size.X
							dims.Size.Y = okDims.Size.Y
						})
					}
				})
				dims.Size.X += padd
				return dims
			})
			incrDims.Size.Y += actionsDims.Size.Y + fullPadd
		})
		return incrDims
	})

	// Dialog background (Body)
	gtx.Constraints.Min = gtx.Constraints.Max
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X -= fullPadd
		gtx.Constraints.Max.Y -= fullPadd
		size := image.Rect(0, 0, Clamp(minW, innerDims.Size.X+fullPadd, maxW), innerDims.Size.Y)
		dialogDims := DrawBox(gtx, Box{
			Size:       size,
			Color:      d.th.Palette.CardBg,
			R:          theme.CornerR(shape, shape, shape, shape),
			GeometryCb: func() { d.protectDialogFromScrim(gtx) },
		})
		OffsetBy(gtx, image.Pt(padd, padd), func(gtx layout.Context) {
			innerM.Add(gtx.Ops)
		})
		return dialogDims
	})
}

func (d *Dialog) GetCursorType() (pointer.Cursor, bool) {
	if !d.isOpen {
		return pointer.CursorDefault, false
	}
	if d.Cancel.Hovered() || d.Ok.Hovered() {
		return pointer.CursorPointer, true
	}
	return pointer.CursorDefault, false
}

type Focusable interface {
	Focus(layout.Context)
	Blur(layout.Context)
}

// Manage focus for 1 or more Focusables. Focus and Blur are evaluated at the end of frame.
// Explanation: scrim is on top, passing pointer events (mimicking web page behavior),
// so we need to know whether any Focusables were hit underneath scrim, if it's the
// Focusable that is currently in focus - cancel blur.
type FocusManager struct {
	inFocus        Focusable
	setFocus       Focusable
	requestedFocus Focusable
	requestedBlur  Focusable
	scrimTag       struct{}
}

func (fm *FocusManager) blur(gtx layout.Context) {
	if fm.inFocus != nil {
		fm.inFocus.Blur(gtx)
	}
	fm.inFocus = nil
	fm.requestedBlur = nil
}
func (fm *FocusManager) focus(gtx layout.Context) {
	if fm.inFocus != fm.requestedFocus {
		fm.inFocus = fm.requestedFocus
		fm.inFocus.Focus(gtx)
	}
	fm.requestedFocus = nil
}

func (fm *FocusManager) update(gtx layout.Context) {
	HandlePointerEvents(gtx, &fm.scrimTag, pointer.Press, func(e pointer.Event) {
		if e.Kind == pointer.Press {
			fm.RequestBlur()
		}
	})
	if (fm.requestedFocus != nil) && (fm.requestedBlur != nil) && (fm.requestedFocus == fm.requestedBlur) {
		fm.requestedBlur, fm.requestedFocus = nil, nil
		return
	}
	if fm.setFocus != nil {
		fm.inFocus = fm.setFocus
		fm.setFocus = nil
	}
	if fm.requestedFocus != nil {
		fm.focus(gtx)
	}
	if fm.requestedBlur != nil {
		fm.blur(gtx)
	}
}

// Tell FocusManager what has focus at the moment
func (fm *FocusManager) SetFocus(it Focusable) {
	fm.setFocus = it
}

// Request focus programatically
func (fm *FocusManager) RequestFocus(it Focusable) {
	fm.requestedFocus = it
}

// Request Blur programatically
func (fm *FocusManager) RequestBlur() {
	fm.requestedBlur = fm.inFocus
}

// This should be placed as lower as possible at the frame creation
func (fm *FocusManager) PlaceScrim(gtx layout.Context) {
	fm.update(gtx)
	if fm.inFocus == nil {
		return
	}
	scrimM, _ := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
		OffsetBy(gtx, image.Pt(-1e3, -1e3), func(gtx layout.Context) {
			passOp := pointer.PassOp{}.Push(gtx.Ops)
			RegisterTag(gtx, &fm.scrimTag, image.Rect(0, 0, 1e6, 1e6))
			passOp.Pop()
		})
		return layout.Dimensions{}
	})
	op.Defer(gtx.Ops, scrimM)
}
