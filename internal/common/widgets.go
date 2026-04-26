package common

import (
	"image"
	"image/color"
	"log"

	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/i18n"
	micons "github.com/spyhere/re-peat/internal/mIcons"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

type Focuser interface {
	RequestFocus(f Focusable)
	RequestBlur(f Focusable)
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
	in.handleSelectionCollapse(gtx)

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
		in.Focuser.RequestBlur(in)
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
		gtx.Execute(op.InvalidateCmd{})
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

func (in *Inputable) IsHovered() bool {
	return in.isHovered
}

func (in *Inputable) IsFocused(gtx layout.Context) bool {
	wasFocused := in.isFocused
	in.isFocused = gtx.Focused(&in.Editor)
	if in.isFocused && !wasFocused {
		// Focus operation is selecting the whole text (if clicked outside of editor)
		// but since focus can happen with keyboard, then selection mechanism will
		// only happen on the next frame, thus we need to call the next frame.
		gtx.Execute(op.InvalidateCmd{})
		in.requestFocus(gtx)
	} else if !in.isFocused && wasFocused {
		in.requestBlur(gtx)
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

func (in *Inputable) handleSelectionCollapse(gtx layout.Context) {
	rightArrow := key.Filter{Name: key.NameRightArrow, Focus: &in.Editor}
	if !in.isFocused {
		rightArrow.Focus = &in
	}
	HandleKeyEvents(gtx, func(e key.Event) {
		if e.State == key.Release {
			return
		}
		if e.Name == key.NameRightArrow {
			start, end := in.Editor.Selection()
			if start != end {
				in.Editor.SetCaret(start, start)
			}
		}
	},
		rightArrow,
	)
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

type comboboxOption struct {
	Text string
	Cl   widget.Clickable
}
type comboboxChip struct {
	Text string
	Tag  struct{}
}
type Comboboxable struct {
	Inputable
	optionsLs          widget.List
	options            []comboboxOption
	selected           string
	chips              []comboboxChip
	removedIdx         int
	hasRemoved         bool
	isChipCloseHovered bool
}

func (c *Comboboxable) HandleOptionEvents(gtx layout.Context, idx int) {
	curOption := &c.options[idx]
	if curOption.Cl.Clicked(gtx) {
		c.setSelectedValue(curOption.Text)
		// When click event is being read "selectedValue" was already checked
		// and since Gio's idiomatic way is to check for events at the start
		// of the frame, there is no other way except request a new frame
		gtx.Execute(op.InvalidateCmd{})
	}
	if curOption.Cl.Hovered() {
		SetCursor(gtx, pointer.CursorPointer)
	}
}

func (c *Comboboxable) HandleChipEvents(gtx layout.Context, idx int) {
	HandlePointerEvents(gtx, &c.chips[idx].Tag, pointer.Enter|pointer.Leave|pointer.Press, func(e pointer.Event) {
		switch e.Kind {
		case pointer.Enter:
			c.isChipCloseHovered = true
		case pointer.Leave:
			c.isChipCloseHovered = false
		case pointer.Press:
			c.setRemovedChip(idx)
			// Read "HandleOptionsEvents comment"
			gtx.Execute(op.InvalidateCmd{})
		}
	})
	if c.isChipCloseHovered {
		SetCursor(gtx, pointer.CursorPointer)
	}
}

func (c *Comboboxable) SetOptions(options []string) {
	for idx, it := range options {
		if idx < len(c.options) {
			c.options[idx].Text = it
		} else {
			c.options = append(c.options, comboboxOption{Text: it})
		}
	}
	c.options = c.options[:len(options)]
}

func (c *Comboboxable) SetChips(values []string) {
	for idx, it := range values {
		if idx < len(c.chips) {
			c.chips[idx].Text = it
		} else {
			c.chips = append(c.chips, comboboxChip{Text: it})
		}
	}
	c.chips = c.chips[:len(values)]
}

func (c *Comboboxable) ResetOptionScroll() {
	c.optionsLs.ScrollTo(0)
}

func (c *Comboboxable) setSelectedValue(v string) {
	c.selected = v
}

func (c *Comboboxable) setRemovedChip(idx int) {
	c.removedIdx = idx
	c.hasRemoved = true
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

func (c *Comboboxable) HasRemovedChip() int {
	if !c.hasRemoved {
		return -1
	}
	c.hasRemoved = false
	return c.removedIdx
}

func (c *Comboboxable) Blur(gtx layout.Context) {
	c.optionsLs.ScrollTo(0)
	c.Inputable.Blur(gtx)
}

type TableProps[T any] struct {
	Axis                 layout.Axis
	HeaderCellsAlignment []layout.Direction
	RowCellsAlignment    []layout.Direction
	RowValueCb           func(int) T
	RowFilterCb          func(T) bool
}

const tableColumnsInitAmount = 16

func NewTable[T any](props TableProps[T]) *Table[T] {
	if props.RowValueCb == nil {
		log.Fatal("You should provide RowValueCb for the table")
	}
	if props.RowFilterCb == nil {
		props.RowFilterCb = func(t T) bool { return true }
	}
	l := widget.List{}
	l.Axis = props.Axis
	return &Table[T]{
		list:             l,
		columnWidths:     make([]int, 0, tableColumnsInitAmount),
		hCellsAllignment: props.HeaderCellsAlignment,
		headCellFuncs:    make([]HeadCellComp, 0, tableColumnsInitAmount),
		rCellsAllignment: props.RowCellsAlignment,
		rowCellFuncs:     make([]CellComp[T], 0, tableColumnsInitAmount),
		cellsBuf:         make([]layout.FlexChild, 0, tableColumnsInitAmount),
		rowsVisibility:   make([]bool, 0, 32),
		rowValueCb:       props.RowValueCb,
		rowFilterCb:      props.RowFilterCb,
	}
}

type HeadCellComp func(gtx layout.Context) layout.Dimensions
type CellComp[T any] func(gtx layout.Context, rowIdx int, rowValue T) layout.Dimensions
type Table[T any] struct {
	rowsAmount       int
	rowValueCb       func(int) T
	rowFilterCb      func(T) bool
	rowsVisibility   []bool
	rowsSkipped      int
	cellsBuf         []layout.FlexChild
	list             widget.List
	columnWidths     []int
	hCellsAllignment []layout.Direction
	rCellsAllignment []layout.Direction
	headCellFuncs    []HeadCellComp
	rowCellFuncs     []CellComp[T]
	BottomMargin     bool
	RefineFilterTxt  func() string
}

func (t *Table[T]) HeadCells(hFuncs ...HeadCellComp) {
	t.headCellFuncs = hFuncs
}

func (t *Table[T]) RowCells(rFuncs ...CellComp[T]) {
	if len(rFuncs) != len(t.headCellFuncs) {
		log.Fatal("Incorrect usage of table! Row: number of cell render functions are not equal to set headers amount")
	}
	t.rowCellFuncs = rFuncs
}

func (t *Table[T]) prefilterRows(rowsAmount int) {
	t.rowsVisibility = t.rowsVisibility[:0]
	t.rowsSkipped = 0
	for idx := range rowsAmount {
		if !t.rowFilterCb(t.rowValueCb(idx)) {
			t.rowsVisibility = append(t.rowsVisibility, false)
			t.rowsSkipped++
		} else {
			t.rowsVisibility = append(t.rowsVisibility, true)
		}
	}
}

const tableXMargin = 8
const tableYMargin = 12

func (t *Table[T]) Layout(gtx layout.Context, th *theme.RepeatTheme, rowsAmount int, colWidths []int) {
	t.rowsAmount = rowsAmount
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
	t.columnWidths = t.columnWidths[:0]
	for idx, it := range colWidths {
		colSum += it
		if colSum > 100 {
			log.Fatal("Learn to count! Column width percentage sum is more than 100")
		}
		v := it * maxX / 100
		cellWidthSum += v
		t.columnWidths = append(t.columnWidths, v)
		if idx < len(colWidths)-1 {
			OffsetBy(gtx, image.Pt(cellWidthSum+xMargin, yMargin), func(gtx layout.Context) {
				DrawDivider(gtx, th, DividerProps{Axis: Vertical})
			})
		}
	}

	t.prefilterRows(rowsAmount)

	if cellWidthSum < gtx.Constraints.Max.X {
		t.columnWidths[len(t.columnWidths)-1] += gtx.Constraints.Max.X - cellWidthSum
	}
	OffsetBy(gtx, image.Pt(xMargin, yMargin), func(gtx layout.Context) {
		gtx.Constraints.Max.Y -= yMargin * 2
		t.layout(gtx, th, yMargin)
	})
}

type tableStyle struct {
	headerH          unit.Dp
	cellMargin       unit.Dp
	rowHeight        unit.Dp
	noEntriesMarginT unit.Dp
}

func defaultTableStyle() tableStyle {
	return tableStyle{
		headerH:          42,
		cellMargin:       5,
		rowHeight:        50,
		noEntriesMarginT: 3,
	}
}

func (t *Table[T]) drawEmptyRowInfo(gtx layout.Context, th *theme.RepeatTheme, s tableStyle, info string) layout.Dimensions {
	textM, textDims := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
		txtStyle := material.H6(th.Theme, info)
		txtStyle.Alignment = text.Middle
		txtStyle.Font.Style = font.Italic
		gtx.Constraints.Min = image.Pt(gtx.Constraints.Max.X, 0)
		return layout.UniformInset(s.cellMargin).Layout(gtx, txtStyle.Layout)
	})
	var bgDims layout.Dimensions
	marginT := gtx.Dp(s.noEntriesMarginT)
	OffsetBy(gtx, image.Pt(0, marginT), func(gtx layout.Context) {
		bgDims = layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			white := th.Palette.CardBg
			white.A = 0xbb
			return DrawBox(gtx, Box{
				Size:  image.Rect(0, 0, textDims.Size.X, textDims.Size.Y),
				Color: white,
			})
		})
		bgDims.Size.Y += marginT
		textM.Add(gtx.Ops)
	})
	return bgDims
}

func (t *Table[T]) layout(gtx layout.Context, th *theme.RepeatTheme, bottomMargin int) {
	s := defaultTableStyle()
	headerH := gtx.Dp(s.headerH)
	t.cellsBuf = t.cellsBuf[:0]
	for colIdx, it := range t.headCellFuncs {
		fxChild := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			columnDims := layout.Dimensions{Size: image.Pt(t.columnWidths[colIdx], headerH)}
			gtx.Constraints.Max = columnDims.Size
			gtx.Constraints.Min = columnDims.Size
			cellAl := layout.Center
			if colIdx < len(t.hCellsAllignment) {
				cellAl = t.hCellsAllignment[colIdx]
			}
			layout.UniformInset(s.cellMargin).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return cellAl.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return it(gtx)
				})
			})
			return columnDims
		})
		t.cellsBuf = append(t.cellsBuf, fxChild)
	}
	layout.Flex{}.Layout(gtx, t.cellsBuf...)

	rowH := gtx.Dp(s.rowHeight)
	OffsetBy(gtx, image.Pt(0, headerH), func(gtx layout.Context) {
		DrawDivider(gtx, th, DividerProps{})
		gtx.Constraints.Max.Y -= bottomMargin
		rowsAmount := t.rowsAmount
		if t.rowsSkipped == t.rowsAmount {
			rowsAmount = min(t.rowsAmount, 1)
		}
		material.List(th.Theme, &t.list).Layout(gtx, rowsAmount, func(gtx layout.Context, rowIdx int) layout.Dimensions {
			rowValue := t.rowValueCb(rowIdx)
			if !t.rowsVisibility[rowIdx] {
				if t.rowsSkipped == t.rowsAmount {
					noMatches := "no matches, refine filters"
					if t.RefineFilterTxt != nil {
						noMatches = t.RefineFilterTxt()
					}
					return t.drawEmptyRowInfo(gtx, th, s, noMatches)
				}
				if rowIdx == t.rowsAmount-1 {
					return t.drawEmptyRowInfo(gtx, th, s, "...")
				}
				return layout.Dimensions{}
			}
			for colIdx, it := range t.rowCellFuncs {
				t.cellsBuf[colIdx] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					columnDims := layout.Dimensions{Size: image.Pt(t.columnWidths[colIdx], rowH)}
					cellAl := layout.Center
					if colIdx < len(t.rCellsAllignment) {
						cellAl = t.rCellsAllignment[colIdx]
					}
					gtx.Constraints.Max = columnDims.Size
					gtx.Constraints.Min = columnDims.Size

					layout.UniformInset(s.cellMargin).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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

// NOTE: This is not needed
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

func (d *Dialog) Info(th *theme.RepeatTheme, title string, w func(gtx layout.Context) layout.Dimensions) {
	d.CancelProps = dialogButton{}
	d.OkProps = dialogButton{}
	d.th = th
	d.title = title
	d.content = w
	d.icon = micons.Info
	d.iconC = th.Palette.Mimosa
	d.CancelProps.IsHidden = true
}

func (d *Dialog) Basic(th *theme.RepeatTheme, title string, w func(gtx layout.Context) layout.Dimensions) {
	d.CancelProps = dialogButton{}
	d.OkProps = dialogButton{}
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

func (d *Dialog) SetLabels(cancel, ok string) {
	d.CancelProps.Text = cancel
	d.OkProps.Text = ok
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
	requestedFocus Focusable
	requestedBlur  Focusable
	scrimTag       struct{}
}

func (fm *FocusManager) blur(gtx layout.Context) {
	if fm.requestedBlur != nil {
		fm.requestedBlur.Blur(gtx)
	}
	if fm.requestedBlur == fm.inFocus {
		fm.inFocus = nil
	}
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
			fm.RequestBlur(nil)
		}
	})
	if (fm.requestedFocus != nil) && (fm.requestedBlur != nil) && (fm.requestedFocus == fm.requestedBlur) {
		fm.requestedBlur, fm.requestedFocus = nil, nil
		return
	}
	if fm.requestedBlur != nil {
		fm.blur(gtx)
	}
	if fm.requestedFocus != nil {
		fm.focus(gtx)
	}
}

// Request focus programatically
func (fm *FocusManager) RequestFocus(it Focusable) {
	fm.requestedFocus = it
}

// Request Blur programatically
func (fm *FocusManager) RequestBlur(it Focusable) {
	if it == nil {
		it = fm.inFocus
	}
	fm.requestedBlur = it
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

func NewI18nSwitcher(l i18n.Lang, fm *FocusManager) I18nSwitcher {
	switcher := I18nSwitcher{
		fm: fm,
		en: I18nMenuOption{
			Lang: i18n.En,
		},
		ru: I18nMenuOption{
			Lang: i18n.Ru,
		},
	}
	if l == i18n.En {
		switcher.Active = switcher.en
	} else {
		switcher.Active = switcher.ru
	}
	return switcher
}

type I18nSwitcher struct {
	en     I18nMenuOption
	ru     I18nMenuOption
	fm     Focuser
	Active I18nMenuOption
	Open   bool
}

func (i *I18nSwitcher) isOptionClicked(gtx layout.Context, o *I18nMenuOption) bool {
	isClicked := false
	HandlePointerEvents(gtx, &o.Tag, pointer.Press, func(e pointer.Event) {
		if e.Kind == pointer.Press {
			isClicked = true
		}
	})
	return isClicked
}

func (i *I18nSwitcher) isOptionHovered(gtx layout.Context, o *I18nMenuOption) bool {
	isHovered := o.Hovered
	HandlePointerEvents(gtx, &o.Tag, pointer.Enter|pointer.Leave, func(e pointer.Event) {
		switch e.Kind {
		case pointer.Enter:
			isHovered = true
		case pointer.Leave:
			isHovered = false
		}
	})
	return isHovered
}

func (i *I18nSwitcher) handleHover(gtx layout.Context) {
	i.en.Hovered = i.isOptionHovered(gtx, &i.en)
	i.ru.Hovered = i.isOptionHovered(gtx, &i.ru)
	i.Active.Hovered = i.isOptionHovered(gtx, &i.Active)
}

func (i *I18nSwitcher) toggleState() {
	i.Open = !i.Open
	if i.Open {
		i.fm.RequestFocus(i)
	} else {
		i.fm.RequestBlur(i)
	}
}

func (i *I18nSwitcher) Update(gtx layout.Context) (i18n.Lang, bool) {
	i.handleHover(gtx)
	if i.isOptionClicked(gtx, &i.Active) {
		i.toggleState()
	}
	active := i.Active.Lang
	if i.isOptionClicked(gtx, &i.en) {
		i.Active = i.en
		i.Open = false
	}
	if i.isOptionClicked(gtx, &i.ru) {
		i.Active = i.ru
		i.Open = false
	}
	if active != i.Active.Lang {
		// Active is a copy of clicked (and ofc hovered) option, so Hovered should be reset since Active !== Clicked
		i.Active.Hovered = false
		return i.Active.Lang, true
	}
	return active, false
}

func (i *I18nSwitcher) GetSecondaryOption() *I18nMenuOption {
	if i.Active.Lang == i.en.Lang {
		return &i.ru
	} else {
		return &i.en
	}
}

func (i *I18nSwitcher) GetCursorType() (pointer.Cursor, bool) {
	return pointer.CursorPointer, i.Active.Hovered || i.en.Hovered || i.ru.Hovered
}

func (i *I18nSwitcher) Focus(gtx layout.Context) {
	i.Open = true
	gtx.Execute(op.InvalidateCmd{})
}

func (i *I18nSwitcher) Blur(gtx layout.Context) {
	i.Open = false
	gtx.Execute(op.InvalidateCmd{})
}

type Hyperlinkable struct {
	tag       struct{}
	isHovered bool
	isPressed bool
}

func (h *Hyperlinkable) update(gtx layout.Context) {
	HandlePointerEvents(gtx, &h.tag, pointer.Enter|pointer.Leave|pointer.Press, func(e pointer.Event) {
		switch e.Kind {
		case pointer.Enter:
			h.isHovered = true
		case pointer.Leave:
			h.isHovered = false
		case pointer.Press:
			h.isPressed = true
		}
	})
}

func (h *Hyperlinkable) IsHovered() bool {
	return h.isHovered
}

func (h *Hyperlinkable) IsPressed() bool {
	v := h.isPressed
	h.isPressed = false
	return v
}

// TODO: Rename all "GetCursorType" -> "Cursor"
func (h *Hyperlinkable) GetCursorType() (pointer.Cursor, bool) {
	if h.isHovered {
		return pointer.CursorPointer, true
	}
	return pointer.CursorDefault, false
}
