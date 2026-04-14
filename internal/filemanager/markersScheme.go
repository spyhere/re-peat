package filemanager

import timemarkers "github.com/spyhere/re-peat/internal/timeMarkers"

type MarkersSaveScheme struct {
	Version int
	FName   string
	FSize   int64
	FLen    float64
	FSRate  int
	Markers timemarkers.TimeMarkers
}
