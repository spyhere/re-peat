package common

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	micons "github.com/spyhere/re-peat/internal/mIcons"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

func colPremul(c, a byte) byte {
	return byte(uint16(c) * uint16(a) / 255)
}

// This is a copy-pasted function from gio's "widget/material/button.go"
func drawInk(gtx layout.Context, c widget.Press) {
	// duration is the number of seconds for the
	// completed animation: expand while fading in, then
	// out.
	const (
		expandDuration = float32(0.5)
		fadeDuration   = float32(0.9)
	)

	now := gtx.Now

	t := float32(now.Sub(c.Start).Seconds())

	end := c.End
	if end.IsZero() {
		// If the press hasn't ended, don't fade-out.
		end = now
	}

	endt := float32(end.Sub(c.Start).Seconds())

	// Compute the fade-in/out position in [0;1].
	var alphat float32
	{
		var haste float32
		if c.Cancelled {
			// If the press was cancelled before the inkwell
			// was fully faded in, fast forward the animation
			// to match the fade-out.
			if h := 0.5 - endt/fadeDuration; h > 0 {
				haste = h
			}
		}
		// Fade in.
		half1 := t/fadeDuration + haste
		if half1 > 0.5 {
			half1 = 0.5
		}

		// Fade out.
		half2 := float32(now.Sub(end).Seconds())
		half2 /= fadeDuration
		half2 += haste
		if half2 > 0.5 {
			// Too old.
			return
		}

		alphat = half1 + half2
	}

	// Compute the expand position in [0;1].
	sizet := t
	if c.Cancelled {
		// Freeze expansion of cancelled presses.
		sizet = endt
	}
	sizet /= expandDuration

	// Animate only ended presses, and presses that are fading in.
	if !c.End.IsZero() || sizet <= 1.0 {
		gtx.Execute(op.InvalidateCmd{})
	}

	if sizet > 1.0 {
		sizet = 1.0
	}

	if alphat > .5 {
		// Start fadeout after half the animation.
		alphat = 1.0 - alphat
	}
	// Twice the speed to attain fully faded in at 0.5.
	t2 := alphat * 2
	// Beziér ease-in curve.
	alphaBezier := t2 * t2 * (3.0 - 2.0*t2)
	sizeBezier := sizet * sizet * (3.0 - 2.0*sizet)
	size := gtx.Constraints.Min.X
	if h := gtx.Constraints.Min.Y; h > size {
		size = h
	}
	// Cover the entire constraints min rectangle and
	// apply curve values to size and color.
	size = int(float32(size) * 2 * float32(math.Sqrt(2)) * sizeBezier)
	alpha := 0.7 * alphaBezier
	const col = 0.8
	ba, bc := byte(alpha*0xff), byte(col*0xff)
	rgba := color.NRGBA{
		A: ba,
		R: colPremul(bc, ba),
		G: colPremul(bc, ba),
		B: colPremul(bc, ba),
	}
	ink := paint.ColorOp{Color: rgba}
	ink.Add(gtx.Ops)
	rr := size / 2
	defer op.Offset(c.Position.Add(image.Point{
		X: -rr,
		Y: -rr,
	})).Push(gtx.Ops).Pop()
	defer clip.UniformRRect(image.Rectangle{Max: image.Pt(size, size)}, rr).Push(gtx.Ops).Pop()
	paint.PaintOp{}.Add(gtx.Ops)
}

type Box struct {
	Size       image.Rectangle
	Color      color.NRGBA
	R          theme.CornerRadii
	StrokeC    color.NRGBA
	StrokeW    unit.Dp
	Clickable  *widget.Clickable
	HideInk    bool
	GeometryCb func()
}

