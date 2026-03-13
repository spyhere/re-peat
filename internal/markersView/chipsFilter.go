package markersview

import (
	"github.com/spyhere/re-peat/internal/common"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
)

func newChipsFilter(capacity int) chipsFilter {
	return chipsFilter{
		all:        make(map[string]struct{}, capacity),
		enabled:    make(map[string]struct{}, capacity),
		enabledBuf: make([]string, 0, capacity),
	}
}

type chipsFilter struct {
	all        map[string]struct{}
	enabled    map[string]struct{}
	enabledBuf []string
}

func (c *chipsFilter) getEnabledChips() []string {
	c.enabledBuf = c.enabledBuf[:0]
	for chipName := range c.enabled {
		c.enabledBuf = append(c.enabledBuf, chipName)
	}
	return c.enabledBuf
}

func (c *chipsFilter) purge() {
	for chip := range c.all {
		delete(c.all, chip)
	}
}

func (c *chipsFilter) recreate(markers tm.TimeMarkers) {
	c.purge()
	for _, marker := range markers {
		for _, tag := range marker.CategoryTags {
			c.all[tag] = struct{}{}
		}
	}
}

func (c *chipsFilter) updateAll(tags []string) {
	for _, tag := range tags {
		if _, ok := c.all[tag]; !ok {
			c.all[tag] = struct{}{}
		}
	}
}

func (c *chipsFilter) updateEnabled(chips []*common.FilterChip) {
	for chip := range c.enabled {
		delete(c.enabled, chip)
	}
	for _, filterChip := range chips {
		if filterChip.Selected {
			c.enabled[filterChip.Text] = struct{}{}
		}
	}
}
