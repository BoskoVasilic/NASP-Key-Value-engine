package structures

import (
	"KVengine/hash"
	utils "KVengine/utils"
	"fmt"
)

type BloomFilter struct {
	NumberOfFunctions uint32
	MainLen           uint32
	ArrayMain         []byte
	HashFunc          []hash.HashWithSeed
}

func NewBloomFilter(numberOfFunctions, mainLen uint32, hashFuncs []hash.HashWithSeed) *BloomFilter {
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
	if m%8 != 0 {
		m = (m/8 + 1) * 8
	}
	k := utils.CalculateK(ExpectedElements, m)
	hashFuncs := hash.CreateHashFunctions(uint32(k))
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
		fmt.Println(adrees)
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
func (b *BloomFilter) SerializeBloomF() []byte {
	listOfBytes := make([]byte, 0)
	NumberOfFunctionsBytes := utils.Uint32ToBytes(b.NumberOfFunctions)
	listOfBytes = append(listOfBytes, NumberOfFunctionsBytes...)
	MainLenBytes := utils.Uint32ToBytes(b.MainLen)
	listOfBytes = append(listOfBytes, MainLenBytes...)
	for _, h := range b.HashFunc {
		listOfBytes = append(listOfBytes, h.Seed...)
	}
	listOfBytes = append(listOfBytes, b.ArrayMain...)

	return listOfBytes
}
func DeserializeBloomF(data []byte) *BloomFilter {
	NumberOfFunctions := utils.BytesToUint32(data[:4])
	MainLen := utils.BytesToUint32(data[4:8])
	hashFuncs := make([]hash.HashWithSeed, NumberOfFunctions)
	for i := 0; i < int(NumberOfFunctions); i++ {
		hashFuncs[i] = hash.HashWithSeed{Seed: data[8+i*4 : 12+i*4]}
	}
	ArrayMain := data[4+4+NumberOfFunctions*4:]
	return &BloomFilter{
		NumberOfFunctions: NumberOfFunctions,
		MainLen:           MainLen,
		ArrayMain:         ArrayMain,
		HashFunc:          hashFuncs,
	}
}
