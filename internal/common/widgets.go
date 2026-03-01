package common

import (
	"image"
	"log"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/ui/theme"
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

type TableProps struct {
	Axis                 layout.Axis
	ColumsNum            int
	HeaderCellsAlignment []layout.Direction
	RowCellsAlignment    []layout.Direction
}

func NewTable(props TableProps) *Table {
	l := widget.List{}
	l.Axis = props.Axis
	return &Table{
		columns:          props.ColumsNum,
		list:             &l,
		columnWidths:     make([]int, props.ColumsNum),
		hCellsAllignment: props.HeaderCellsAlignment,
		headCellFuncs:    make([]HeadCellComp, props.ColumsNum),
		rCellsAllignment: props.RowCellsAlignment,
		rowCellFuncs:     make([]CellComp, props.ColumsNum),
		cellsBuf:         make([]layout.FlexChild, props.ColumsNum),
	}
}

type HeadCellComp func(gtx layout.Context) layout.Dimensions
type CellComp func(gtx layout.Context, rowIdx, colIdx int) layout.Dimensions
type Table struct {
	columns          int
	Rows             int
	cellsBuf         []layout.FlexChild
	list             *widget.List
	columnWidths     []int
	hCellsAllignment []layout.Direction
	rCellsAllignment []layout.Direction
	headCellFuncs    []HeadCellComp
	rowCellFuncs     []CellComp
	BottomMargin     bool
}

func (t *Table) HeadCells(hFuncs ...HeadCellComp) {
	if len(hFuncs) != t.columns {
		log.Fatal("Incorrect usage of table! Header: number of cell render functions are not equal to set columns amount")
	}
	t.headCellFuncs = hFuncs
}

func (t *Table) RowCells(rFuncs ...CellComp) {
	if len(rFuncs) != t.columns {
		log.Fatal("Incorrect usage of table! Row: number of cell render functions are not equal to set columns amount")
	}
	t.rowCellFuncs = rFuncs
}

const tableXMargin = 8
const tableYMargin = 12

func (t *Table) Layout(gtx layout.Context, th *theme.RepeatTheme, colWidths []int) {
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
	maxX := gtx.Constraints.Max.X
	for idx, it := range colWidths {
		v := it * maxX / 100
		cellWidthSum += v
		t.columnWidths[idx] = v
		if idx < len(colWidths)-1 {
			OffsetBy(gtx, image.Pt(cellWidthSum+xMargin, yMargin), func() {
				DrawDivider(gtx, th, DividerProps{Axis: Vertical})
			})
		}

	}
	if cellWidthSum < gtx.Constraints.Max.X {
		t.columnWidths[len(t.columnWidths)-1] += gtx.Constraints.Max.X - cellWidthSum
	}
	OffsetBy(gtx, image.Pt(xMargin, yMargin), func() {
		t.layout(gtx, th)
	})
}

const headerHDP = 40
const cellMarginDP = 5
const rowHeightDP = 38

func (t *Table) layout(gtx layout.Context, th *theme.RepeatTheme) {
	ls := material.List(th.Theme, t.list)
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
	OffsetBy(gtx, image.Pt(0, headerH), func() {
		DrawDivider(gtx, th, DividerProps{})
		ls.Layout(gtx, t.Rows, func(gtx layout.Context, rowIdx int) layout.Dimensions {
			for colIdx, it := range t.rowCellFuncs {
				t.cellsBuf[colIdx] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					columnDims := layout.Dimensions{Size: image.Pt(t.columnWidths[colIdx], rowH)}
					cellAl := t.rCellsAllignment[colIdx]
					gtx.Constraints.Max = columnDims.Size
					gtx.Constraints.Min = columnDims.Size

					layout.UniformInset(cellMarginDP).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return cellAl.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							// TODO: Do you really need column index?
							return it(gtx, rowIdx, colIdx)
						})
					})
					return columnDims
				})
			}
			return layout.Flex{}.Layout(gtx, t.cellsBuf...)
		})
	})
}