func DrawBox(gtx layout.Context, b Box) layout.Dimensions {
	r := b.R
	rrect := clip.RRect{Rect: b.Size, SE: r.SE, SW: r.SW, NE: r.NE, NW: r.NW}
	rrectStack := rrect.Push(gtx.Ops)
	paint.ColorOp{Color: b.Color}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	if b.Clickable != nil {
		b.Clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: rrect.Rect.Max}
		})
		if !b.HideInk {
			for _, it := range b.Clickable.History() {
				gtx.Constraints.Min = rrect.Rect.Max
				drawInk(gtx, it)
			}
		}
	}
	if b.GeometryCb != nil {
		b.GeometryCb()
	}
	rrectStack.Pop()

	if b.StrokeW != 0 {
		half := int(float32(gtx.Dp(b.StrokeW)) / 2)
		rrect.Rect.Min.X += half
		rrect.Rect.Min.Y += half
		rrect.Rect.Max.X -= half
		rrect.Rect.Max.Y -= half
		stroke := clip.Stroke{
			Path:  rrect.Path(gtx.Ops),
			Width: float32(gtx.Dp(b.StrokeW)),
		}.Op().Push(gtx.Ops)
		paint.ColorOp{Color: b.StrokeC}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		stroke.Pop()
	}
	return layout.Dimensions{Size: b.Size.Size()}
}

func DrawBackground(gtx layout.Context, col color.NRGBA) {
	DrawBox(gtx, Box{
		Size:  image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Max.Y),
		Color: col,
	})
}

type searchMaterialSpecs struct {
	height         unit.Dp
	minWidth       unit.Dp
	maxWidth       unit.Dp
	xPadding       unit.Dp
	iconXPadding   unit.Dp
	iconSize       unit.Dp
	fontFace       font.Typeface
	fontWeight     font.Weight
	fontSize       unit.Sp
	fontLineHeight unit.Sp
}

var searchSpecs = searchMaterialSpecs{
	height:         56,
	minWidth:       360,
	maxWidth:       720,
	xPadding:       16,
	iconXPadding:   16,
	iconSize:       24,
	fontFace:       "Roboto",
	fontWeight:     400,
	fontSize:       16,
	fontLineHeight: 24,
}

type SProps struct {
	DefaultText string
	*Inputable
}

func DrawSearch(gtx layout.Context, th *theme.RepeatTheme, props SProps) layout.Dimensions {
	props.Inputable.Update(gtx)

	containerH := gtx.Dp(searchSpecs.height)
	containerHHalft := containerH / 2
	containerW := gtx.Dp(searchSpecs.minWidth)
	// Background
	bDims := DrawBox(gtx, Box{
		Size:       image.Rect(0, 0, containerW, containerH),
		Color:      th.Palette.Search.Enabled.Bg,
		R:          theme.CornerR(containerHHalft, containerHHalft, containerHHalft, containerHHalft),
		Clickable:  &props.Clickable,
		GeometryCb: func() { props.Inputable.Subscribe(gtx) },
	})
	// Hovered layer
	if props.IsHovered() {
		DrawBox(gtx, Box{
			Size:  image.Rect(0, 0, containerW, containerH),
			Color: th.Palette.Search.Hovered.Bg,
			R:     theme.CornerR(containerHHalft, containerHHalft, containerHHalft, containerHHalft),
		})
	} else if props.IsFocused() {
		// Focused layer
		DrawBox(gtx, Box{
			Size:  image.Rect(0, 0, containerW, containerH),
			Color: th.Palette.Search.Pressed.Bg,
			R:     theme.CornerR(containerHHalft, containerHHalft, containerHHalft, containerHHalft),
		})
	}

	xPadd := gtx.Dp(searchSpecs.xPadding)
	iconSz, iconPadding := gtx.Dp(searchSpecs.iconSize), gtx.Dp(searchSpecs.iconXPadding)

	// Inner text
	c := th.Palette.Search.Enabled
	if props.IsFocused() {
		c = th.Palette.Search.Pressed
	}
	textH := gtx.Sp(searchSpecs.fontLineHeight)
	OffsetBy(gtx, image.Pt(xPadd*2, bDims.Size.Y-textH-textH/2), func(gtx layout.Context) {
		gtx.Constraints.Max = image.Pt(bDims.Size.X-xPadd*2-iconSz-iconPadding*2, textH)
		props.Editor.SingleLine = true
		text := props.GetInput()
		if text == "" {
			text = props.DefaultText
		}
		ed := material.Editor(th.Theme, &props.Editor, text)
		ed.Font.Typeface = "Roboto"
		ed.Color = c.Text
		ed.HintColor = th.Palette.Search.Enabled.Text
		ed.LineHeight = searchSpecs.fontLineHeight
		ed.TextSize = searchSpecs.fontSize
		ed.Font.Weight = 400
		passOp := pointer.PassOp{}.Push(gtx.Ops)
		ed.Layout(gtx)
		passOp.Pop()
	})

	OffsetBy(gtx, image.Pt(bDims.Size.X-iconPadding-xPadd-iconSz/2, bDims.Size.Y/2-iconSz/2), func(gtx layout.Context) {
		gtx.Constraints.Min.X = iconSz
		if len(props.GetInput()) > 0 {
			micons.Cancel.Layout(gtx, th.Palette.Search.Enabled.Icon)
			DrawBox(gtx, Box{
				Size:      image.Rect(0, 0, iconSz, iconSz),
				Clickable: &props.Cancel,
			})
		} else {
			micons.Search.Layout(gtx, th.Palette.Search.Enabled.Icon)
		}
	})
	return layout.Dimensions{Size: image.Pt(containerW, containerH)}
}

