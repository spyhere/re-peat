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
	Notes        string
	CategoryTags []string
	List         widget.List
	ListTags
	EditorTags
}

type ListTags struct {
	Play    widget.Clickable
	Comment widget.Clickable
	Edit    widget.Clickable
	Delete  widget.Clickable
}

type EditorTags struct {
	Flag  *struct{}
	Pole  *struct{}
	Label *struct{}
}

func (m *TimeMarker) MarkDead() {
	m.isDead = true
}
func (m *TimeMarker) IsAlive() bool {
	return !m.isDead
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
		EditorTags: EditorTags{
			Flag:  &struct{}{},
			Pole:  &struct{}{},
			Label: &struct{}{},
		},
		List: widget.List{},
	}
	*t = append(*t, newT)
	return newT
}

func (t *TimeMarkers) DeleteDead() (hasDeletion bool) {
	hasDeletion = false
	*t = slices.DeleteFunc(*t, func(it *TimeMarker) bool {
		if it.isDead {
			hasDeletion = true
			return true
		}
		return false
	})
	t.Sort()
	return hasDeletion
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

func (t *TimeMarkers) Sort() {
	if slices.IsSortedFunc(*t, t.sortCb) {
		return
	}
	seq := slices.Values(*t)
	*t = slices.SortedStableFunc(seq, t.sortCb)
}

func (t *TimeMarkers) Sorted() TimeMarkers {
	t.Sort()
	return *t
}
