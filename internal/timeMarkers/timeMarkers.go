package timemarkers

import "slices"

const Limit = 100

type TimeMarkers []*TimeMarker

func NewTimeMarkers() TimeMarkers {
	return make([]*TimeMarker, 0, Limit)
}

type TimeMarker struct {
	Pcm    int64
	Name   string
	Tags   *Tags
	isDead bool
}

// Gio geometry Tags for Editor view
type Tags struct {
	Flag  *struct{}
	Pole  *struct{}
	Label *struct{}
}

func (tm *TimeMarker) MarkDead() {
	tm.isDead = true
}

func (t *TimeMarkers) NewMarker(pcm int64) *TimeMarker {
	if len(*t)+1 > Limit {
		// TODO: display error
		return nil
	}
	newT := &TimeMarker{
		Pcm: pcm,
		Tags: &Tags{
			Flag:  &struct{}{},
			Pole:  &struct{}{},
			Label: &struct{}{},
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