type inputFieldMaterialSpecs struct {
	shape             int
	defaultH          unit.Dp
	yPadding          unit.Dp
	outterIconPadding unit.Dp
	textXPadding      unit.Dp
	supTextPadding    unit.Dp
	supTextTopPadding unit.Dp
	bLineFocused      unit.Dp
	bLineUnfocused    unit.Dp
	icon              unit.Dp
	lblSizeBig        unit.Sp
	lblHeightBig      unit.Sp
	lblSizeSmall      unit.Sp
	lblHeightSmall    unit.Sp
	lblWeight         font.Weight
	supTxtSize        unit.Sp
	supTxtHeight      unit.Sp
}

var inputSpecs = inputFieldMaterialSpecs{
	shape:             4,
	defaultH:          56,
	yPadding:          8,
	outterIconPadding: 12,
	textXPadding:      16,
	supTextPadding:    16,
	supTextTopPadding: 4,
	bLineFocused:      4,
	bLineUnfocused:    1,
	icon:              24,
	lblSizeBig:        16,
	lblHeightBig:      24,
	lblSizeSmall:      12,
	lblHeightSmall:    16,
	lblWeight:         400,
	supTxtSize:        12,
	supTxtHeight:      16,
}

type InputFieldBase struct {
	LabelText    string
	LeadingIcon  *widget.Icon
	TrailingIcon *widget.Icon
	*Inputable
}

type inputFieldBaseProps struct {
	InputFieldBase
	content      func(gtx layout.Context, c theme.InputFieldPalette) layout.Dimensions
	chipsPresent bool
}

