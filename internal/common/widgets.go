package common

import (
	"image"
	"image/color"
	"log"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

type Inputable struct {
	isHovered    bool
	isFocused    bool
	isDirty      bool
	value        string
	hasSelection bool
	hasSubmitted bool
	widget.Editor
	widget.Clickable
	Cancel widget.Clickable
}

func (in *Inputable) Update(gtx layout.Context) {
	in.processEditorEvents(gtx)

	if in.Cancel.Clicked(gtx) {
		in.Editor.SetText("")
		in.value = ""
		in.Blur(gtx)
	}
	HandlePointerEvents(gtx, &in.Editor, pointer.Press|pointer.Move|pointer.Leave, func(e pointer.Event) {
		switch e.Kind {
		case pointer.Move:
			in.isHovered = true
		case pointer.Leave:
			in.isHovered = false
		case pointer.Press:
			// Even if user missed the input field and pressed container the caret will be set anyway
			if in.Focus(gtx) {
				if e.Position.X < 100 {
					in.Editor.SetCaret(0, 0)
				} else {
					txtLen := len(in.GetInput())
					in.Editor.SetCaret(txtLen, txtLen)
				}
			}
		}
	})

	HandleKeyEvents(gtx, func(e key.Event) {
		switch e.Name {
		case key.NameEscape:
			in.Blur(gtx)
		}
	},
		key.Filter{Name: key.NameEscape},
	)
}

func (in *Inputable) Subscribe(gtx layout.Context) {
	defer pointer.PassOp{}.Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, &in.Editor)
}

func (in *Inputable) Blur(gtx layout.Context) {
	gtx.Execute(key.FocusCmd{Tag: nil})
	in.isFocused = false
}

func (in *Inputable) Focus(gtx layout.Context) (wasFocusedBefore bool) {
	if in.isFocused {
		return true
	}
	gtx.Execute(key.FocusCmd{Tag: &in.Editor})
	txtLen := len(in.Editor.Text())
	if txtLen > 0 {
		in.Editor.SetCaret(txtLen, 0)
	}
	in.isFocused = true
	return false
}

func (in *Inputable) GetInput() string {
	if in.isDirty {
		in.value = in.Editor.Text()
		in.isDirty = false
	}
	return in.value
}

func (in *Inputable) HasSelection() bool {
	if in.hasSelection {
		in.hasSelection = !in.hasSelection
		return !in.hasSelection
	}
	return false
}

func (in *Inputable) HasSubmit() bool {
	if in.hasSubmitted {
		in.hasSubmitted = !in.hasSubmitted
		return !in.hasSubmitted
	}
	return false
}

func (in *Inputable) IsHovered() bool {
	return in.isHovered
}

func (in *Inputable) IsFocused() bool {
	return in.isFocused
}

func (in *Inputable) GetCursorType() (cursor pointer.Cursor, ok bool) {
	if in.IsHovered() {
		if in.IsFocused() {
			return pointer.CursorText, true
		} else {
			return pointer.CursorPointer, true
		}
	}
	if in.Cancel.Hovered() {
		return pointer.CursorPointer, true
	}
	return pointer.CursorDefault, false
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
		case widget.SelectEvent:
			in.hasSelection = true
		case widget.SubmitEvent:
			in.hasSubmitted = true
		}
	}
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

const headerHDP = 40
const cellMarginDP = 5
const rowHeightDP = 50

func (t *Table[T]) layout(gtx layout.Context, th *theme.RepeatTheme, bottomMargin int) {
	headerH := gtx.Dp(headerHDP)
	for colIdx, it := range t.headCellFuncs {
		t.cellsBuf[colIdx] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			columnDims := layout.Dimensions{Size: image.Pt(t.columnWidths[colIdx], headerH)}
			gtx.Constraints.Max = columnDims.Size
			gtx.Constraints.Min = columnDims.Size
			cellAl := t.hCellsAllignment[colIdx]
			layout.UniformInset(cellMarginDP).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return cellAl.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return it(gtx)
				})
			})
			return columnDims
		})
	}
	layout.Flex{}.Layout(gtx, t.cellsBuf...)

	rowH := gtx.Dp(rowHeightDP)
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

					layout.UniformInset(cellMarginDP).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
	th          *theme.RepeatTheme
	isOpen      bool
	variant     dialogType
	Ok          widget.Clickable
	OkProps     dialogButton
	Cancel      widget.Clickable
	Scrim       widget.Clickable
	CancelProps dialogButton
	title       string
	icon        *widget.Icon
	iconC       color.NRGBA
	hasIcon     bool
	content     func(layout.Context) layout.Dimensions
	contentLs   widget.List
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
}

func (d *Dialog) Hide() {
	d.isOpen = false
	d.icon = nil
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

	prevConstr := gtx.Constraints
	contentM, contentDims := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X = maxW - fullPadd
		gtx.Constraints.Min.X = minW
		gtx.Constraints.Min.Y = 0
		dims := d.content(gtx)
		dims.Size.X += fullPadd
		return dims
	})

	innerM, innerDims := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
		currentWidth := Clamp(minW, contentDims.Size.X, maxW)
		gtx.Constraints.Max.X = currentWidth
		gtx.Constraints.Max.Y = maxH - fullPadd
		var incrDims layout.Dimensions
		{
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
	gtx.Constraints = prevConstr

	gtx.Constraints.Min = gtx.Constraints.Max
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X -= fullPadd
		gtx.Constraints.Max.Y -= fullPadd
		size := image.Rect(0, 0, Clamp(minW, innerDims.Size.X+fullPadd, maxW), innerDims.Size.Y)
		dialogDims := DrawBox(gtx, Box{
			Size:    size,
			Color:   d.th.Palette.CardBg,
			R:       theme.CornerR(shape, shape, shape, shape),
			HideInk: true,
		})
		RegisterTag(gtx, &d, size)
		OffsetBy(gtx, image.Pt(padd, padd), func(gtx layout.Context) {
			innerM.Add(gtx.Ops)
		})
		return dialogDims
	})
}

func (d *Dialog) GetCursorType() (pointer.Cursor, bool) {
	if d.Scrim.Hovered() || d.Cancel.Hovered() || d.Ok.Hovered() {
		return pointer.CursorPointer, true
	}
	return pointer.CursorDefault, false
}
