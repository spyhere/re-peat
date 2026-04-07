package common

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/fonts"
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

	// TODO: Fix stroke visual glitch
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
	Disabled    bool
	DefaultText string
	*Inputable
}

func DrawSearch(gtx layout.Context, th *theme.RepeatTheme, props SProps) layout.Dimensions {
	c := th.Palette.Search.Enabled
	if props.Disabled {
		c = th.Palette.Search.Disabled
		gtx = gtx.Disabled()
	}
	props.Inputable.Update(gtx)

	containerH := gtx.Dp(searchSpecs.height)
	containerHHalft := containerH / 2
	containerW := gtx.Dp(searchSpecs.minWidth)
	// Background
	bDims := DrawBox(gtx, Box{
		Size:       image.Rect(0, 0, containerW, containerH),
		Color:      c.Bg,
		R:          theme.CornerR(containerHHalft, containerHHalft, containerHHalft, containerHHalft),
		GeometryCb: func() { props.Inputable.Subscribe(gtx) },
	})
	// Hovered layer
	isFocused := props.IsFocused(gtx)
	if props.IsHovered() {
		DrawBox(gtx, Box{
			Size:  image.Rect(0, 0, containerW, containerH),
			Color: th.Palette.Search.Hovered.Bg,
			R:     theme.CornerR(containerHHalft, containerHHalft, containerHHalft, containerHHalft),
		})
	} else if isFocused {
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
	if isFocused {
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
		gtx.Constraints.Min = image.Point{}
		ed := material.Editor(th.Theme, &props.Editor, text)
		ed.Font = fonts.GoMedium(font.Medium, font.Regular)
		ed.Color = c.Text
		ed.HintColor = th.Palette.Search.Enabled.Text
		ed.LineHeight = searchSpecs.fontLineHeight
		ed.TextSize = searchSpecs.fontSize
		if !props.Disabled {
			passOp := pointer.PassOp{}.Push(gtx.Ops)
			ed.Layout(gtx)
			passOp.Pop()
		}
	})

	OffsetBy(gtx, image.Pt(bDims.Size.X-iconPadding-xPadd-iconSz/2, bDims.Size.Y/2-iconSz/2), func(gtx layout.Context) {
		gtx.Constraints.Min.X = iconSz
		if len(props.GetInput()) > 0 {
			micons.Cancel.Layout(gtx, c.Icon)
			DrawBox(gtx, Box{
				Size:      image.Rect(0, 0, iconSz, iconSz),
				Clickable: &props.Cancel,
			})
		} else {
			micons.Search.Layout(gtx, c.Icon)
		}
	})
	return layout.Dimensions{Size: image.Pt(containerW, containerH)}
}