func drawInputFieldBase(gtx layout.Context, th *theme.RepeatTheme, props inputFieldBaseProps) layout.Dimensions {
	props.Inputable.Update(gtx)

	yPadding, defaultH := gtx.Dp(inputSpecs.yPadding), gtx.Dp(inputSpecs.defaultH)
	outterIconPadding, textXPadding := gtx.Dp(inputSpecs.outterIconPadding), gtx.Dp(inputSpecs.textXPadding)
	supTextPadding, supTextTopPadding := gtx.Dp(inputSpecs.supTextPadding), gtx.Dp(inputSpecs.supTextTopPadding)

	iconSize := gtx.Dp(inputSpecs.icon)
	// Determine dimensions
	iconWidthSum := 0
	if props.LeadingIcon != nil {
		iconWidthSum += iconSize + outterIconPadding
	}
	if props.TrailingIcon != nil {
		iconWidthSum += iconSize + outterIconPadding
	}
	c := th.Palette.Input.Enabled
	contentM, contentDims := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min = image.Point{}
		gtx.Constraints.Max.X -= iconWidthSum + textXPadding*2
		return props.content(gtx, c)
	})
	var defaultContentH unit.Sp = inputSpecs.lblHeightBig
	contentH := max(gtx.Sp(defaultContentH), contentDims.Size.Y)
	height := max(defaultH, textXPadding*2+contentH)

	// Background
	contArea := image.Rect(0, 0, gtx.Constraints.Max.X, height)
	contDims := DrawBox(gtx, Box{
		Size:       contArea,
		Color:      c.Bg,
		R:          theme.CornerR(0, 0, inputSpecs.shape, inputSpecs.shape),
		Clickable:  &props.Clickable,
		GeometryCb: func() { props.Inputable.Subscribe(gtx) },
	})
	gtx.Constraints.Max.Y = defaultH
	gtx.Constraints.Max.X = contDims.Size.X
	indicatorH := gtx.Dp(inputSpecs.bLineUnfocused)
	if props.IsFocused() {
		indicatorH = gtx.Dp(inputSpecs.bLineFocused)
	}
	// Indicator
	OffsetBy(gtx, image.Pt(0, contDims.Size.Y-indicatorH), func(gtx layout.Context) {
		DrawBox(gtx, Box{
			Size:  image.Rect(0, 0, gtx.Constraints.Max.X, indicatorH),
			Color: c.Indicator,
		})
	})

	var incrDims layout.Dimensions
	// Leading icon
	if props.LeadingIcon != nil {
		OffsetBy(gtx, image.Pt(outterIconPadding, contDims.Size.Y/2-iconSize/2), func(gtx layout.Context) {
			gtx.Constraints.Min.X = iconSize
			props.LeadingIcon.Layout(gtx, c.Icon)
		})
		incrDims.Size.X += iconSize + outterIconPadding
	}
	gtx.Constraints.Max = contDims.Size

	// Label
	OffsetBy(gtx, image.Pt(textXPadding+incrDims.Size.X, yPadding), func(gtx layout.Context) {
		lblTxtAlign := layout.W
		var lblTxtSize unit.Sp = inputSpecs.lblSizeBig
		var lblTxtHeight unit.Sp = inputSpecs.lblHeightBig
		if props.IsFocused() || (props.chipsPresent || len(props.GetInput()) > 0) {
			lblTxtAlign = layout.NW
			lblTxtSize = inputSpecs.lblSizeSmall
			lblTxtHeight = inputSpecs.lblHeightSmall
		}
		gtx.Constraints.Max.X -= incrDims.Size.X + textXPadding*2
		gtx.Constraints.Max.Y -= yPadding * 2
		gtx.Constraints.Min = gtx.Constraints.Max
		lblTxtAlign.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = image.Point{}
			labelTxt := material.Body2(th.Theme, props.LabelText)
			labelTxt.Font.Typeface = "Roboto"
			labelTxt.TextSize = lblTxtSize
			labelTxt.LineHeight = lblTxtHeight
			labelTxt.Font.Weight = inputSpecs.lblWeight
			labelTxt.Color = c.LabelText
			txtDims := labelTxt.Layout(gtx)
			incrDims.Size.Y += txtDims.Size.Y
			return txtDims
		})
	})
	// Content
	OffsetBy(gtx, image.Pt(textXPadding+incrDims.Size.X, yPadding+incrDims.Size.Y), func(gtx layout.Context) {
		trailingIcon := 0
		if props.TrailingIcon != nil {
			trailingIcon += iconSize + outterIconPadding
		}
		gtx.Constraints.Max.X -= incrDims.Size.X + textXPadding*2 + trailingIcon
		gtx.Constraints.Max.Y -= yPadding * 2
		gtx.Constraints.Min = gtx.Constraints.Max
		contentM.Add(gtx.Ops)
		incrDims.Size.X += contentDims.Size.X + textXPadding
	})
	// Trailing icon
	if props.TrailingIcon != nil {
		OffsetBy(gtx, image.Pt(contDims.Size.X-outterIconPadding-iconSize, contDims.Size.Y/2-iconSize/2), func(gtx layout.Context) {
			gtx.Constraints.Min.X = iconSize
			props.TrailingIcon.Layout(gtx, c.Icon)
			iconSizeHalf := iconSize / 2
			DrawBox(gtx, Box{
				Size:      image.Rect(0, 0, iconSize, iconSize),
				R:         theme.CornerR(iconSizeHalf, iconSizeHalf, iconSizeHalf, iconSizeHalf),
				Clickable: &props.Cancel,
				HideInk:   true,
			})
		})
	}
	// Hover layer
	if props.Hovered() {
		DrawBox(gtx, Box{
			Size:  contArea,
			Color: th.Palette.Input.Hovered.Bg,
		})
	}
	// Supporting text (character limit)
	if props.MaxLen > 0 {
		supTextM, supTextDims := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = image.Point{}
			txt := material.Body2(th.Theme, fmt.Sprintf("%d/%d", len(props.GetInput()), props.MaxLen))
			txt.Color = c.SupportingText
			txt.TextSize = inputSpecs.supTxtSize
			txt.LineHeight = inputSpecs.supTxtHeight
			return txt.Layout(gtx)
		})
		incrDims.Size.Y += supTextDims.Size.Y
		OffsetBy(gtx, image.Pt(contDims.Size.X-supTextPadding-supTextDims.Size.X, height+supTextTopPadding), func(gtx layout.Context) {
			supTextM.Add(gtx.Ops)
		})
		contDims.Size.Y += supTextDims.Size.Y + supTextTopPadding
	}
	return contDims
}

