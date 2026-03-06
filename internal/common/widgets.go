package common

import (
	"image"
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

type Searchable struct {
	isHovered bool
	isFocused bool
	isDirty   bool
	value     string
	widget.Editor
	widget.Clickable
	Cancel widget.Clickable
}

func (s *Searchable) Update(gtx layout.Context) {
	if s.gotDirty(gtx) {
		s.isDirty = true
	}

	if s.Cancel.Clicked(gtx) {
		s.Editor.SetText("")
		s.Blur(gtx)
	}
	HandlePointerEvents(gtx, &s.Editor, pointer.Press|pointer.Move|pointer.Leave, func(e pointer.Event) {
		switch e.Kind {
		case pointer.Move:
			s.isHovered = true
		case pointer.Leave:
			s.isHovered = false
		case pointer.Press:
			// Even if user missed the input field and pressed container the caret will be set anyway
			if s.Focus(gtx) {
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
			s.Blur(gtx)
		}
	},
		key.Filter{Name: key.NameEscape},
	)
}

func (s *Searchable) Subscribe(gtx layout.Context) {
	defer pointer.PassOp{}.Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, &s.Editor)
}

func (s *Searchable) Blur(gtx layout.Context) {
	gtx.Execute(key.FocusCmd{Tag: nil})
	s.isFocused = false
}

func (s *Searchable) Focus(gtx layout.Context) (wasFocusedBefore bool) {
	if s.isFocused {
		return true
	}
	gtx.Execute(key.FocusCmd{Tag: &s.Editor})
	txtLen := len(s.Editor.Text())
	if txtLen > 0 {
		s.Editor.SetCaret(txtLen, 0)
	}
	s.isFocused = true
	return false
}

func (s *Searchable) GetInput() string {
	if s.isDirty {
		s.value = s.Editor.Text()
		s.isDirty = false
	}
	return s.value
}

func (s *Searchable) IsHovered() bool {
	return s.isHovered
}

func (s *Searchable) IsFocused() bool {
	return s.isFocused
}

func (s *Searchable) GetCursorType() (cursor pointer.Cursor, ok bool) {
	if s.IsHovered() {
		if s.IsFocused() {
			return pointer.CursorText, true
		} else {
			return pointer.CursorPointer, true
		}
	}
	if s.Cancel.Hovered() {
		return pointer.CursorPointer, true
	}
	return pointer.CursorDefault, false
}

func (s *Searchable) gotDirty(gtx layout.Context) bool {
	for {
		we, ok := s.Editor.Update(gtx)
		if !ok {
			break
		}
		if _, ok := we.(widget.ChangeEvent); ok {
			return true
		}
	}
	return false
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
	hasIcon     bool
	content     func(layout.Context) layout.Dimensions
	contentLs   widget.List
}

func (d *Dialog) SetIcon(icon *widget.Icon) {
	d.icon = icon
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
	if d.th == nil {
		panic("Dialog: no theme available")
	}
	if !d.isOpen {
		return layout.Dimensions{}
	}

	bg := image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Max.Y)
	if d.Scrim.Hovered() {
		SetCursor(gtx, pointer.CursorPointer)
	}
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
		gtx.Constraints.Max.X = maxW
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
					d.icon.Layout(gtx, d.th.Palette.Backdrop)
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
						if d.Cancel.Hovered() {
							SetCursor(gtx, pointer.CursorPointer)
						}
						button = material.Button(d.th.Theme, &d.Cancel, txt)
						button.Background.A = 0x00
						button.Color = d.th.Palette.Backdrop
						cancelDims := button.Layout(gtx)
						dims = cancelDims
						dims.Size.X += betweenButtonsPad
					}
					if !d.OkProps.IsHidden {
						OffsetBy(gtx, image.Pt(dims.Size.X, 0), func(gtx layout.Context) {
							txt := "Ok"
							if d.OkProps.Text != "" {
								txt = d.OkProps.Text
							}
							if d.Ok.Hovered() {
								SetCursor(gtx, pointer.CursorPointer)
							}
							button = material.Button(d.th.Theme, &d.Ok, txt)
							button.Background.A = 0x00
							button.Color = d.th.Palette.Backdrop
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
