package theme

import (
	"gioui.org/font/gofont"
	"gioui.org/text"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/fonts"
)

type RepeatTheme struct {
	*material.Theme

	Palette palette
	Sizing  sizing
}

func New() (*RepeatTheme, error) {
	fontFaces, err := fonts.LoadFonts(gofont.Collection())
	if err != nil {
		return nil, err
	}
	newTheme := &RepeatTheme{
		Theme:   material.NewTheme(),
		Palette: repeatPalette,
		Sizing:  repeatSizing,
	}
	newTheme.Shaper = text.NewShaper(text.WithCollection(fontFaces))
	return newTheme, nil
}