type InputFieldProps struct {
	Base        InputFieldBase
	MaxLen      int
	Filter      string
	Placeholder string
}

func DrawInputField(gtx layout.Context, th *theme.RepeatTheme, props InputFieldProps) layout.Dimensions {
	editorRender := func(gtx layout.Context, c theme.InputFieldPalette) layout.Dimensions {
		props.Base.Editor.MaxLen = props.MaxLen
		props.Base.Editor.SingleLine = true
		props.Base.Editor.Submit = true
		props.Base.Editor.Filter = props.Filter
		placeholder := ""
		if props.Base.IsFocused() {
			placeholder = props.Placeholder
		}
		ed := material.Editor(th.Theme, &props.Base.Editor, placeholder)
		ed.Font.Typeface = "Roboto"
		ed.Color = c.InputText
		ed.LineHeight = inputSpecs.lblHeightBig
		ed.TextSize = inputSpecs.lblSizeBig
		ed.Font.Weight = inputSpecs.lblWeight
		passOp := pointer.PassOp{}.Push(gtx.Ops)
		edDims := ed.Layout(gtx)
		passOp.Pop()
		return edDims
	}
	return drawInputFieldBase(gtx, th, inputFieldBaseProps{
		InputFieldBase: props.Base,
		content:        editorRender,
	})
}

const comboboxMaxDropdown unit.Dp = 128

type ComboboxOption struct {
	Text string
	Cl   *widget.Clickable
}

type ComboboxProps struct {
	Base        InputFieldBase
	MaxLen      int
	Placeholder string
	Chips       []string
	Dropdown    *op.CallOp
	Options     []ComboboxOption
}

