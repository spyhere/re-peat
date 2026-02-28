package theme

import "image/color"

var repeatPalette = palette{
	Backdrop: argb(0xdd000000),
	Divider:  rgb(0xCAC4D0),
	Search: searchStatesPalette{
		Enabled: searchPalette{
			Bg:      rgb(0xECE6F0),
			Icon:    rgb(0x49454F),
			SupText: rgb(0x49454F),
		},
		Hovered: searchPalette{
			Bg:      argb(0x141D1B20),
			Icon:    rgb(0x49454F),
			SupText: rgb(0x49454F),
		},
		Pressed: searchPalette{
			Bg:      argb(0x191D1B20),
			Icon:    rgb(0x49454F),
			SupText: rgb(0x49454F),
		},
	},
	CardBg: rgb(0xF7F2FA),
	SegButtons: segButtonsStatesPalette{
		Enabled: segButtonsPalette{
			Outline:   rgb(0x79747E),
			Selected:  rgb(0xE8DEF8),
			SelText:   rgb(0x4A4458),
			UnSelText: rgb(0x1D1B20),
		},
		Disabled: segButtonsPalette{
			Outline:   argb(0x1f1D1B20),
			Selected:  rgb(0x5C5863),
			SelText:   rgb(0x4A4458),
			UnSelText: argb(0x611D1B20),
		},
		Hovered: segButtonsPalette{
			Selected:   argb(0x141D1B20),
			UnSelected: argb(0x141D1B20),
		},
	},
	MarkersViewBg: rgb(0x7EB6D7),
	Editor: editorPalette{
		Bg:        tan,
		SoundWave: blackRF,
		Playhead:  white,
		AddMarker: cyan,
		MarkerDev: 8,
		Grid: gridPalette{
			Tick:    rgb(0x000000),
			Tick5s:  white,
			Tick10s: white,
		},
	},
}

var (
	blackFL    = rgb(0x240030)
	blackRF    = rgb(0x010101)
	cyan       = rgb(0x71f8ff)
	darkBlue   = rgb(0x1c2143)
	mimosa     = rgb(0xf0be60)
	red        = rgb(0xff0000)
	roseQuartz = rgb(0xf1dcd9)
	tan        = rgb(0xcfb196)
	white      = rgb(0xffffff)
)

type palette struct {
	Backdrop      color.NRGBA
	Divider       color.NRGBA
	Search        searchStatesPalette
	CardBg        color.NRGBA
	SegButtons    segButtonsStatesPalette
	MarkersViewBg color.NRGBA
	Editor        editorPalette
}

type searchPalette struct {
	Bg      color.NRGBA
	Icon    color.NRGBA
	SupText color.NRGBA
}

type searchStatesPalette struct {
	Enabled searchPalette
	Pressed searchPalette
	Hovered searchPalette
}

type segButtonsStatesPalette struct {
	Enabled  segButtonsPalette
	Disabled segButtonsPalette
	Hovered  segButtonsPalette
}

type segButtonsPalette struct {
	Outline    color.NRGBA
	Selected   color.NRGBA
	UnSelected color.NRGBA
	SelText    color.NRGBA
	UnSelText  color.NRGBA
}

type editorPalette struct {
	SoundWave color.NRGBA
	Bg        color.NRGBA
	Playhead  color.NRGBA
	Grid      gridPalette
	AddMarker color.NRGBA
	MarkerDev int // Color deviation for stacked markers, so they can be distinguished
}

type gridPalette struct {
	Tick    color.NRGBA
	Tick10s color.NRGBA
	Tick5s  color.NRGBA
}

// **
// Fallen Leaves
// mimosa
// blackFL
// white

// **
// Raw&Fresh
// roseQuartz
// blackRF
// red

// **
// Raw&Fresh
// tab
// blackRF
// white

// *?
// Stormy weather
// darkBlue
// cyan
// red

// *
// Stormy weather
// darkBlue
// mimosa
// white

// Saw this trick in github.com/chapar-rest/chapar
func rgb(c uint32) color.NRGBA {
	return argb(0xff000000 | c)
}

func argb(c uint32) color.NRGBA {
	return color.NRGBA{A: uint8(c >> 24), R: uint8(c >> 16), G: uint8(c >> 8), B: uint8(c)}
}
