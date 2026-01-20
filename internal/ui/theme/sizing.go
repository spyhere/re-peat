package theme

var repeatSizing = sizing{
	Editor: editorSizing{
		Grid: gridSizing{
			Tick:            10,
			Tick5s:          20,
			Tick10s:         30,
			MinTimeInterval: 100,
		},
	},
}

type sizing struct {
	Editor editorSizing
}
type editorSizing struct {
	Grid gridSizing
}

// In px
type gridSizing struct {
	MinTimeInterval int
	Tick            int
	Tick10s         int
	Tick5s          int
}