func DrawCombobox(gtx layout.Context, th *theme.RepeatTheme, props ComboboxProps) layout.Dimensions {
	editorRender := func(gtx layout.Context, c theme.InputFieldPalette) layout.Dimensions {
		props.Base.Editor.MaxLen = props.MaxLen
		props.Base.Editor.Submit = true
		placeholder := ""
		if props.Base.IsFocused() {
			placeholder = props.Placeholder
		}
		ed := material.Editor(th.Theme, &props.Base.Editor, placeholder)
		ed.Font.Typeface = "Roboto"
		ed.Color = c.InputText
		ed.LineHeight = inputSpecs.lblHeightBig
		ed.TextSize = inputSpecs.lblSizeBig
		ed.Font.Weight = inputSpecs.lblWeight
		passOp := pointer.PassOp{}.Push(gtx.Ops)

		var dims layout.Dimensions
		// Chips
		offset, gap, chipHeight := 0, gtx.Dp(5), gtx.Dp(chipSpecs.height)
		if len(props.Chips) > 0 {
			for _, it := range props.Chips {
				chipM, chipDims := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
					return DrawChip(gtx, th, ChipProps{Text: it})
				})
				if offset != 0 && gtx.Constraints.Max.X-offset-chipDims.Size.X < 0 {
					offset = 0
					dims.Size.Y += chipHeight + gap
				}
				OffsetBy(gtx, image.Pt(offset, dims.Size.Y), func(gtx layout.Context) {
					chipM.Add(gtx.Ops)
					offset += chipDims.Size.X + gap
				})
			}
		}
		if offset > 0 {
			dims.Size.Y += chipHeight + gap
		}
		OffsetBy(gtx, image.Pt(0, dims.Size.Y), func(gtx layout.Context) {
			edDims := ed.Layout(gtx)
			dims.Size.X += edDims.Size.X
			dims.Size.Y += edDims.Size.Y
		})
		passOp.Pop()
		return dims
	}
	inputFieldDims := drawInputFieldBase(gtx, th, inputFieldBaseProps{
		InputFieldBase: props.Base,
		content:        editorRender,
		chipsPresent:   len(props.Chips) > 0,
	})

	// Dropdown
	if len(props.Options) > 0 {
		// This is not working when inside macro. There is no other way to center/align it without macro
		// Remove cringy AbsoluteOffset
		*props.Dropdown, _ = MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
			x0, y0 := AbsoluteOffset.X-inputFieldDims.Size.X/2, AbsoluteOffset.Y+inputFieldDims.Size.Y
			lsM, lsDims := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
				ls := &props.Base.Inputable.List
				ls.Axis = layout.Vertical
				gtx.Constraints.Min = image.Point{}
				gtx.Constraints.Max = image.Pt(inputFieldDims.Size.X, gtx.Dp(comboboxMaxDropdown))
				return material.List(th.Theme, ls).Layout(gtx, len(props.Options), func(gtx layout.Context, index int) layout.Dimensions {
					curOption := props.Options[index]
					txt := material.Body2(th.Theme, curOption.Text)
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					txt.LineHeight = inputSpecs.lblHeightBig
					txt.TextSize = inputSpecs.lblSizeBig
					txt.Alignment = text.Middle
					dims := layout.UniformInset(5).Layout(gtx, txt.Layout)
					DrawBox(gtx, Box{
						Size:      image.Rect(0, 0, dims.Size.X, dims.Size.Y),
						Clickable: curOption.Cl,
						HideInk:   true,
					})
					return dims
				})
			})
			var dims layout.Dimensions
			OffsetBy(gtx, image.Pt(x0, y0), func(gtx layout.Context) {
				shadowOff := gtx.Dp(2)
				DrawBox(gtx, Box{
					Size:  image.Rect(0, 0, shadowOff+inputFieldDims.Size.X, shadowOff+lsDims.Size.Y),
					Color: th.Palette.Backdrop,
					R:     theme.CornerR(4, 4, 0, 0),
				})
				dims = DrawBox(gtx, Box{
					Size:  image.Rect(0, 0, inputFieldDims.Size.X, lsDims.Size.Y),
					Color: th.Palette.Input.Enabled.Bg,
					R:     theme.CornerR(4, 4, 0, 0),
				})
				lsM.Add(gtx.Ops)
			})
			return dims
		})
	}
	return inputFieldDims
}

type dividerAxis int

const (
	Horizontal dividerAxis = iota
	Vertical
)

type dividerInsetType int

const (
	DividerFullWidth dividerInsetType = iota
	DividerInset
	DividerMiddleInset
)

type dividerMaterialSpecs struct {
	thickness    unit.Dp
	margin       unit.Dp
	bottomMargin unit.Dp
}

