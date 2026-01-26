package theme

var repeatSizing = sizing{
	Editor: editorSizing{
		PlayheadW: 2,
		WaveM:     "32%",
		Grid: gridSizing{
			Tick:            10,
			Tick5s:          20,
			Tick10s:         30,
			MinTimeInterval: 100,
		},
		Markers: markers,
	},
}

type sizing struct {
	Editor editorSizing
}
type editorSizing struct {
	PlayheadW int
	WaveM     string
	Grid      gridSizing
	Markers   markersSizing
}

// In px
type gridSizing struct {
	MinTimeInterval int
	Tick            int
	Tick10s         int
	Tick5s          int
}
