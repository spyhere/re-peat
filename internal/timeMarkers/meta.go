package timemarkers

import (
	"strconv"
)

func NewMarkersMeta(markers TimeMarkers) MarkersMeta {
	withComments := 0
	for _, it := range markers {
		if it.Notes != "" {
			withComments++
		}
	}
	return MarkersMeta{
		init:         true,
		Amount:       len(markers),
		WithComments: withComments,
	}
}

type MarkersMeta struct {
	init         bool
	Amount       int
	WithComments int
}

func (m MarkersMeta) AmountString() string {
	if !m.init {
		return ""
	}
	return strconv.Itoa(m.Amount)
}

func (m MarkersMeta) WithCommentsString() string {
	if !m.init {
		return ""
	}
	return strconv.Itoa(m.WithComments)
}
