package theme

var markers = markersSizing{
	Lbl: labelSizing{
		MinW:      70,
		MaxW:      150,
		H:         50,
		MargB:     "18%",
		IconW:     35,
		InvisPadE: 5,
	},
	Pole: poleSizing{
		W:          4,
		ActiveWPad: 10,
		Pad:        "40%",
		FlagW:      30,
		FlagH:      50,
		FlagCorn:   45,
		Dash:       8,
	},
}

type markersSizing struct {
	Lbl  labelSizing
	Pole poleSizing
}

type labelSizing struct {
	MinW      int
	MaxW      int
	H         int
	MargB     string
	IconW     int
	InvisPadE int // invisible padding for East to make overlapping more smooth
}

type poleSizing struct {
	Pad        string
	W          int
	ActiveWPad int // padding for width to make grabbing easier
	Dash       int
	FlagH      int
	FlagW      int
	FlagCorn   float64 // flag's notch corner
}
