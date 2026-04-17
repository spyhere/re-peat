package i18n

var enStr = Strings{
	Common: Common{
		LoadingFile: "Loading file...",
	},
	Generic: Generic{
		Amount:        "Amount",
		Audio:         "Audio",
		AudioChannels: "Audio Channels",
		Cancel:        "Cancel",
		Editor:        "Editor",
		Length:        "Length",
		Markers:       "Markers",
		Modified:      "Modified",
		Name:          "Name",
		Notes:         "Notes",
		Ok:            "OK",
		Project:       "Project",
		SampleRate:    "Sample Rate",
		Save:          "Save",
		SaveAs:        "Save As",
		Size:          "Size",
		Tags:          "Tags",
		Time:          "Time",
		WithComments:  "With comments",
	},
	Markers: MarkersView{
		MCreate:            "Create marker",
		MDeleteALl:         "Delete all markers",
		MEdit:              "Edit marker",
		MNamePlaceholder:   "marker's name...",
		MNote:              "Notes",
		SearchBPlaceholder: "search by name...",
	},
	Project: ProjectView{
		MConflictLoadBody:  "These markers were initially saved for \"%s\", but currently loaded \"%s\".\nStill want to load them for this audio file?\n\nMarkers exceeding audio length will be set to 0 and have \"Redacted\" tag added.",
		MConflictLoadTitle: "Markers loading conflict",
	},
}