type inputFieldMaterialStyle struct {
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

func inputFieldBase() inputFieldMaterialStyle {
	return inputFieldMaterialStyle{
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
		lblWeight:         font.Medium,
		supTxtSize:        12,
		supTxtHeight:      16,
	}
}

type InputFieldBase struct {
	LabelText    string
	LeadingIcon  *widget.Icon
	TrailingIcon *widget.Icon
}

type inputFieldBaseProps struct {
	InputFieldBase
	*Inputable
	content      func(gtx layout.Context, c theme.InputFieldPalette) layout.Dimensions
	chipsPresent bool
}

func (s inputFieldMaterialStyle) layout(gtx layout.Context, th *theme.RepeatTheme, props inputFieldBaseProps) layout.Dimensions {
	props.Inputable.Update(gtx)

	yPadding, defaultH := gtx.Dp(s.yPadding), gtx.Dp(s.defaultH)
	outterIconPadding, textXPadding := gtx.Dp(s.outterIconPadding), gtx.Dp(s.textXPadding)
	supTextPadding, supTextTopPadding := gtx.Dp(s.supTextPadding), gtx.Dp(s.supTextTopPadding)

	iconSize := gtx.Dp(s.icon)
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
	var defaultContentH unit.Sp = s.lblHeightBig
	contentH := max(gtx.Sp(defaultContentH), contentDims.Size.Y)
	height := max(defaultH, textXPadding*2+contentH)

	// Background
	contArea := image.Rect(0, 0, gtx.Constraints.Max.X, height)
	contDims := DrawBox(gtx, Box{
		Size:       contArea,
		Color:      c.Bg,
		R:          theme.CornerR(0, 0, s.shape, s.shape),
		GeometryCb: func() { props.Inputable.Subscribe(gtx) },
	})
	gtx.Constraints.Max.Y = defaultH
	gtx.Constraints.Max.X = contDims.Size.X
	indicatorH := gtx.Dp(s.bLineUnfocused)
	isFocused := props.IsFocused(gtx)
	if isFocused {
		indicatorH = gtx.Dp(s.bLineFocused)
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
		var lblTxtSize unit.Sp = s.lblSizeBig
		var lblTxtHeight unit.Sp = s.lblHeightBig
		if isFocused || (props.chipsPresent || len(props.GetInput()) > 0) {
			lblTxtAlign = layout.NW
			lblTxtSize = s.lblSizeSmall
			lblTxtHeight = s.lblHeightSmall
		}
		gtx.Constraints.Max.X -= incrDims.Size.X + textXPadding*2
		gtx.Constraints.Max.Y -= yPadding * 2
		gtx.Constraints.Min = gtx.Constraints.Max
		lblTxtAlign.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = image.Point{}
			labelTxt := material.Body2(th.Theme, props.LabelText)
			labelTxt.Font = fonts.GoMedium(s.lblWeight, font.Regular)
			labelTxt.TextSize = lblTxtSize
			labelTxt.LineHeight = lblTxtHeight
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
	if props.IsHovered() {
		DrawBox(gtx, Box{
			Size:  contArea,
			Color: th.Palette.Input.Hovered.Bg,
		})
	}
	// Supporting text (character limit)
	if props.MaxLen > 0 {
		supTextM, supTextDims := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = image.Point{}
			txt := material.Body2(th.Theme, fmt.Sprintf("%d/%d", strlen(props.GetInput()), props.MaxLen))
			txt.Color = c.SupportingText
			txt.TextSize = s.supTxtSize
			txt.LineHeight = s.supTxtHeight
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
	Base InputFieldBase
	*Inputable
	MaxLen      int
	Filter      string
	Placeholder string
}

func DrawInputField(gtx layout.Context, th *theme.RepeatTheme, props InputFieldProps) layout.Dimensions {
	editorRender := func(gtx layout.Context, c theme.InputFieldPalette) layout.Dimensions {
		props.Editor.MaxLen = props.MaxLen
		props.Editor.SingleLine = true
		props.Editor.Submit = true
		props.Editor.Filter = props.Filter
		placeholder := ""
		if props.IsFocused(gtx) {
			placeholder = props.Placeholder
		}
		ed := material.Editor(th.Theme, &props.Editor, placeholder)
		ed.Color = c.InputText
		inputFieldSpecs := inputFieldBase()
		ed.LineHeight = inputFieldSpecs.lblHeightBig
		ed.TextSize = inputFieldSpecs.lblSizeBig
		ed.Font = fonts.GoMedium(inputFieldSpecs.lblWeight, font.Regular)
		passOp := pointer.PassOp{}.Push(gtx.Ops)
		edDims := ed.Layout(gtx)
		passOp.Pop()
		return edDims
	}
	return inputFieldBase().layout(gtx, th, inputFieldBaseProps{
		InputFieldBase: props.Base,
		Inputable:      props.Inputable,
		content:        editorRender,
	})
}

type TextFieldProps struct {
	Base InputFieldBase
	*Inputable
	MaxLen      int
	Filter      string
	Placeholder string
}

func DrawTextField(gtx layout.Context, th *theme.RepeatTheme, props TextFieldProps) layout.Dimensions {
	inputFieldStyle := inputFieldBase()
	editorRender := func(gtx layout.Context, c theme.InputFieldPalette) layout.Dimensions {
		props.Editor.MaxLen = props.MaxLen
		placeholder := ""
		if props.IsFocused(gtx) {
			placeholder = props.Placeholder
		}
		ed := material.Editor(th.Theme, &props.Editor, placeholder)
		ed.Font = fonts.GoMedium(inputFieldStyle.lblWeight, font.Regular)
		ed.Color = c.InputText
		ed.LineHeight = inputFieldStyle.lblHeightBig
		ed.TextSize = inputFieldStyle.lblSizeBig
		passOp := pointer.PassOp{}.Push(gtx.Ops)
		edDims := ed.Layout(gtx)
		passOp.Pop()
		return edDims
	}
	inputFieldStyle.defaultH *= 4
	return inputFieldStyle.layout(gtx, th, inputFieldBaseProps{
		InputFieldBase: props.Base,
		Inputable:      props.Inputable,
		content:        editorRender,
	})
}

const comboboxMaxDropdown unit.Dp = 128

type ComboboxProps struct {
	Base InputFieldBase
	*Comboboxable
	MaxLen      int
	Placeholder string
	Chips       []string
	OptionsF    func() []string // This is a function since editor events are happening before rendering dropdown
}

func DrawCombobox(gtx layout.Context, th *theme.RepeatTheme, props ComboboxProps) layout.Dimensions {
	isFocused := props.IsFocused(gtx)
	inputFieldStyle := inputFieldBase()
	editorRender := func(gtx layout.Context, c theme.InputFieldPalette) layout.Dimensions {
		props.Editor.MaxLen = props.MaxLen
		props.Editor.Submit = true
		placeholder := ""
		if isFocused {
			placeholder = props.Placeholder
		}
		ed := material.Editor(th.Theme, &props.Editor, placeholder)
		ed.Color = c.InputText
		ed.LineHeight = inputFieldStyle.lblHeightBig
		ed.TextSize = inputFieldStyle.lblSizeBig
		ed.Font.Weight = inputFieldStyle.lblWeight
		passOp := pointer.PassOp{}.Push(gtx.Ops)

		var dims layout.Dimensions
		// Chips
		offset, gap, chipHeight := 0, gtx.Dp(5), gtx.Dp(chipSpecs.height)
		if len(props.Chips) > 0 {
			props.Comboboxable.SetChips(props.Chips)
			for idx := range props.Chips {
				props.Comboboxable.HandleChipEvents(gtx, idx)
				comboChip := &props.Comboboxable.chips[idx]
				chipM, chipDims := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
					return DrawChip(gtx, th, ChipProps{Text: comboChip.Text, CloseTag: &comboChip.Tag})
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
	inputFieldDims := inputFieldStyle.layout(gtx, th, inputFieldBaseProps{
		InputFieldBase: props.Base,
		Inputable:      &props.Inputable,
		content:        editorRender,
		chipsPresent:   len(props.Chips) > 0,
	})

	// Dropdown
	options := props.OptionsF()
	if isFocused && len(options) > 0 {
		props.Comboboxable.SetOptions(options)
		inputFieldStyle := inputFieldBase()
		dropdown, _ := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
			lsM, lsDims := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
				ls := &props.optionsLs
				ls.Axis = layout.Vertical
				gtx.Constraints.Min = image.Point{}
				gtx.Constraints.Max = image.Pt(inputFieldDims.Size.X, gtx.Dp(comboboxMaxDropdown))
				list := material.List(th.Theme, ls)
				list.AnchorStrategy = material.Overlay
				return list.Layout(gtx, len(options), func(gtx layout.Context, index int) layout.Dimensions {
					props.Comboboxable.HandleOptionEvents(gtx, index)
					curOption := &props.Comboboxable.options[index]
					txt := material.Body2(th.Theme, curOption.Text)
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					txt.LineHeight = inputFieldStyle.lblHeightBig
					txt.TextSize = inputFieldStyle.lblSizeBig
					txt.Alignment = text.Middle
					var bgC color.NRGBA
					if curOption.Cl.Hovered() {
						bgC = th.Palette.ComboOption.Hovered.Bg
						txt.Color = th.Palette.ComboOption.Hovered.Fg
					}
					txtM, txtDims := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.UniformInset(5).Layout(gtx, txt.Layout)
					})
					DrawBox(gtx, Box{
						Size:      image.Rect(0, 0, txtDims.Size.X, txtDims.Size.Y),
						Clickable: &curOption.Cl,
						Color:     bgC,
						HideInk:   true,
					})
					props.Inputable.Subscribe(gtx)
					txtM.Add(gtx.Ops)
					return txtDims
				})
			})

			var dims layout.Dimensions
			OffsetBy(gtx, image.Pt(0, inputFieldDims.Size.Y), func(gtx layout.Context) {
				shadowOff := gtx.Dp(2)
				shape := inputFieldStyle.shape
				DrawBox(gtx, Box{
					Size:  image.Rect(0, 0, shadowOff+inputFieldDims.Size.X, shadowOff+lsDims.Size.Y),
					Color: th.Palette.Backdrop,
					R:     theme.CornerR(shape, shape, 0, 0),
				})
				bgC := th.Palette.Input.Enabled.Bg
				bgC.A = 0x77
				dims = DrawBox(gtx, Box{
					Size:  image.Rect(0, 0, inputFieldDims.Size.X, lsDims.Size.Y),
					Color: bgC,
					R:     theme.CornerR(shape, shape, 0, 0),
				})
				lsM.Add(gtx.Ops)
			})
			return dims
		})
		op.Defer(gtx.Ops, dropdown)
	} else {
		props.Comboboxable.ResetOptionScroll()
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
	HideIcon bool
	Cl       *widget.Clickable
	CloseTag event.Tag
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
		txt.Font = fonts.GoMedium(font.Medium, font.Regular)
		txt.LineHeight = 20
		txt.TextSize = 14
		txt.WrapPolicy = text.WrapWords
		return txt.Layout(gtx)
	})

	// Bg
	h := gtx.Dp(chipSpecs.height)
	shape := gtx.Dp(chipSpecs.shape)
	xPadding := gtx.Dp(chipSpecs.xPadding)
	chipSize := image.Rect(0, 0, textDim.Size.X+xPadding*2, h)
	iconSize := gtx.Dp(chipSpecs.iconSize)
	inBetweenPad := gtx.Dp(chipSpecs.betweenElementsPadding)

	shouldShowCheckIcon := props.Selected && !props.HideIcon
	if shouldShowCheckIcon {
		chipSize.Max.X += iconSize
	}
	shouldShowCloseIcon := props.CloseTag != nil && !props.HideIcon
	if shouldShowCloseIcon {
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
	if shouldShowCheckIcon {
		OffsetBy(gtx, image.Pt(inBetweenPad, chipSize.Max.Y/2-iconSize/2), func(gtx layout.Context) {
			gtx.Constraints.Min.X = iconSize
			micons.Check.Layout(gtx, c.Text)
		})
		textOff += iconSize
	}
	OffsetBy(gtx, image.Pt(textOff, textDim.Size.Y/2), func(gtx layout.Context) {
		textM.Add(gtx.Ops)
	})
	if shouldShowCloseIcon {
		OffsetBy(gtx, image.Pt(chipSize.Max.X-inBetweenPad-iconSize, chipSize.Max.Y/2-iconSize/2), func(gtx layout.Context) {
			gtx.Constraints.Min.X = iconSize
			micons.Close.Layout(gtx, th.Palette.Backdrop)
			RegisterTag(gtx, props.CloseTag, image.Rect(0, 0, iconSize, iconSize))
		})
	}
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
	Cl       widget.Clickable
}

