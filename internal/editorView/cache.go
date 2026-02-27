package editorview

func newCache() cache {
	return cache{
		peakMap: make(map[int][][2]float32),
		levels:  make([]int, maxScrollLvl+1),
		workers: make([]*cacheWorker, maxScrollLvl+1),
	}
}

// Stores peak map where "samplesPerPx" is key (level)
type cache struct {
	peakMap     map[int][][2]float32
	curSlice    [][2]float32
	workers     []*cacheWorker
	isPopulated bool
	levels      []int // Stores possible "samplesPerPx" values
	curLvl      int
	leftB       int
}

// Used to build one level (samplesPerPx) of cache
type cacheWorker struct {
	min          float32
	max          float32
	samplesPerPx int
	count        int
	sliceIdx     int
}

func (c cache) getLevel(spp float32) int {
	for i := len(c.levels) - 1; i >= 0; i-- {
		if float32(c.levels[i]) >= spp {
			return c.levels[i]
		}
	}
	return c.levels[0]
}
