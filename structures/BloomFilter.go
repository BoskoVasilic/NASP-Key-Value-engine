package structures

import (
	hahs "KVengine/hash"
	utils "KVengine/utils"
)

type BloomFilter struct {
	NumberOfFunctions uint32
	MainLen           uint32
	ArrayMain         []byte
	HashFunc          []hahs.HashWithSeed
}

func NewBloomFilter(numberOfFunctions, mainLen uint32, hashFuncs []hahs.HashWithSeed) *BloomFilter {
	return &BloomFilter{
		NumberOfFunctions: numberOfFunctions,
		MainLen:           mainLen,
		ArrayMain:         make([]byte, mainLen),
		HashFunc:          hashFuncs,
	}
}

// NewBloomFilterWithParams creates a new BloomFilter with the given parameters.
// ExpectedElements is the number of elements that are expected to be inserted into the filter.
// falsePositiveRate is the desired false positive rate.
func NewBloomFilterWithParams(ExpectedElements int, falsePositiveRate float64) *BloomFilter {
	m := utils.CalculateM(ExpectedElements, falsePositiveRate)
	k := utils.CalculateK(ExpectedElements, m)
	hashFuncs := hahs.CreateHashFunctions(uint32(k))
	return &BloomFilter{
		NumberOfFunctions: uint32(k),
		MainLen:           uint32(m),
		ArrayMain:         make([]byte, m/8),
		HashFunc:          hashFuncs,
	}
}
func (b *BloomFilter) Put(data []byte) {
	for _, h := range b.HashFunc {
		adrees := h.Hash(data) % uint64(b.MainLen)
		index := adrees / 8
		bit := adrees % 8
		b.ArrayMain[index] |= (1 << bit)
	}
}
func (b *BloomFilter) Check(data []byte) bool {
	for _, h := range b.HashFunc {
		adrees := h.Hash(data) % uint64(b.MainLen)
		index := adrees / 8
		bit := adrees % 8
		if b.ArrayMain[index]&(1<<bit) == 0 {
			return false
		}
	}
	return true
}
