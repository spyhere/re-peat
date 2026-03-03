package common

import (
	"image"
	"image/color"
	"math"

	"gioui.org/font"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
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
		for _, it := range b.Clickable.History() {
			gtx.Constraints.Min = rrect.Rect.Max
			drawInk(gtx, it)
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
	*Searchable
}

func DrawSearch(gtx layout.Context, th *theme.RepeatTheme, props SProps) layout.Dimensions {
	props.Searchable.Update(gtx)

	containerH := gtx.Dp(searchSpecs.height)
	containerHHalft := containerH / 2
	containerW := gtx.Dp(searchSpecs.minWidth)
	bDims := DrawBox(gtx, Box{
		Size:       image.Rect(0, 0, containerW, containerH),
		Color:      th.Palette.Search.Enabled.Bg,
		R:          theme.CornerR(containerHHalft, containerHHalft, containerHHalft, containerHHalft),
		Clickable:  &props.Clickable,
		GeometryCb: func() { props.Searchable.Subscribe(gtx) },
	})
	if props.IsHovered() {
		DrawBox(gtx, Box{
			Size:  image.Rect(0, 0, containerW, containerH),
			Color: th.Palette.Search.Hovered.Bg,
			R:     theme.CornerR(containerHHalft, containerHHalft, containerHHalft, containerHHalft),
		})
	} else if props.IsFocused() {
		DrawBox(gtx, Box{
			Size:  image.Rect(0, 0, containerW, containerH),
			Color: th.Palette.Search.Pressed.Bg,
			R:     theme.CornerR(containerHHalft, containerHHalft, containerHHalft, containerHHalft),
		})
	}

	xPadd := gtx.Dp(searchSpecs.xPadding)

	iconSz := gtx.Dp(searchSpecs.iconSize)
	iconPadding := gtx.Dp(searchSpecs.iconXPadding)

	textH := gtx.Sp(searchSpecs.fontLineHeight)
	OffsetBy(gtx, image.Pt(xPadd*2, bDims.Size.Y-textH-textH/2), func() {
		gtx.Constraints.Max = image.Pt(bDims.Size.X-xPadd*2-iconSz-iconPadding*2, textH)
		if props.isFocused {
			props.Editor.SingleLine = true
			ed := material.Editor(th.Theme, &props.Editor, "")
			ed.Font.Typeface = "Roboto"
			ed.Color = th.Palette.Search.Enabled.SupText
			ed.LineHeight = searchSpecs.fontLineHeight
			ed.TextSize = searchSpecs.fontSize
			ed.Font.Weight = 400
			passOp := pointer.PassOp{}.Push(gtx.Ops)
			ed.Layout(gtx)
			passOp.Pop()
			gtx.Execute(key.FocusCmd{Tag: &props.Editor})
		} else {
			text := props.GetInput()
			if text == "" {
				text = props.DefaultText
			}
			txt := material.Body2(th.Theme, text)
			txt.Font.Typeface = "Roboto"
			txt.Color = th.Palette.Search.Enabled.SupText
			txt.LineHeight = searchSpecs.fontLineHeight
			txt.TextSize = searchSpecs.fontSize
			txt.Font.Weight = 400
			txt.Layout(gtx)
		}
	})

	OffsetBy(gtx, image.Pt(bDims.Size.X-iconPadding-xPadd-iconSz/2, bDims.Size.Y/2-iconSz/2), func() {
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
	Text string
}

func DrawChip(gtx layout.Context, th *theme.RepeatTheme, props ChipProps) layout.Dimensions {
	var textDim layout.Dimensions
	textM := MakeMacro(gtx.Ops, func() {
		txt := material.Body2(th.Theme, props.Text)
		txt.Color = th.Palette.Chip.Enabled.Text
		txt.Font.Typeface = "Roboto"
		txt.LineHeight = 20
		txt.TextSize = 14
		txt.Font.Weight = 500
		textDim = txt.Layout(gtx)
	})
	h := gtx.Dp(chipSpecs.height)
	shape := gtx.Dp(chipSpecs.shape)
	xPadding := gtx.Dp(chipSpecs.xPadding)
	chipSize := image.Rect(0, 0, textDim.Size.X+xPadding*2, h)
	chipDims := DrawBox(gtx, Box{
		Size:    chipSize,
		R:       theme.CornerR(shape, shape, shape, shape),
		StrokeC: th.Palette.Chip.Enabled.Outline,
		StrokeW: chipSpecs.outline,
	})
	OffsetBy(gtx, image.Pt(xPadding, textDim.Size.Y/2), func() {
		textM.Add(gtx.Ops)
	})
	return chipDims
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
	Text  string
	Width iconButtonWidth
	Size  iconButtonSize
	Bg    color.NRGBA
	Fg    color.NRGBA
	Cl    *widget.Clickable
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
	boxDim := DrawBox(gtx, Box{
		Size:      image.Rect(0, 0, w, h),
		Color:     props.Bg,
		R:         theme.CornerR(shape, shape, shape, shape),
		Clickable: props.Cl,
	})
	iconSizeHalf := iconSize / 2
	OffsetBy(gtx, image.Pt(w/2-iconSizeHalf, h/2-iconSizeHalf), func() {
		gtx.Constraints.Min.X = iconSize
		props.Icon.Layout(gtx, props.Fg)
	})
	return boxDim
}
