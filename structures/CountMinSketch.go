package structures

import (
	"KVengine/hash"
	"KVengine/utils"
)

type CountMinSketch struct {
	Depth     uint32
	width     uint32
	Counts    [][]int
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
	for i := range int(cms.Depth) {
		idx := cms.hash(int(i), data)
		cms.Counts[i][idx]++
	}
}

func (cms *CountMinSketch) Estimate(data string) int {
	min := int(^uint(0) >> 1) // Max int
	for i := range int(cms.Depth) {
		idx := cms.hash(i, data)
		if cms.Counts[i][idx] < min {
			min = cms.Counts[i][idx]
		}
	}
	return min
}

func (cms *CountMinSketch) Serialize() []byte {
	data := make([]byte, 0)
	data = append(data, utils.Uint32ToBytes(cms.Depth)...)
	data = append(data, utils.Uint32ToBytes(cms.width)...)
	for _, hf := range cms.hashFuncs {
		data = append(data, hf.Seed...)
	}
	for i := range cms.Counts {
		for _, count := range cms.Counts[i] {
			data = append(data, utils.Uint32ToBytes(uint32(count))...)
		}
	}
	return data
}

func DeserializeCMS(data []byte) *CountMinSketch {
	offset := 0
	depth := utils.BytesToUint32(data[offset : offset+4])
	offset += 4
	width := utils.BytesToUint32(data[offset : offset+4])
	offset += 4
	hashFuncs := make([]hash.HashWithSeed, depth)
	for i := range hashFuncs {
		hashFuncs[i].Seed = data[offset : offset+4]
		offset += 4
	}
	counts := make([][]int, depth)
	for i := uint32(0); i < depth; i++ {
		counts[i] = make([]int, width)
		for j := uint32(0); j < width; j++ {
			counts[i][j] = int(utils.BytesToUint32(data[offset : offset+4]))
			offset += 4
		}
	}
	return &CountMinSketch{depth, width, counts, hashFuncs}
}
