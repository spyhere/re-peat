package theme

import (
	"gioui.org/widget/material"
)

type RepeatTheme struct {
	*material.Theme

	Palette palette
	Sizing  sizing
}

func New() *RepeatTheme {
	return &RepeatTheme{
		Theme:   material.NewTheme(),
		Palette: repeatPalette,
		Sizing:  repeatSizing,
	}
}
