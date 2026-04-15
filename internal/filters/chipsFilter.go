package filters

import (
	"github.com/spyhere/re-peat/internal/common"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
)

func NewChipsFilter(capacity int) ChipsFilter {
	return ChipsFilter{
		All:        make(map[string]struct{}, capacity),
		EnabledMap: make(map[string]struct{}, capacity),
		Enabled:    make([]string, 0, capacity),
	}
}

type ChipsFilter struct {
	All        map[string]struct{}
	EnabledMap map[string]struct{} // optimisation to quickly look up when creating tags filter masonry
	Enabled    []string
}

func (c *ChipsFilter) GetEnabledChips() []string {
	return c.Enabled
}

func (c *ChipsFilter) Purge() {
	for chip := range c.All {
		delete(c.All, chip)
	}
	for chip := range c.EnabledMap {
		delete(c.EnabledMap, chip)
	}
	c.Enabled = c.Enabled[:0]
}

// Calculate all unique tags from all markers
func (c *ChipsFilter) Recreate(markers tm.TimeMarkers) {
	c.Purge()
	for _, marker := range markers {
		if !marker.IsAlive() {
			continue
		}
		for _, tag := range marker.CategoryTags {
			c.All[tag] = struct{}{}
		}
	}
}

// Update all tags with unique tags if there are any
func (c *ChipsFilter) UpdateAll(tags []string) {
	for _, tag := range tags {
		c.All[tag] = struct{}{}
	}
}

// Check whether enabled chips still exist in "all"
func (c *ChipsFilter) ReconcileEnabled(markers tm.TimeMarkers) {
	c.Recreate(markers)
	idx := 0
	for _, enabledChip := range c.Enabled {
		if _, ok := c.All[enabledChip]; ok {
			c.Enabled[idx] = enabledChip
			idx++
		} else {
			delete(c.EnabledMap, enabledChip)
		}
	}
	c.Enabled = c.Enabled[:idx]
}

// Incremental update list of enabled tags
func (c *ChipsFilter) UpdateEnabled(chips []common.FilterChip) {
	c.Enabled = c.Enabled[:0]
	for chip := range c.EnabledMap {
		delete(c.EnabledMap, chip)
	}
	for _, filterChip := range chips {
		if filterChip.Selected {
			c.Enabled = append(c.Enabled, filterChip.Text)
			c.EnabledMap[filterChip.Text] = struct{}{}
		}
	}
}
