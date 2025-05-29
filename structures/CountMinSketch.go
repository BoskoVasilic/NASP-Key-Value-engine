package structures

import (
	"KVengine/hash"
)

type CountMinSketch struct {
	depth     uint32
	width     uint32
	counts    [][]int
	hashFuncs []hash.HashWithSeed
}

func NewCountMinSketch(depth, width uint32) *CountMinSketch {
	counts := make([][]int, depth)
	hashFuncs := hash.CreateHashFunctions(depth)
	for i := range int(depth) {
		counts[i] = make([]int, int(width))
	}
	return &CountMinSketch{depth, width, counts, hashFuncs}
}

func (cms *CountMinSketch) hash(i int, data string) int {
	return int(uint32(cms.hashFuncs[i].Hash([]byte(data))) % cms.width)
}

func (cms *CountMinSketch) Add(data string) {
	for i := range int(cms.depth) {
		idx := cms.hash(int(i), data)
		cms.counts[i][idx]++
	}
}

func (cms *CountMinSketch) Estimate(data string) int {
	min := int(^uint(0) >> 1) // Max int
	for i := range int(cms.depth) {
		idx := cms.hash(i, data)
		if cms.counts[i][idx] < min {
			min = cms.counts[i][idx]
		}
	}
	return min
}
