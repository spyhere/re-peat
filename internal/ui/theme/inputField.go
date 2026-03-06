package theme

import "image/color"

type inputFieldStatePalette struct {
	Enabled  InputFieldPalette
	Disabled InputFieldPalette
	Hovered  InputFieldPalette
	Focused  InputFieldPalette
}
type InputFieldPalette struct {
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
	Enabled: InputFieldPalette{
		Bg:             rgb(0xE6E0E9),
		LabelText:      rgb(0x49454F),
		LabelTextEmpty: rgb(0x049454F),
		Icon:           rgb(0x49454F),
		Indicator:      rgb(0x49454F),
		SupportingText: rgb(0x49454F),
		InputText:      rgb(0x1D1B20),
		Caret:          rgb(0x6750A4),
	},
	Hovered: InputFieldPalette{
		Bg: argb(0x141D1B20),
	},
}
