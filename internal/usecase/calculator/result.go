package calculator

import "sort"

// PackLine is a stable, ordered representation of one pack size and its quantity.
type PackLine struct {
	Pack     int
	Quantity int
}

// Result is an application-level output for one calculation:
// ordered list of pack sizes and quantities (ascending by pack size).
type Result struct {
	Packs []PackLine
}

// breakdownSorted turns the map from MinPacks into ascending order by pack size
// so API responses and clients get a deterministic, stable order.
func breakdownSorted(m map[int]int) []PackLine {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	out := make([]PackLine, 0, len(keys))
	for _, k := range keys {
		out = append(out, PackLine{Pack: k, Quantity: m[k]})
	}
	return out
}

