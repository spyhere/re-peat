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
		List: widget.List{},
	}
	*t = append(*t, newT)
	return newT
}

func (t *TimeMarkers) DeleteDead() {
	*t = slices.DeleteFunc(*t, func(it *TimeMarker) bool {
		return it.isDead
	})
}

func (t *TimeMarkers) Get(idx int, asc bool) *TimeMarker {
	if idx < 0 || idx > len(*t)-1 {
		return nil
	}
	if asc {
		return (*t)[idx]
	}
	return (*t)[len(*t)-1-idx]
}

func (t *TimeMarkers) GetIndex(m *TimeMarker, asc bool) int {
	if m == nil {
		return -1
	}
	ind := slices.Index(*t, m)
	if ind == -1 {
		return -1
	}
	if asc {
		return ind
	}
	return len(*t) - 1 - ind
}

func (t *TimeMarkers) sortCb(a, b *TimeMarker) int {
	return int(a.Pcm - b.Pcm)
}

// TODO: Sort only on marker manipulation
func (t *TimeMarkers) Sorted() TimeMarkers {
	if slices.IsSortedFunc(*t, t.sortCb) {
		return *t
	}
	seq := slices.Values(*t)
	*t = slices.SortedStableFunc(seq, t.sortCb)
	return *t
}
