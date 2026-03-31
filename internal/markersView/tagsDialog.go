package markersview

import (
	"slices"
	"strings"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/spyhere/re-peat/internal/common"
)

func newTagsDialog(capacity int) tagsDialog {
	return tagsDialog{
		filterChips: make([]*common.FilterChip, 0, capacity),
	}
}

type tagsDialog struct {
	filterChips []*common.FilterChip
}

func (t *tagsDialog) createFreshChips(allChips, enabledChips map[string]struct{}) []*common.FilterChip {
	t.filterChips = t.filterChips[:0]
	for chip := range allChips {
		isEnabled := false
		if _, exists := enabledChips[chip]; exists {
			isEnabled = true
		}
		newChip := &common.FilterChip{
			Text:     chip,
			Selected: isEnabled,
			Cl:       &widget.Clickable{},
		}
		t.filterChips = append(t.filterChips, newChip)
	}
	t.filterChips = slices.SortedFunc(slices.Values(t.filterChips), func(fc1, fc2 *common.FilterChip) int {
		return strings.Compare(fc1.Text, fc2.Text)
	})
	return t.filterChips
}

func (t *tagsDialog) getCursorAndHandleEvents(gtx layout.Context) (pointer.Cursor, bool) {
	curCursor := pointer.CursorDefault
	for _, it := range t.filterChips {
		if it.Cl.Clicked(gtx) {
			it.Selected = !it.Selected
		}
		if it.Cl.Hovered() {
			curCursor = pointer.CursorPointer
		}
	}
	if curCursor == pointer.CursorDefault {
		return curCursor, false
	}
	return curCursor, true
}
