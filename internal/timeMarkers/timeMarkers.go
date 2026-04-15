package timemarkers

import (
	"slices"

	"gioui.org/widget"
)

const (
	Limit     = 100
	TagsLimit = 10
)

// TODO: Maybe we can live without pointers here
type TimeMarkers []*TimeMarker

func NewTimeMarkers() TimeMarkers {
	return make([]*TimeMarker, 0, Limit)
}

type TimeMarker struct {
	Samples      int    `json:"samples,omitempty"`
	Name         string `json:"name,omitempty"`
	isDead       bool
	Notes        string      `json:"notes,omitempty"`
	CategoryTags []string    `json:"category_tags,omitempty"`
	List         widget.List `json:"-"`
	ListTags     `json:"-"`
	EditorTags   `json:"-"`
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

func (t *TimeMarkers) NewMarker(samples int) *TimeMarker {
	if len(*t)+1 > Limit {
		// TODO: display error
		return nil
	}
	newT := &TimeMarker{
		Samples:      samples,
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

func (t *TimeMarkers) AttachNewMarker(newT TimeMarker) bool {
	if len(*t)+1 > Limit {
		// TODO: display error
		return false
	}
	newT.EditorTags = EditorTags{
		Flag:  &struct{}{},
		Pole:  &struct{}{},
		Label: &struct{}{},
	}
	*t = append(*t, &newT)
	return true
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
	return int(a.Samples - b.Samples)
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

func (t *TimeMarkers) IsEmpty() bool {
	return len(*t) == 0
}

func (t *TimeMarkers) SanitizeSamples(maxSamples int) {
	for _, it := range *t {
		if it.Samples > maxSamples {
			it.Samples = 0
			it.CategoryTags = append(it.CategoryTags, "Redacted")
		}
	}
}
