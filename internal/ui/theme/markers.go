package theme

var markers = markersSizing{
	Lbl: labelSizing{
		MinW:      30,
		MaxW:      150,
		H:         50,
		Margin:    10,
		OffsetY:   "18%",
		IconW:     35,
		InvisPad:  5,
		MaxGlyphs: 12,
		CRound:    CornerR(10, 0, 0, 10),
	},
	Pole: poleSizing{
		W:          2,
		ActiveWPad: 10,
		Pad:        "40%",
		FlagW:      30,
		FlagH:      50,
		FlagCorn:   45,
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
	Margin    int
	OffsetY   string
	IconW     int
	InvisPad  int // invisible padding primarily for East and North to make overlapping more smooth
	MaxGlyphs int
	CRound    CornerRadii
}

type poleSizing struct {
	Pad        string
	W          int
	ActiveWPad int // padding for width to make grabbing easier
	FlagH      int
	FlagW      int
	FlagCorn   float64 // flag's notch corner
}
