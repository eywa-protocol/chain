/*
* Copyright 2021 by EYWA chain <blockchain@digiu.ai>
*/
package common

import "sort"

type Uint64Slice []uint64

func (p Uint64Slice) Len() int           { return len(p) }
func (p Uint64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func SortUint64s(l []uint64) {
	sort.Stable(Uint64Slice(l))
}
