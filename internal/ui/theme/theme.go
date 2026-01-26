package theme

import (
	"gioui.org/widget/material"
)

type RepeatTheme struct {
	*material.Theme

	Palette palette
	// TODO: Move wave padding and playhead width to Sizing
	Sizing sizing
}

func New() *RepeatTheme {
	return &RepeatTheme{
		Theme:   material.NewTheme(),
		Palette: repeatPalette,
		Sizing:  repeatSizing,
	}
}
