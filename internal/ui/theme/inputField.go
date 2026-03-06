package theme

import "image/color"

type inputFieldStatePalette struct {
	Enabled  inputFieldPalette
	Disabled inputFieldPalette
	Hovered  inputFieldPalette
	Focused  inputFieldPalette
}
type inputFieldPalette struct {
	Bg             color.NRGBA
	LabelText      color.NRGBA
	LabelTextEmpty color.NRGBA
	Icon           color.NRGBA
	Indicator      color.NRGBA
	SupportingText color.NRGBA
	InputText      color.NRGBA
	Caret          color.NRGBA
}

var inputFieldP = inputFieldStatePalette{
	Enabled: inputFieldPalette{
		Bg:             rgb(0xE6E0E9),
		LabelText:      rgb(0x49454F),
		LabelTextEmpty: rgb(0x049454F),
		Icon:           rgb(0x49454F),
		Indicator:      rgb(0x49454F),
		SupportingText: rgb(0x49454F),
		InputText:      rgb(0x1D1B20),
		Caret:          rgb(0x6750A4),
	},
	Hovered: inputFieldPalette{
		Bg: argb(0x141D1B20),
	},
}