var dividerSpecs = dividerMaterialSpecs{
	thickness:    1,
	margin:       16,
	bottomMargin: 8,
}

type DividerProps struct {
	Axis  dividerAxis
	Inset dividerInsetType
}

func DrawDivider(gtx layout.Context, th *theme.RepeatTheme, props DividerProps) {
	var size image.Rectangle
	margin, thickness := gtx.Dp(dividerSpecs.margin), gtx.Dp(dividerSpecs.thickness)

	var margin0, margin1 int
	switch props.Inset {
	case DividerInset:
		margin0 = margin
	case DividerMiddleInset:
		margin0, margin1 = margin, margin
	}

	if props.Axis == Horizontal {
		size = image.Rect(margin0, 0, gtx.Constraints.Max.X-margin1, thickness)
	} else {
		size = image.Rect(0, margin0, thickness, gtx.Constraints.Max.Y-margin1)
	}
	DrawBox(gtx, Box{
		Size:  size,
		Color: th.Palette.Divider,
	})
}

type chipMaterialSpecs struct {
	outline                unit.Dp
	height                 unit.Dp
	shape                  unit.Dp
	iconSize               unit.Dp
	xPadding               unit.Dp
	iconsPadding           unit.Dp
	betweenElementsPadding unit.Dp
}

var chipSpecs = chipMaterialSpecs{
	outline:                1,
	height:                 32,
	shape:                  8,
	iconSize:               18,
	xPadding:               16,
	betweenElementsPadding: 8,
}

type ChipProps struct {
	Text     string
	Selected bool
	Cl       *widget.Clickable
}

func DrawChip(gtx layout.Context, th *theme.RepeatTheme, props ChipProps) layout.Dimensions {
	c := th.Palette.Chip.Enabled
	outlineW := chipSpecs.outline
	if props.Selected {
		c = th.Palette.Chip.Focused
		outlineW = 0
	}

	// Text (macro)
	textM, textDim := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min = image.Point{}
		txt := material.Body2(th.Theme, props.Text)
		txt.Color = c.Text
		txt.Font.Typeface = "Roboto"
		txt.LineHeight = 20
		txt.TextSize = 14
		txt.Font.Weight = 500
		txt.WrapPolicy = text.WrapWords
		return txt.Layout(gtx)
	})

	// Bg
	h := gtx.Dp(chipSpecs.height)
	shape := gtx.Dp(chipSpecs.shape)
	xPadding := gtx.Dp(chipSpecs.xPadding)
	chipSize := image.Rect(0, 0, textDim.Size.X+xPadding*2, h)
	iconSize := gtx.Dp(chipSpecs.iconSize)
	if props.Selected {
		chipSize.Max.X += iconSize
	}
	chipDims := DrawBox(gtx, Box{
		Size:      chipSize,
		Color:     c.Bg,
		R:         theme.CornerR(shape, shape, shape, shape),
		StrokeC:   c.Outline,
		StrokeW:   outlineW,
		Clickable: props.Cl,
	})

	// Icon + text
	textOff := xPadding
	if props.Selected {
		OffsetBy(gtx, image.Pt(xPadding/2, chipSize.Max.Y/2-iconSize/2), func(gtx layout.Context) {
			gtx.Constraints.Min.X = gtx.Dp(chipSpecs.iconSize)
			micons.Check.Layout(gtx, c.Text)
		})
		textOff += iconSize
	}
	OffsetBy(gtx, image.Pt(textOff, textDim.Size.Y/2), func(gtx layout.Context) {
		textM.Add(gtx.Ops)
	})
	return chipDims
}

type filterChipsSizeSpecs struct {
	xGap unit.Dp
	yGap unit.Dp
}

var filterChipsSpecs = filterChipsSizeSpecs{
	xGap: 5,
	yGap: 12,
}

type FilterChip struct {
	Text     string
	Selected bool
	Cl       *widget.Clickable
}

