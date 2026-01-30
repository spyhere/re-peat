package theme

var repeatSizing = sizing{
	Editor: editorSizing{
		PlayheadW: 4,
		WaveM:     "32%",
		Grid: gridSizing{
			MargT:           "-15%",
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
	MargT           string
	MinTimeInterval int
	TickW           int
	TickH           int
	Tick10s         int
	Tick5s          int
}
