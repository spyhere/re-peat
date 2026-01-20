package theme

import "image/color"

var repeatPalette = palette{
	Editor: editorPalette{
		Bg:        tan,
		SoundWave: blackRF,
		Playhead:  white,
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
	Editor editorPalette
}

type editorPalette struct {
	SoundWave color.NRGBA
	Bg        color.NRGBA
	Playhead  color.NRGBA
	Grid      gridPalette
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
