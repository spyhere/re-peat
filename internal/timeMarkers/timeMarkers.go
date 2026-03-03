package timemarkers

import (
	"slices"

	"gioui.org/widget"
)

const (
	Limit     = 100
	TagsLimit = 10
)

type TimeMarkers []*TimeMarker

func NewTimeMarkers() TimeMarkers {
	return make([]*TimeMarker, 0, Limit)
}

type TimeMarker struct {
	Pcm          int64
	Name         string
	isDead       bool
	CategoryTags []string
	List         widget.List
	*ListTags
	*EditorTags
}

type ListTags struct {
	Play   *widget.Clickable
	Edit   *widget.Clickable
	Delete *widget.Clickable
}

type EditorTags struct {
	Flag  *struct{}
	Pole  *struct{}
	Label *struct{}
}

func (tm *TimeMarker) MarkDead() {
	tm.isDead = true
}

func (tm *TimeMarkers) MarkAllDead() {
	for _, it := range *tm {
		it.MarkDead()
	}
}

func (t *TimeMarkers) NewMarker(pcm int64) *TimeMarker {
	if len(*t)+1 > Limit {
		// TODO: display error
		return nil
	}
	newT := &TimeMarker{
		Pcm:          pcm,
		CategoryTags: make([]string, 0, TagsLimit),
		EditorTags: &EditorTags{
			Flag:  &struct{}{},
			Pole:  &struct{}{},
			Label: &struct{}{},
		},
		ListTags: &ListTags{
			Play:   &widget.Clickable{},
			Edit:   &widget.Clickable{},
			Delete: &widget.Clickable{},
		},
	}
	*t = append(*t, newT)
	return newT
}

func (t *TimeMarkers) DeleteDead() {
	*t = slices.DeleteFunc(*t, func(it *TimeMarker) bool {
		return it.isDead
	})
}

func (t *TimeMarkers) GetAsc(idx int) *TimeMarker {
	return (*t)[len(*t)-1-idx]
}
func (t *TimeMarkers) GetDesc(idx int) *TimeMarker {
	return (*t)[idx]
}

func (t *TimeMarkers) sortCb(a, b *TimeMarker) int {
	return int(b.Pcm - a.Pcm)
}

func (t *TimeMarkers) Sorted() TimeMarkers {
	if slices.IsSortedFunc(*t, t.sortCb) {
		return *t
	}
	seq := slices.Values(*t)
	*t = slices.SortedStableFunc(seq, t.sortCb)
	return *t
}
