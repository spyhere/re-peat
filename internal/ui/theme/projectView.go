package theme

import "image/color"

type projectPalette struct {
	Bg           color.NRGBA
	CargBg       color.NRGBA
	CardStroke   color.NRGBA
	LoadButtonBg color.NRGBA
	SaveButtonBg color.NRGBA
	SaveButtonFg color.NRGBA
}

var project = projectPalette{
	Bg:           rgb(0xE0E0E0),
	CargBg:       rgb(0xF2F2F2),
	CardStroke:   rgb(0x6750A4),
	LoadButtonBg: rgb(0x4F378A),
	SaveButtonBg: rgb(0xD0BCFF),
	SaveButtonFg: rgb(0x4A4459),
}
