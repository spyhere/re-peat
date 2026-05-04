package theme

var repeatSizing = sizing{
	SegButtonsTopM: 30,
	Editor: editorSizing{
		PlayheadW:    4,
		CreateButtMT: 85.0,
		WaveM:        32.0,
		Grid: gridSizing{
			MargT:           85.0,
			TickW:           2,
			TickH:           10,
			Tick5s:          20,
			Tick10s:         30,
			MinTimeInterval: 100,
		},
		Markers: markers,
	},
}

type sizing struct {
	SegButtonsTopM int
	Editor         editorSizing
}

type editorSizing struct {
	PlayheadW    int
	CreateButtMT float32 // create button margin top
	WaveM        float32
	Grid         gridSizing
	Markers      markersSizing
}

// In px
type gridSizing struct {
	MargT           float32
	MinTimeInterval int
	TickW           int
	TickH           int
	Tick10s         int
	Tick5s          int
}

type CornerRadii struct {
	SE, SW, NW, NE int
}

func CornerR(se, sw, nw, ne int) CornerRadii {
	return CornerRadii{SE: se, SW: sw, NW: nw, NE: ne}
}
