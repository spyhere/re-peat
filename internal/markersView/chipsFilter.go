package markersview

import (
	"github.com/spyhere/re-peat/internal/common"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
)

func newChipsFilter(capacity int) chipsFilter {
	return chipsFilter{
		all:        make(map[string]struct{}, capacity),
		enabledMap: make(map[string]struct{}, capacity),
		enabled:    make([]string, 0, capacity),
	}
}

type chipsFilter struct {
	all        map[string]struct{}
	enabledMap map[string]struct{} // optimisation to quickly look up when creating tags filter masonry
	enabled    []string
}

func (c *chipsFilter) getEnabledChips() []string {
	return c.enabled
}

func (c *chipsFilter) purge() {
	for chip := range c.all {
		delete(c.all, chip)
	}
}

// Calculate all unique tags from all markers
func (c *chipsFilter) recreate(markers tm.TimeMarkers) {
	c.purge()
	for _, marker := range markers {
		if !marker.IsAlive() {
			continue
		}
		for _, tag := range marker.CategoryTags {
			c.all[tag] = struct{}{}
		}
	}
}

// Update all tags with unique tags if there are any
func (c *chipsFilter) updateAll(tags []string) {
	for _, tag := range tags {
		if _, ok := c.all[tag]; !ok {
			c.all[tag] = struct{}{}
		}
	}
}

// Check whether enabled chips still exist in "all"
func (c *chipsFilter) reconcileEnabled(markers tm.TimeMarkers) {
	c.recreate(markers)
	idx := 0
	for _, enabledChip := range c.enabled {
		if _, ok := c.all[enabledChip]; ok {
			c.enabled[idx] = enabledChip
			idx++
		} else {
			delete(c.enabledMap, enabledChip)
		}
	}
	c.enabled = c.enabled[:idx]
}

// Incremental update list of enabled tags
func (c *chipsFilter) updateEnabled(chips []common.FilterChip) {
	c.enabled = c.enabled[:0]
	for chip := range c.enabledMap {
		delete(c.enabledMap, chip)
	}
	for _, filterChip := range chips {
		if filterChip.Selected {
			c.enabled = append(c.enabled, filterChip.Text)
			c.enabledMap[filterChip.Text] = struct{}{}
		}
	}
}
