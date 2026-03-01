package common

import (
	"fmt"
	"image"

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

type HeadCellComp func(gtx layout.Context) layout.Dimensions
type CellComp func(gtx layout.Context, rowIdx, colIdx int) layout.Dimensions
type Table struct {
	Columns          int
	Rows             int
	CellsBuf         []layout.FlexChild
	List             *widget.List
	ColumnWidths     []int
	HCellsAllignment []layout.Direction
	RCellsAllignment []layout.Direction
	HeadCellFuncs    []HeadCellComp
	RowCellFuncs     []CellComp
	BottomMargin     bool
}

func (t *Table) HeadCells(hFuncs ...HeadCellComp) {
	t.HeadCellFuncs = hFuncs
}

func (t *Table) RowCells(rFuncs ...CellComp) {
	t.RowCellFuncs = rFuncs
}

const tableXMargin = 8
const tableYMargin = 8

func (t *Table) Layout(gtx layout.Context, th *theme.RepeatTheme) {
	gtx.Constraints.Min = image.Point{}
	DrawBox(gtx, Box{
		Size:  image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Max.Y),
		Color: th.Palette.CardBg,
		R:     theme.CornerR(0, 0, 20, 20),
	})
	xMargin, yMargin := gtx.Dp(tableXMargin), gtx.Dp(tableYMargin)

	OffsetBy(gtx, image.Pt(xMargin, yMargin), func() {
		gtx.Constraints.Max.X -= xMargin * 2
		gtx.Constraints.Max.Y -= yMargin
		if t.BottomMargin {
			gtx.Constraints.Max.Y -= yMargin
		}
		t.layout(gtx, th)
	})
}

const headerHDP = 42
const cellMarginDP = 5
const rowHeightDP = 38

func (t *Table) layout(gtx layout.Context, th *theme.RepeatTheme) {
	maxX := gtx.Constraints.Max.X
	ls := material.List(th.Theme, t.List)
	headerH := gtx.Dp(headerHDP)
	for colIdx, it := range t.HeadCellFuncs {
		t.CellsBuf[colIdx] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			curWidth := PrcToPx(maxX, fmt.Sprintf("%d%", t.ColumnWidths[colIdx]))
			columnDims := layout.Dimensions{Size: image.Pt(curWidth, headerH)}
			gtx.Constraints.Max = columnDims.Size
			gtx.Constraints.Min = columnDims.Size
			cellAl := t.RCellsAllignment[colIdx]
			layout.UniformInset(cellMarginDP).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return cellAl.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return it(gtx)
				})
			})
			return columnDims
		})
	}
	layout.Flex{}.Layout(gtx, t.CellsBuf...)

	rowH := gtx.Dp(rowHeightDP)
	OffsetBy(gtx, image.Pt(0, headerH), func() {
		DrawDivider(gtx, th, DividerProps{})
		ls.Layout(gtx, t.Rows, func(gtx layout.Context, rowIdx int) layout.Dimensions {
			for colIdx, it := range t.RowCellFuncs {
				t.CellsBuf[colIdx] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					curWidth := PrcToPx(maxX, fmt.Sprintf("%d%", t.ColumnWidths[colIdx]))
					columnDims := layout.Dimensions{Size: image.Pt(curWidth, rowH)}
					cellAl := t.RCellsAllignment[colIdx]
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
			return layout.Flex{}.Layout(gtx, t.CellsBuf...)
		})
	})
}
