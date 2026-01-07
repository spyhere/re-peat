package waverenderer

import "cmp"

// Lock "this" between "from" and "to"
func clamp[T cmp.Ordered](from T, this T, to T) T {
	return min(max(from, this), to)
}
