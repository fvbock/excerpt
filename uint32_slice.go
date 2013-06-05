package excerpt

import (
	"sort"
)

type Uint32Slice []uint32

func (p Uint32Slice) Len() int           { return len(p) }
func (p Uint32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Uint32s sorts a slice of uint32s in increasing order.
func SortUint32s(a []uint32) { sort.Sort(Uint32Slice(a)) }
