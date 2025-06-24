package structures

import (
	utils "KVengine/utils"
)

type SSTable struct {
	Podaci   *utils.Iterator[any, utils.ExportEntry] // Iterator over the data in the SSTable
	Filter   *BloomFilter
	blokSize uint64 // Size of the block in bytes
}

func NewSSTable(podaci *utils.Iterator[any, utils.ExportEntry], NumberOfEnyerys int, blokSize uint64) *SSTable {
	return &SSTable{
		Podaci:   podaci,
		Filter:   NewBloomFilterWithParams(NumberOfEnyerys, 0.01), // Example parameters for the Bloom filter
		blokSize: blokSize,
	}
}

func (s *SSTable) GenerateSSTabel() {
	ofsset := uint64(0)

	for {
		entry, success := s.Podaci.Next()
		if !success {
			break
		}
		s.Filter.Put([]byte(entry.Key))
		writing := entry.GetBytes()
		writingIndex := append([]byte(entry.Key), utils.Uint64ToBytes(ofsset)...)
		ofsset = ofsset + uint64(len(writing))

	}

}
