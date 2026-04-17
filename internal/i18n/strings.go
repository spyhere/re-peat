package i18n

type Strings struct {
	Common  Common
	Generic Generic
	Markers MarkersView
	Project ProjectView
}

type Generic struct {
	Amount        string
	Audio         string
	AudioChannels string
	Cancel        string
	Editor        string
	Length        string
	Markers       string
	Modified      string
	Mono          string
	Name          string
	Notes         string
	Ok            string
	Project       string
	SampleRate    string
	Save          string
	SaveAs        string
	Size          string
	Stereo        string
	Tags          string
	Time          string
	WithComments  string
}

type Common struct {
	LoadingFile string
}

type ProjectView struct {
	MConflictLoadBody  string
	MConflictLoadTitle string
}

type MarkersView struct {
	MCreate            string
	MDeleteALlBody     string
	MDeleteALlTitle    string
	MEdit              string
	MNamePlaceholder   string
	MNote              string
	MNotePlaceholder   string
	NoMatches          string
	SearchBPlaceholder string
	TagsFilter         string
}