func DrawChipsFilter(gtx layout.Context, th *theme.RepeatTheme, chips []FilterChip) layout.Dimensions {
	xGap, yGap := gtx.Dp(filterChipsSpecs.xGap), gtx.Dp(filterChipsSpecs.yGap)
	xOffset, yOffset := 0, 0
	for idx := range chips {
		it := &chips[idx]
		chipM, chipDims := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
			return DrawChip(gtx, th, ChipProps{
				Text:     it.Text,
				Selected: it.Selected,
				Cl:       &it.Cl,
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

type buttonMaterialSpecs struct {
	mH           unit.Dp
	mPad         unit.Dp
	mIconTextPad unit.Dp
	mShape       int
	mIcon        unit.Dp
}

type ButtonStyle struct {
	th        *theme.RepeatTheme
	Bg        color.NRGBA
	Fg        color.NRGBA
	Cl        *widget.Clickable
	Icon      *widget.Icon
	Text      string
	WExpanded bool
	buttonMaterialSpecs
}

func Button(th *theme.RepeatTheme, cl *widget.Clickable, icon *widget.Icon, text string) ButtonStyle {
	return ButtonStyle{
		th:   th,
		Bg:   th.Palette.IconButton.Enabled.Bg,
		Fg:   th.Palette.IconButton.Enabled.Icon,
		Cl:   cl,
		Icon: icon,
		Text: text,
		buttonMaterialSpecs: buttonMaterialSpecs{
			mH:           56,
			mPad:         24,
			mIconTextPad: 8,
			mShape:       16,
			mIcon:        18,
		},
	}
}

func (b ButtonStyle) Layout(gtx layout.Context) layout.Dimensions {
	textM, dims := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
		txtStyle := material.Label(b.th.Theme, 14, b.Text)
		txtStyle.LineHeight = 20
		txtStyle.Color = b.Fg
		txtStyle.Font = fonts.GoMedium(font.Medium, font.Regular)
		gtx.Constraints.Min = image.Point{}
		return txtStyle.Layout(gtx)
	})
	mPad, mIcon, mIconTextPad := gtx.Dp(b.mPad), gtx.Dp(b.mIcon), gtx.Dp(b.mIconTextPad)
	mH := gtx.Dp(b.mH)
	dims.Size.X += mPad * 2

	if b.Icon != nil {
		dims.Size.X += mIcon + mIconTextPad
	}

	btnDims := layout.Dimensions{Size: image.Pt(dims.Size.X, mH)}
	if b.WExpanded {
		btnDims.Size.X = gtx.Constraints.Min.X
		mPad += (btnDims.Size.X - dims.Size.X) / 2
	}

	DrawBox(gtx, Box{
		Size:      image.Rect(0, 0, btnDims.Size.X, btnDims.Size.Y),
		Color:     b.Bg,
		Clickable: b.Cl,
		R:         theme.CornerR(b.mShape, b.mShape, b.mShape, b.mShape),
	})
	OffsetBy(gtx, image.Pt(mPad, 0), func(gtx layout.Context) {
		var xOffset layout.Dimensions
		if b.Icon != nil {
			OffsetBy(gtx, image.Pt(0, mH/2-mIcon/2), func(gtx layout.Context) {
				gtx.Constraints.Min.X = mIcon
				b.Icon.Layout(gtx, b.Fg)
			})
			xOffset.Size.X += mIcon + mIconTextPad
		}
		OffsetBy(gtx, image.Pt(xOffset.Size.X, mH/2-dims.Size.Y/2), func(gtx layout.Context) {
			textM.Add(gtx.Ops)
		})
	})
	return btnDims
}