func DrawChipsFilter(gtx layout.Context, th *theme.RepeatTheme, chips []*FilterChip) layout.Dimensions {
	xGap, yGap := gtx.Dp(filterChipsSpecs.xGap), gtx.Dp(filterChipsSpecs.yGap)
	xOffset, yOffset := 0, 0
	for _, it := range chips {
		chipM, chipDims := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
			return DrawChip(gtx, th, ChipProps{
				Text:     it.Text,
				Selected: it.Selected,
				Cl:       it.Cl,
			})
		})
		if gtx.Constraints.Max.X-chipDims.Size.X-xOffset < 0 {
			xOffset = 0
			yOffset += chipDims.Size.Y + yGap
		}
		OffsetBy(gtx, image.Pt(xOffset, yOffset), func(gtx layout.Context) {
			chipM.Add(gtx.Ops)
			xOffset += chipDims.Size.X + xGap
		})
	}
	chipHeight := gtx.Dp(chipSpecs.height)
	return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, yOffset+chipHeight+yGap)}
}

type iconBMaterialWidth struct {
	esIcon unit.Dp
	esW    unit.Dp
	esH    unit.Dp
	sIcon  unit.Dp
	sW     unit.Dp
	sH     unit.Dp
	mIcon  unit.Dp
	mW     unit.Dp
	mH     unit.Dp
}

type iconBMaterialSpecs struct {
	standard iconBMaterialWidth
	wide     iconBMaterialWidth
}

var iconBSpecs = iconBMaterialSpecs{
	standard: iconBMaterialWidth{
		esIcon: 20,
		esW:    32,
		esH:    32,
		sIcon:  24,
		sW:     40,
		sH:     40,
		mIcon:  24,
		mW:     56,
		mH:     56,
	},
	wide: iconBMaterialWidth{
		esIcon: 20,
		esW:    40,
		esH:    32,
		sIcon:  24,
		sW:     52,
		sH:     40,
		mIcon:  24,
		mW:     72,
		mH:     56,
	},
}

type iconButtonWidth int

const (
	IconButtonStandard iconButtonWidth = iota
	IconButtonWide
)

type iconButtonSize int

const (
	IconButtomExtraSmall iconButtonSize = iota
	IconButtonSmall
	IconButtonMedium
)

type IconButtonProps struct {
	Icon  *widget.Icon
	Th    *theme.RepeatTheme
	Text  string
	Width iconButtonWidth
	Size  iconButtonSize
	Cl    *widget.Clickable
	IsOff bool
}

func DrawIconButton(gtx layout.Context, props IconButtonProps) layout.Dimensions {
	var sz iconBMaterialWidth
	switch props.Width {
	case IconButtonStandard:
		sz = iconBSpecs.standard
	case IconButtonWide:
		sz = iconBSpecs.wide
	}
	var iconSize, w, h int
	switch props.Size {
	case IconButtomExtraSmall:
		iconSize, w, h = gtx.Dp(sz.esIcon), gtx.Dp(sz.esW), gtx.Dp(sz.esH)
	case IconButtonSmall:
		iconSize, w, h = gtx.Dp(sz.sIcon), gtx.Dp(sz.sW), gtx.Dp(sz.sH)
	case IconButtonMedium:
		iconSize, w, h = gtx.Dp(sz.mIcon), gtx.Dp(sz.mW), gtx.Dp(sz.mH)
	}
	shape := h / 2
	cl := props.Cl
	color := props.Th.Palette.IconButton.Enabled
	if props.IsOff {
		cl = nil
		color = props.Th.Palette.IconButton.Disabled
	}
	boxDim := DrawBox(gtx, Box{
		Size:      image.Rect(0, 0, w, h),
		Color:     color.Bg,
		R:         theme.CornerR(shape, shape, shape, shape),
		Clickable: cl,
	})
	iconSizeHalf := iconSize / 2
	OffsetBy(gtx, image.Pt(w/2-iconSizeHalf, h/2-iconSizeHalf), func(gtx layout.Context) {
		gtx.Constraints.Min.X = iconSize
		props.Icon.Layout(gtx, color.Icon)
	})
	return boxDim
}
