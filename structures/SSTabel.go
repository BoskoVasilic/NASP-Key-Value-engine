package structures

import (
	"KVengine/structures/auxiliary"
	utils "KVengine/utils"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

const (
	uvarMaxSize = 10 // Maximum size of a uvarint in bytes
)

type SSTable[T any, H any] struct {
	Podaci           *utils.Iterator[T, H, *Entry] // Iterator over the data in the SSTable
	Filter           *BloomFilter
	numberOfEnterys  int
	oneFile          bool   // Indicates if the SSTable is stored in a single file
	SumeeryIndex     uint64 //NumberOfitems in the summary index
	BlockManager     *BlockManager
	globalDictionary *GlobalDictionary
	comprest         bool // Indicates if the SSTable is compressed
	startPath        string
}

func NewSSTable[T any, H any](podaci *utils.Iterator[T, H, *Entry], NumberOfEnyerys int, SumeeryIndex uint64, oneFile bool, BlocMenager *BlockManager, dictinery *GlobalDictionary, isCompreset bool) *SSTable[T, H] {
	return &SSTable[T, H]{
		Podaci:           podaci,
		Filter:           NewBloomFilterWithParams(NumberOfEnyerys, 0.01), // Example parameters for the Bloom filter
		SumeeryIndex:     SumeeryIndex,
		numberOfEnterys:  NumberOfEnyerys,
		oneFile:          oneFile,
		BlockManager:     BlocMenager,
		globalDictionary: dictinery,
		comprest:         isCompreset,
		startPath:        "./data/SSTabel" + fmt.Sprintf("%d", time.Now().UnixMicro()),
	}
}

func (s *SSTable[T, H]) GenerateSSTabel() string {
	treeMerkle := NewMerkleTree()
	if s.comprest {
		data, firstKey, lastKey := AddItems(s.globalDictionary, s.Podaci)

		return WriteToFilesCompresed(s.startPath, s.oneFile, s.BlockManager, s.numberOfEnterys, s.Filter, s.SumeeryIndex, data, firstKey, lastKey, treeMerkle)
	} else {
		data, firstKey, lastkey := AddItemsWithNoCompresion(s.globalDictionary, s.Podaci)
		return WriteToFile(s.startPath, s.oneFile, s.BlockManager, s.numberOfEnterys, s.Filter, s.SumeeryIndex, data, firstKey, lastkey, treeMerkle)
	}
}
func WriteToFilesCompresed[T any, H any](StartPath string, oneFile bool, BlockManager *BlockManager, numberOfEnterys int, Filter *BloomFilter, SumeeryIndex uint64,
	data *utils.Iterator[T, H, *ExportEntry], firstKey string, lastkey string, treeMerkle *MerkleTree) string {
	if oneFile {
		file := NewBlockManagerWrap(StartPath, BlockManager)
		file.WriteToFile([]byte{1}) // Write 1 to indicate it is compressed
		file.WriteToFile([]byte{1}) // Write 1 to indicate single file format
		//Write 5 Uint64 values for the offsets
		file.WriteToFile(utils.Uint64ToBytes(uint64(numberOfEnterys))) // Number of entries
		file.WriteToFile(utils.Uint64ToBytes(uint64(0)))               // Offset for Data
		file.WriteToFile(utils.Uint64ToBytes(uint64(0)))               // Offset for Blom filter
		file.WriteToFile(utils.Uint64ToBytes(uint64(0)))               // Offset for IndexSummery
		file.WriteToFile(utils.Uint64ToBytes(uint64(0)))               // Offset for Index blocks
		file.WriteToFile(utils.Uint64ToBytes(uint64(0)))               // Offset for Merkle tree
		file.WriteToFile(utils.Uint64ToBytes(uint64(0)))               // Offset for Merkle tree
		IndexSummery := make([]byte, 0)
		IndexBloc := make([]byte, 0)
		size, EntryNumbers := WriteDataBlockCompresed(file, data, &IndexSummery, &IndexBloc, Filter, SumeeryIndex, treeMerkle)
		treeMerkle.build() // Build the Merkle tree after writing all data
		// Write the Bloom filter
		bloomFilterData := Filter.SerializeBloomF()
		file.WriteToFile(bloomFilterData)
		bloomFilterOffset := uint64(len(bloomFilterData))
		// Write the Index summary
		file.WriteToFile(IndexSummery)
		indexSummaryOffset := uint64(len(IndexSummery))
		// Write the Index blocks
		file.WriteToFile(IndexBloc)
		indexBlockOffset := uint64(len(IndexBloc))
		// Write the offsets to the file
		//Write Merkel tree in this section
		dataMerkle := treeMerkle.SerializeMerkle()
		dataMerkleLength := uint64(len(dataMerkle))
		file.WriteToFile(dataMerkle)
		// Generating heeder with offsets
		Heeder := make([]byte, 0)
		offsetStart := uint64(58) // Starting offset for the header
		Heeder = append(Heeder, utils.Uint64ToBytes(EntryNumbers)...)
		Heeder = append(Heeder, utils.Uint64ToBytes(offsetStart)...)
		Heeder = append(Heeder, utils.Uint64ToBytes(offsetStart+size)...)
		Heeder = append(Heeder, utils.Uint64ToBytes(size+bloomFilterOffset+offsetStart)...)
		Heeder = append(Heeder, utils.Uint64ToBytes(size+bloomFilterOffset+indexSummaryOffset+offsetStart)...)
		Heeder = append(Heeder, utils.Uint64ToBytes(size+bloomFilterOffset+indexBlockOffset+indexSummaryOffset+offsetStart)...)
		Heeder = append(Heeder, utils.Uint64ToBytes(size+bloomFilterOffset+indexBlockOffset+indexSummaryOffset+offsetStart+dataMerkleLength)...) //Need Merkle tree offset
		//file.FlushBuffer(false)
		// Flush the buffer to write the data to the file
		file.WriteToFile(utils.Uint64ToBytes(uint64(len(firstKey))))
		file.WriteToFile([]byte(firstKey))
		file.WriteToFile(utils.Uint64ToBytes(uint64(len(lastkey))))
		file.WriteToFile([]byte(lastkey))
		// Write the Merkle tree data
		file.SetCurrentPosition(2)
		file.OverwriteDataInFile(Heeder) // Overwrite the header with the offsets
		file.FlushBuffer(false)          // Flush the buffer to write the header to the file
		return StartPath
	} else {
		// Handle multiple files case
		// Implement logic to write data to multiple files
		// For example, you can create a new file for each block of data
		// and write the corresponding index and bloom filter information.
		// This is a placeholder implementation.
		file := NewBlockManagerWrap(StartPath+"TOC", BlockManager)

		file.WriteToFile([]byte{1})                                    // Write 1 to say it is compressed
		file.WriteToFile([]byte{0})                                    // Write 0 to indicate multiple files format
		file.WriteToFile(utils.Uint64ToBytes(uint64(numberOfEnterys))) // Offset for Data
		// Write paths to other files
		data_path := StartPath + "dataSstable"
		indexSummery_path := StartPath + "indexSummerySstable"
		indexBlock_path := StartPath + "indexBlockSstable"
		bloomFilter_path := StartPath + "bloomFilterSstable"
		merkleTree_path := StartPath + "merkleTreeSstable"
		Data := make([]byte, 0)
		Data = binary.AppendUvarint(Data, uint64(len(data_path)))
		Data = append(Data, []byte(data_path)...)
		Data = binary.AppendUvarint(Data, uint64(len(bloomFilter_path)))
		Data = append(Data, []byte(bloomFilter_path)...)
		Data = binary.AppendUvarint(Data, uint64(len(indexSummery_path)))
		Data = append(Data, []byte(indexSummery_path)...)
		Data = binary.AppendUvarint(Data, uint64(len(indexBlock_path)))
		Data = append(Data, []byte(indexBlock_path)...)
		Data = binary.AppendUvarint(Data, uint64(len(merkleTree_path)))
		Data = append(Data, []byte(merkleTree_path)...)
		Data = append(Data, utils.Uint64ToBytes(uint64(len(firstKey)))...)
		Data = append(Data, []byte(firstKey)...)
		Data = append(Data, utils.Uint64ToBytes(uint64(len(lastkey)))...)
		Data = append(Data, []byte(lastkey)...)
		file.WriteToFile(Data) // Write the paths to the file
		file.FlushBuffer(false)
		fileData := NewBlockManagerWrap(data_path, BlockManager)
		indexSummery := make([]byte, 0)
		indexBlock := make([]byte, 0)
		_, EntryNumbers := WriteDataBlockCompresed(fileData, data, &indexSummery, &indexBlock, Filter, SumeeryIndex, treeMerkle)
		fileData.FlushBuffer(false)
		treeMerkle.build() // Build the Merkle tree after writing all data
		file.SetCurrentPosition(2)
		file.OverwriteDataInFile(utils.Uint64ToBytes(uint64(EntryNumbers)))
		// Overwrite the number of entries in the header
		// Write the Bloom filter
		bloomFilterData := Filter.SerializeBloomF()
		fileBloom := NewBlockManagerWrap(bloomFilter_path, BlockManager)
		fileBloom.WriteToFile(binary.AppendUvarint([]byte{}, uint64(len(bloomFilterData))))
		fileBloom.WriteToFile(bloomFilterData)
		fileBloom.FlushBuffer(false)
		fileIndexSummery := NewBlockManagerWrap(indexSummery_path, BlockManager)
		fileIndexSummery.WriteToFile(indexSummery)
		fileIndexSummery.FlushBuffer(false)
		fileIndexBlock := NewBlockManagerWrap(indexBlock_path, BlockManager)
		fileIndexBlock.WriteToFile(indexBlock)
		fileIndexBlock.FlushBuffer(false)
		fileMerckle := NewBlockManagerWrap(merkleTree_path, BlockManager)
		dataMerkle := treeMerkle.SerializeMerkle()
		dataMerkleLength := uint64(len(dataMerkle))
		// Write the Merkle tree data
		fileMerckle.WriteToFile(utils.Uint64ToBytes(dataMerkleLength)) // Write the length of the Merkle tree data
		fileMerckle.WriteToFile(dataMerkle)
		fileMerckle.FlushBuffer(false)
		//Creating Merkle tree file need to be finished
		return StartPath + "TOC"
	}
}
func WriteToFile[T any, H any](StartPath string, oneFile bool, BlockManager *BlockManager, numberOfEnterys int,
	Filter *BloomFilter, SumeeryIndex uint64, data *utils.Iterator[T, H, *Entry], firstKey string, lastkey string, tree *MerkleTree) string {
	if oneFile {
		file := NewBlockManagerWrap(StartPath, BlockManager)
		file.WriteToFile([]byte{0}) // Write 0 to indicate it is not compressed
		file.WriteToFile([]byte{1}) // Write 1 to indicate single file format
		//Write 5 Uint64 values for the offsets
		file.WriteToFile(utils.Uint64ToBytes(uint64(numberOfEnterys))) // Number of entries
		file.WriteToFile(utils.Uint64ToBytes(uint64(0)))               // Offset for Data
		file.WriteToFile(utils.Uint64ToBytes(uint64(0)))               // Offset for Blom filter
		file.WriteToFile(utils.Uint64ToBytes(uint64(0)))               // Offset for IndexSummery
		file.WriteToFile(utils.Uint64ToBytes(uint64(0)))               // Offset for Index blocks
		file.WriteToFile(utils.Uint64ToBytes(uint64(0)))               // Offset for Merkle tree
		file.WriteToFile(utils.Uint64ToBytes(uint64(0)))               // Offset for Merkle tree
		IndexSummery := make([]byte, 0)
		IndexBloc := make([]byte, 0)
		size, EntryNumbers := WriteDataBlock(file, data, &IndexSummery, &IndexBloc, Filter, SumeeryIndex, tree)
		// Write the Bloom filter
		tree.build() // Build the Merkle tree after writing all data
		bloomFilterData := Filter.SerializeBloomF()
		file.WriteToFile(bloomFilterData)
		bloomFilterOffset := uint64(len(bloomFilterData))
		// Write the Index summary
		file.WriteToFile(IndexSummery)
		indexSummaryOffset := uint64(len(IndexSummery))
		// Write the Index blocks
		file.WriteToFile(IndexBloc)
		indexBlockOffset := uint64(len(IndexBloc))
		// Write the offsets to the file
		//Write Merkel tree in this section
		dataMerkle := tree.SerializeMerkle()
		dataMerkleLength := uint64(len(dataMerkle))
		file.WriteToFile(dataMerkle) // Write the Merkle tree data
		// Generating heeder with offsets
		Heeder := make([]byte, 0)
		offsetStart := uint64(58) // Starting offset for the header
		Heeder = append(Heeder, utils.Uint64ToBytes(EntryNumbers)...)
		Heeder = append(Heeder, utils.Uint64ToBytes(offsetStart)...)
		Heeder = append(Heeder, utils.Uint64ToBytes(offsetStart+size)...)
		Heeder = append(Heeder, utils.Uint64ToBytes(size+bloomFilterOffset+offsetStart)...)
		Heeder = append(Heeder, utils.Uint64ToBytes(size+bloomFilterOffset+indexSummaryOffset+offsetStart)...)
		Heeder = append(Heeder, utils.Uint64ToBytes(size+bloomFilterOffset+indexBlockOffset+indexSummaryOffset+offsetStart)...)
		Heeder = append(Heeder, utils.Uint64ToBytes(size+bloomFilterOffset+indexBlockOffset+indexSummaryOffset+offsetStart+dataMerkleLength)...) //Need Merkle tree offset
		//file.FlushBuffer(false)

		// Flush the buffer to write the data to the file
		file.WriteToFile(utils.Uint64ToBytes(uint64(len(firstKey))))
		file.WriteToFile([]byte(firstKey))
		file.WriteToFile(utils.Uint64ToBytes(uint64(len(lastkey))))
		file.WriteToFile([]byte(lastkey))
		file.SetCurrentPosition(2)
		file.OverwriteDataInFile(Heeder) // Overwrite the header with the offsets
		file.FlushBuffer(false)
		return StartPath // Flush the buffer to write the header to the file
		// Flush the buffer to write the header to the file
	} else {
		// Handle multiple files case
		// Implement logic to write data to multiple files
		// For example, you can create a new file for each block of data
		// and write the corresponding index and bloom filter information.
		// This is a placeholder implementation.
		file := NewBlockManagerWrap(StartPath+"TOC", BlockManager)

		file.WriteToFile([]byte{0}) // Write 0 to say it is not compressed
		file.WriteToFile([]byte{0}) // Write 0 to indicate multiple files format
		// Write paths to other files
		file.WriteToFile(utils.Uint64ToBytes(uint64(numberOfEnterys))) // Offset for Data
		data_path := StartPath + "dataSstable"
		indexSummery_path := StartPath + "indexSummerySstable"
		indexBlock_path := StartPath + "indexBlockSstable"
		bloomFilter_path := StartPath + "bloomFilterSstable"
		merkleTree_path := StartPath + "merkleTreeSstable"
		Data := make([]byte, 0)
		Data = binary.AppendUvarint(Data, uint64(len(data_path)))
		Data = append(Data, []byte(data_path)...)
		Data = binary.AppendUvarint(Data, uint64(len(bloomFilter_path)))
		Data = append(Data, []byte(bloomFilter_path)...)
		Data = binary.AppendUvarint(Data, uint64(len(indexSummery_path)))
		Data = append(Data, []byte(indexSummery_path)...)
		Data = binary.AppendUvarint(Data, uint64(len(indexBlock_path)))
		Data = append(Data, []byte(indexBlock_path)...)
		Data = binary.AppendUvarint(Data, uint64(len(merkleTree_path)))
		Data = append(Data, []byte(merkleTree_path)...)
		Data = append(Data, utils.Uint64ToBytes(uint64(len(firstKey)))...)
		Data = append(Data, []byte(firstKey)...)
		Data = append(Data, utils.Uint64ToBytes(uint64(len(lastkey)))...)
		Data = append(Data, []byte(lastkey)...)
		file.WriteToFile(Data) // Write the paths to the file
		file.FlushBuffer(false)
		fileData := NewBlockManagerWrap(data_path, BlockManager)
		indexSummery := make([]byte, 0)
		indexBlock := make([]byte, 0)
		_, EntryEnters := WriteDataBlock(fileData, data, &indexSummery, &indexBlock, Filter, SumeeryIndex, tree)
		tree.build()
		file.SetCurrentPosition(2)
		file.OverwriteDataInFile(utils.Uint64ToBytes(uint64(EntryEnters))) // Overwrite the number of entries in the header
		fileData.FlushBuffer(false)
		// Write the Bloom filter
		bloomFilterData := Filter.SerializeBloomF()
		fileBloom := NewBlockManagerWrap(bloomFilter_path, BlockManager)
		fileBloom.WriteToFile(binary.AppendUvarint([]byte{}, uint64(len(bloomFilterData))))
		fileBloom.WriteToFile(bloomFilterData)
		fileBloom.FlushBuffer(false)
		fileIndexSummery := NewBlockManagerWrap(indexSummery_path, BlockManager)
		fileIndexSummery.WriteToFile(indexSummery)
		fileIndexSummery.FlushBuffer(false)
		fileIndexBlock := NewBlockManagerWrap(indexBlock_path, BlockManager)
		fileIndexBlock.WriteToFile(indexBlock)
		fileIndexBlock.FlushBuffer(false)
		//Creating Merkle tree file need to be finished
		fileMerckle := NewBlockManagerWrap(merkleTree_path, BlockManager)
		dataMerkle := tree.SerializeMerkle()
		dataMerkleLength := uint64(len(dataMerkle))
		// Write the Merkle tree data
		fileMerckle.WriteToFile(utils.Uint64ToBytes(dataMerkleLength)) // Write the length of the Merkle tree data
		fileMerckle.WriteToFile(dataMerkle)
		fileMerckle.FlushBuffer(false)
		return StartPath + "TOC"
	}
}

func WriteDataBlock[T any, H any](file *BlockManagerWrap, itter *utils.Iterator[T, H, *Entry], IndexSummery *[]byte, IndexBloc *[]byte,
	Filter *BloomFilter, BlockSummery uint64, tree *MerkleTree) (uint64, uint64) {
	// TODO: Implement the function logic here
	entry, ok := itter.Next()
	indexoffset := uint64(0)
	indexTogo := 0
	firtKEyEver := entry.GetKey()
	firstIndexindexSummery := "NoKey"
	lastKey := firtKEyEver
	offset := uint64(0)
	i := uint64(0)
	if !ok {
		fmt.Println("No entries to write.")
		return 0, 0 // No entries to write
	}
	for true {
		i += 1
		if entry == nil {
			continue
		}

		// Create an index entry

		data := entry.toBytes()
		tree.Put(data)
		//fmt.Println(entry.GetKey(), "Data to write:", data)
		Filter.Put([]byte(entry.GetKey())) // Add the key to the global dictionary TODO SOLVE THIS PROBLEM
		file.WriteToFile(data)
		// Write the data to the file
		if firstIndexindexSummery == "NoKey" {
			firstIndexindexSummery = entry.GetKey() // Set firstKey to the first key encountered
		}
		// Create an index entry
		//fmt.Println("Creating index entry for key:", entry.GetKey(), "at offset:", offset)
		IndexBlockEntry := make([]byte, 0)
		IndexBlockEntry = binary.AppendUvarint(IndexBlockEntry, uint64(len(entry.GetKey())))
		IndexBlockEntry = append(IndexBlockEntry, []byte(entry.GetKey())...)
		IndexBlockEntry = binary.AppendUvarint(IndexBlockEntry, offset)
		*IndexBloc = append(*IndexBloc, IndexBlockEntry...)
		offset += uint64(len(data))
		//fmt.Println("Data written to file:", data)
		//fmt.Println("Offset after writing data:", offset)
		indexTogo += 1
		if indexTogo%int(BlockSummery) == 0 {
			// Create an index entry
			indexEntry := make([]byte, 0)
			indexEntry = binary.AppendUvarint(indexEntry, uint64(len(firstIndexindexSummery)))
			indexEntry = append(indexEntry, []byte(firstIndexindexSummery)...)
			indexEntry = binary.AppendUvarint(indexEntry, uint64(len(entry.GetKey())))
			indexEntry = append(indexEntry, []byte(entry.GetKey())...)
			indexEntry = binary.AppendUvarint(indexEntry, indexoffset)
			*IndexSummery = append(*IndexSummery, indexEntry...)
			indexoffset = uint64(len(*IndexBloc))
			firstIndexindexSummery = "NoKey" // Reset firstKey for the next block
		}
		if entry.GetKey() > lastKey {
			lastKey = entry.GetKey()
		}
		// Update the offset
		entry, ok = itter.Next()
		if !ok {
			if firstIndexindexSummery != "NoKey" {
				// If we reach the end of the data and firstKey is still set, add the last entry
				indexEntryS := make([]byte, 0)
				indexEntryS = binary.AppendUvarint(indexEntryS, uint64(len(firstIndexindexSummery)))
				indexEntryS = append(indexEntryS, []byte(firstIndexindexSummery)...)
				indexEntryS = binary.AppendUvarint(indexEntryS, uint64(len(lastKey)))
				indexEntryS = append(indexEntryS, []byte(lastKey)...)
				indexEntryS = binary.AppendUvarint(indexEntryS, indexoffset)
				*IndexSummery = append(*IndexSummery, indexEntryS...)
			}
			break
		}
	}
	//Add to IndexSummery first and last key
	firstKeyEverEncoded := make([]byte, 0)
	firstKeyEverEncoded = binary.AppendUvarint(firstKeyEverEncoded, uint64(len(firtKEyEver)))
	firstKeyEverEncoded = append(firstKeyEverEncoded, []byte(firtKEyEver)...)
	*IndexSummery = append(firstKeyEverEncoded, *IndexSummery...)
	lastKeyEncoded := make([]byte, 0)
	lastKeyEncoded = binary.AppendUvarint(lastKeyEncoded, uint64(len(lastKey)))
	lastKeyEncoded = append(lastKeyEncoded, []byte(lastKey)...)
	*IndexSummery = append(lastKeyEncoded, *IndexSummery...)
	return offset, i
	// Return a default value to avoid missing return error
}

func WriteDataBlockCompresed[T any, H any](file *BlockManagerWrap, itter *utils.Iterator[T, H, *ExportEntry], IndexSummery *[]byte, IndexBloc *[]byte, Filter *BloomFilter, BlockSummery uint64, tree *MerkleTree) (uint64, uint64) {
	entry, ok := itter.Next()
	indexoffset := uint64(0)
	indexTogo := 0
	firtKEyEver := entry.GetKey()
	firstIndexindexSummery := uint64(0)
	lastKey := firtKEyEver
	offset := uint64(0)
	i := uint64(0)
	if !ok {
		fmt.Println("No entries to write.")
		return 0, 0 // No entries to write
	}
	for true {
		i += 1
		if entry == nil {
			continue
		}

		// Create an index entry

		data := entry.toBytes()
		tree.Put(data)
		Filter.Put(utils.Uint64ToBytes(entry.GetKey()))
		file.WriteToFile(data)
		// Write the data to the file
		if firstIndexindexSummery == 0 {
			firstIndexindexSummery = entry.GetKey() // Set firstKey to the first key encountered
		}
		// Create an index entry
		IndexBlockEntry := make([]byte, 0)
		IndexBlockEntry = binary.AppendUvarint(IndexBlockEntry, entry.GetKey())
		IndexBlockEntry = binary.AppendUvarint(IndexBlockEntry, offset)
		*IndexBloc = append(*IndexBloc, IndexBlockEntry...)
		offset += uint64(len(data))
		indexTogo += 1
		if indexTogo%int(BlockSummery) == 0 {
			// Create an index entry
			indexEntry := make([]byte, 0)
			indexEntry = binary.AppendUvarint(indexEntry, uint64(firstIndexindexSummery))
			indexEntry = binary.AppendUvarint(indexEntry, uint64(entry.GetKey()))
			indexEntry = binary.AppendUvarint(indexEntry, indexoffset)
			*IndexSummery = append(*IndexSummery, indexEntry...)
			indexoffset = uint64(len(*IndexBloc))
			firstIndexindexSummery = 0 // Reset firstKey for the next block
		}
		if entry.GetKey() > lastKey {
			lastKey = entry.GetKey()
		}
		// Update the offset
		entry, ok = itter.Next()
		if !ok {
			if firstIndexindexSummery != 0 {
				// If we reach the end of the data and firstKey is still set, add the last entry
				indexEntryS := make([]byte, 0)
				indexEntryS = binary.AppendUvarint(indexEntryS, uint64(firstIndexindexSummery))
				indexEntryS = binary.AppendUvarint(indexEntryS, uint64(lastKey))
				indexEntryS = binary.AppendUvarint(indexEntryS, indexoffset)
				*IndexSummery = append(*IndexSummery, indexEntryS...)
			}
			break
		}
	}
	//Add to IndexSummery first and last key
	firstKeyEverEncoded := make([]byte, 0)
	firstKeyEverEncoded = binary.AppendUvarint(firstKeyEverEncoded, uint64(firtKEyEver))
	*IndexSummery = append(firstKeyEverEncoded, *IndexSummery...)
	lastKeyEncoded := make([]byte, 0)
	lastKeyEncoded = binary.AppendUvarint(lastKeyEncoded, uint64(lastKey))
	*IndexSummery = append(lastKeyEncoded, *IndexSummery...)
	return offset, i
}

func ReadSStabel(pathtoToc string, key string, BlockManager *BlockManager, globalDictionary *GlobalDictionary) (*Entry, bool) {
	file := NewBlockManagerWrap(pathtoToc, BlockManager)
	_, Compresion := file.ReadFromCurrentPosition(1)
	_, oneFile := file.ReadFromCurrentPosition(1)
	file.MoveFromCurrentPosition(8) // Skip the number of entries
	if Compresion[0] == 1 {
		keyEntry, _ := globalDictionary.SkipList.Get(key)
		if keyEntry.GetKey() == "" {
			return &Entry{}, false
		}
		keyUint64 := utils.BytesToUint64(keyEntry.GetValue())
		if oneFile[0] == 1 {
			_, dataOffsetBytes := file.ReadFromCurrentPosition(8)
			_, bloomFilterOffsetBytes := file.ReadFromCurrentPosition(8)
			_, indexSummeryOffsetBytes := file.ReadFromCurrentPosition(8)
			_, indexBlockOffsetBytes := file.ReadFromCurrentPosition(8)
			file.ReadFromCurrentPosition(8)
			// Read the offsets from the file
			dataOffset := utils.BytesToUint64(dataOffsetBytes)
			bloomFilterOffset := utils.BytesToUint64(bloomFilterOffsetBytes)
			indexSummeryOffset := utils.BytesToUint64(indexSummeryOffsetBytes)
			indexBlockOffset := utils.BytesToUint64(indexBlockOffsetBytes)
			file.SetCurrentPosition(int(bloomFilterOffset))
			_, data := file.ReadFromCurrentPosition(int(indexSummeryOffset - bloomFilterOffset)) // Read the data block
			Filter := DeserializeBloomF(data)
			if !Filter.Check(utils.Uint64ToBytes(keyUint64)) {
				return &Entry{}, false
			}

			file.SetCurrentPosition(int(indexSummeryOffset))
			offset, found := ReadIndexSummeryComprested(file, keyUint64)
			if !found {
				return &Entry{}, false
			}
			file.SetCurrentPosition(int(indexBlockOffset + offset))
			offset, found = ReadIndexBlockComprested(file, keyUint64)
			if !found {
				return &Entry{}, false
			}
			file.SetCurrentPosition(int(dataOffset + offset))
			entry := ReadDataCompressed(file)
			data = entry.value
			return &Entry{
				key:     key,
				value:   data,
				deleted: entry.tombstone,
				time:    entry.time,
			}, true
		} else {
			// Read the paths from the TOC file
			m, dataPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
			dataPathLength, n := binary.Uvarint(dataPathBytes)
			file.MoveFromCurrentPosition(int(n) - int(m))
			m, dataPath := file.ReadFromCurrentPosition(int(dataPathLength))
			dataPathString := string(dataPath[:dataPathLength])
			m, bloomFilterPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
			bloomFilterPathLength, n := binary.Uvarint(bloomFilterPathBytes)
			file.MoveFromCurrentPosition(int(n) - int(m))
			m, bloomFilterPath := file.ReadFromCurrentPosition(int(bloomFilterPathLength))
			bloomFilterPathString := string(bloomFilterPath[:bloomFilterPathLength])
			m, indexSummeryPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
			indexSummeryPathLength, n := binary.Uvarint(indexSummeryPathBytes)
			file.MoveFromCurrentPosition(int(n) - int(m))
			m, indexSummeryPath := file.ReadFromCurrentPosition(int(indexSummeryPathLength))
			indexSummeryPathString := string(indexSummeryPath[:indexSummeryPathLength])
			m, indexBlockPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
			indexBlockPathLength, n := binary.Uvarint(indexBlockPathBytes)
			file.MoveFromCurrentPosition(int(n) - int(m))
			m, indexBlockPath := file.ReadFromCurrentPosition(int(indexBlockPathLength))
			indexBlockPathString := string(indexBlockPath[:indexBlockPathLength])
			m, merkleTreePathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
			merkleTreePathLength, n := binary.Uvarint(merkleTreePathBytes)
			file.MoveFromCurrentPosition(int(n) - int(m))
			file.ReadFromCurrentPosition(int(merkleTreePathLength))
			// Read the Bloom filter from the specified path
			bloomFilterFile := NewBlockManagerWrap(bloomFilterPathString, BlockManager)
			m, bloomFilterSizeBytes := bloomFilterFile.ReadFromCurrentPosition(uvarMaxSize)
			bloomFilterSize, n := binary.Uvarint(bloomFilterSizeBytes)
			bloomFilterFile.MoveFromCurrentPosition(int(n) - int(m))
			m, bloomFilterData := bloomFilterFile.ReadFromCurrentPosition(int(bloomFilterSize))
			bloomFilter := DeserializeBloomF(bloomFilterData)
			if !bloomFilter.Check(utils.Uint64ToBytes(keyUint64)) {
				return &Entry{}, false
			}
			// Read the index summary from the specified path
			indexSummeryFile := NewBlockManagerWrap(indexSummeryPathString, BlockManager)
			indexSummeryFile.SetCurrentPosition(0)
			offset, found := ReadIndexSummeryComprested(indexSummeryFile, keyUint64)
			if !found {
				return &Entry{}, false
			}
			// Read the index block from the specified path
			indexBlockFile := NewBlockManagerWrap(indexBlockPathString, BlockManager)
			indexBlockFile.SetCurrentPosition(0)
			offset, found = ReadIndexBlockComprested(indexBlockFile, keyUint64)
			if !found {
				return &Entry{}, false
			}
			// Read the data block from the specified path
			dataFile := NewBlockManagerWrap(dataPathString, BlockManager)
			dataFile.SetCurrentPosition(int(offset))
			entry := ReadDataCompressed(dataFile)
			data := entry.value
			return &Entry{
				key:     key,
				value:   data,
				deleted: entry.tombstone,
				time:    entry.time,
			}, true

		}
	} else {
		if oneFile[0] == 1 {
			_, dataOffsetBytes := file.ReadFromCurrentPosition(8)
			_, bloomFilterOffsetBytes := file.ReadFromCurrentPosition(8)
			_, indexSummeryOffsetBytes := file.ReadFromCurrentPosition(8)
			_, indexBlockOffsetBytes := file.ReadFromCurrentPosition(8)
			file.ReadFromCurrentPosition(8)
			// Read the offsets from the file
			dataOffset := utils.BytesToUint64(dataOffsetBytes)
			bloomFilterOffset := utils.BytesToUint64(bloomFilterOffsetBytes)
			indexSummeryOffset := utils.BytesToUint64(indexSummeryOffsetBytes)
			indexBlockOffset := utils.BytesToUint64(indexBlockOffsetBytes)
			file.SetCurrentPosition(int(bloomFilterOffset))
			_, data := file.ReadFromCurrentPosition(int(indexSummeryOffset - bloomFilterOffset)) // Read the data block
			Filter := DeserializeBloomF(data)
			if !Filter.Check([]byte(key)) {
				return &Entry{}, false
			}
			file.SetCurrentPosition(int(indexSummeryOffset))
			offset, found := ReadIndexSummery(file, key)
			if !found {
				return &Entry{}, false
			}
			file.SetCurrentPosition(int(indexBlockOffset + offset))
			offset, found = ReadIndexBlock(file, key)
			if !found {
				return &Entry{}, false
			}
			file.SetCurrentPosition(int(dataOffset + offset))
			entry := ReadData(file)
			return entry, true
		} else {
			// Read the paths from the TOC file
			m, dataPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
			dataPathLength, n := binary.Uvarint(dataPathBytes)
			file.MoveFromCurrentPosition(int(n) - int(m))
			m, dataPath := file.ReadFromCurrentPosition(int(dataPathLength))
			dataPathString := string(dataPath[:dataPathLength])
			m, bloomFilterPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
			bloomFilterPathLength, n := binary.Uvarint(bloomFilterPathBytes)
			file.MoveFromCurrentPosition(int(n) - int(m))
			m, bloomFilterPath := file.ReadFromCurrentPosition(int(bloomFilterPathLength))
			bloomFilterPathString := string(bloomFilterPath[:bloomFilterPathLength])
			m, indexSummeryPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
			indexSummeryPathLength, n := binary.Uvarint(indexSummeryPathBytes)
			file.MoveFromCurrentPosition(int(n) - int(m))
			m, indexSummeryPath := file.ReadFromCurrentPosition(int(indexSummeryPathLength))
			indexSummeryPathString := string(indexSummeryPath[:indexSummeryPathLength])
			m, indexBlockPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
			indexBlockPathLength, n := binary.Uvarint(indexBlockPathBytes)
			file.MoveFromCurrentPosition(int(n) - int(m))
			m, indexBlockPath := file.ReadFromCurrentPosition(int(indexBlockPathLength))
			indexBlockPathString := string(indexBlockPath[:indexBlockPathLength])
			m, merkleTreePathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
			merkleTreePathLength, n := binary.Uvarint(merkleTreePathBytes)
			file.MoveFromCurrentPosition(int(n) - int(m))
			file.ReadFromCurrentPosition(int(merkleTreePathLength))
			// Read the Bloom filter from the specified path
			bloomFilterFile := NewBlockManagerWrap(bloomFilterPathString, BlockManager)
			m, bloomFilterSizeBytes := bloomFilterFile.ReadFromCurrentPosition(uvarMaxSize)
			bloomFilterSize, n := binary.Uvarint(bloomFilterSizeBytes)
			bloomFilterFile.MoveFromCurrentPosition(int(n) - int(m))
			m, bloomFilterData := bloomFilterFile.ReadFromCurrentPosition(int(bloomFilterSize))
			Filter := DeserializeBloomF(bloomFilterData)
			if !Filter.Check([]byte(key)) {
				return &Entry{}, false
			}
			// Read the index summary from the specified path
			indexSummeryFile := NewBlockManagerWrap(indexSummeryPathString, BlockManager)
			indexSummeryFile.SetCurrentPosition(0)
			offset, found := ReadIndexSummery(indexSummeryFile, key)
			if !found {
				return &Entry{}, false
			}
			// Read the index block from the specified path
			indexBlockFile := NewBlockManagerWrap(indexBlockPathString, BlockManager)
			indexBlockFile.SetCurrentPosition(int(offset))
			offset, found = ReadIndexBlock(indexBlockFile, key)
			if !found {
				return &Entry{}, false
			}
			// Read the data block from the specified path
			dataFile := NewBlockManagerWrap(dataPathString, BlockManager)
			dataFile.SetCurrentPosition(int(offset))
			data := ReadData(dataFile)
			return data, true
		}
	}
}

func ReadIndexSummeryComprested(file *BlockManagerWrap, key uint64) (uint64, bool) {
	n, buffer := file.ReadFromCurrentPosition(uvarMaxSize)
	lastKey, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n)) // Move the file pointer to the next position
	// Move the file pointer to the next position
	n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
	firstKey, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	if key < firstKey || key > lastKey {
		return 0, false
	}
	n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
	firstKeySummery, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
	secondKeySummery, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
	offset, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	// Reading indes summery enyry per entry
	for key > secondKeySummery {
		n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
		firstKeySummery, m = binary.Uvarint(buffer)
		file.MoveFromCurrentPosition(int(m) - int(n))
		n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
		secondKeySummery, m = binary.Uvarint(buffer)
		file.MoveFromCurrentPosition(int(m) - int(n))
		n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
		offset, m = binary.Uvarint(buffer)
		file.MoveFromCurrentPosition(int(m) - int(n))
		if key < firstKeySummery {
			return 0, false
		}
	}
	return offset, true
}
func ReadIndexBlockComprested(file *BlockManagerWrap, key uint64) (uint64, bool) {
	// Read the index block entry
	n, buffer := file.ReadFromCurrentPosition(uvarMaxSize)
	firstKey, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	if key < firstKey {
		return 0, false
	}
	n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
	offset, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	for key >= firstKey {
		if key == firstKey {
			return offset, true
		}

		n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
		if n == 0 {
			break // End of file reached
		}
		firstKey, m = binary.Uvarint(buffer)
		file.MoveFromCurrentPosition(int(m) - int(n))
		n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
		offset, m = binary.Uvarint(buffer)
		file.MoveFromCurrentPosition(int(m) - int(n))
	}
	return 0, false
}
func ReadDataCompressed(file *BlockManagerWrap) *ExportEntry {
	// Read the data block at the specified offset Need to change when i get CRU and Timestamp
	data := make([]byte, 0)
	n, buffer := file.ReadFromCurrentPosition(uvarMaxSize)
	CRU, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
	key, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
	valueSize, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	n, data = file.ReadFromCurrentPosition(int(valueSize - 1))
	n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
	timestamp, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	ExportEntry := &ExportEntry{
		key:       key,
		value:     data,
		tombstone: valueSize == 0,
		time:      int64(timestamp),
	}
	if !ExportEntry.CompareHashes(CRU) {
		fmt.Println("Data integrity check failed for key:", ExportEntry.key)
		return nil
	}
	return ExportEntry
}
func ReadIndexSummery(file *BlockManagerWrap, key string) (uint64, bool) {
	n, buffer := file.ReadFromCurrentPosition(uvarMaxSize)
	lastKeyLenght, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n)) // Move the file pointer to the next position
	// Move the file pointer to the next position
	n, lastKeyBytes := file.ReadFromCurrentPosition(int(lastKeyLenght))
	lastKey := string(lastKeyBytes[:lastKeyLenght])
	n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
	firstKeyLenght, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	n, firstKeyBytes := file.ReadFromCurrentPosition(int(firstKeyLenght))
	firstKey := string(firstKeyBytes)
	if key < firstKey || key > lastKey {
		return 0, false
	}
	n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
	firstKeySummeryLenght, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	//fmt.Println("First Key Summery Length:", firstKeySummeryLenght)
	_, firstKeySummeryBytes := file.ReadFromCurrentPosition(int(firstKeySummeryLenght))
	firstKeySummery := string(firstKeySummeryBytes)
	//fmt.Println("First Key Summery:", firstKeySummery)
	n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
	secondKeySummeryLenght, m := binary.Uvarint(buffer)
	//fmt.Println("Second Key Summery Length:", secondKeySummeryLenght)
	file.MoveFromCurrentPosition(int(m) - int(n))
	_, secondKeySummeryBytes := file.ReadFromCurrentPosition(int(secondKeySummeryLenght))
	secondKeySummery := string(secondKeySummeryBytes)
	//fmt.Println("Second Key Summery:", secondKeySummery)
	n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
	offset, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	// Reading indes summery enyry per

	for key > secondKeySummery {
		//fmt.Println("Key:", firstKey, "Second Key Summery:", secondKeySummery)
		n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
		firstKeySummeryLenght, m := binary.Uvarint(buffer)
		file.MoveFromCurrentPosition(int(m) - int(n))
		_, firstKeySummeryBytes := file.ReadFromCurrentPosition(int(firstKeySummeryLenght))
		firstKeySummery = string(firstKeySummeryBytes)
		n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
		secondKeySummeryLenght, m := binary.Uvarint(buffer)
		file.MoveFromCurrentPosition(int(m) - int(n))
		_, secondKeySummeryBytes := file.ReadFromCurrentPosition(int(secondKeySummeryLenght))
		secondKeySummery = string(secondKeySummeryBytes)
		n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
		offset, m = binary.Uvarint(buffer)
		file.MoveFromCurrentPosition(int(m) - int(n))
		if key < firstKeySummery {
			return 0, false
		}
	}
	return offset, true
}
func ReadIndexBlock(file *BlockManagerWrap, key string) (uint64, bool) {
	// Read the index block entry
	n, buffer := file.ReadFromCurrentPosition(uvarMaxSize)
	firstKeyLenght, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	//fmt.Println("First Key Length:", firstKeyLenght)
	_, firstKeyBytes := file.ReadFromCurrentPosition(int(firstKeyLenght))
	firstKey := string(firstKeyBytes[:firstKeyLenght])
	if key < firstKey {
		return 0, false
	}
	n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
	offset, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	for key >= firstKey {
		if key == firstKey {
			return offset, true
		}

		n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
		if n == 0 {
			break // End of file reached
		}
		firstKeyLenght, m = binary.Uvarint(buffer)
		file.MoveFromCurrentPosition(int(m) - int(n))
		_, firstKeyBytes = file.ReadFromCurrentPosition(int(firstKeyLenght))
		firstKey = string(firstKeyBytes[:firstKeyLenght])
		n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
		offset, m = binary.Uvarint(buffer)
		file.MoveFromCurrentPosition(int(m) - int(n))
	}
	return 0, false
}
func ReadData(file *BlockManagerWrap) *Entry {
	// Read the data block at the specified offset Need to change when i get CRU and Timestamp
	data := make([]byte, 0)
	n, buffer := file.ReadFromCurrentPosition(uvarMaxSize)
	CRU, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
	keySize, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
	valueSize, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	n, buffer = file.ReadFromCurrentPosition(int(keySize))
	key := string(buffer[:keySize])
	if valueSize != 0 {
		n, data = file.ReadFromCurrentPosition(int(valueSize - 1))
	}
	n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
	timestamp, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	// Create an Entry object with the read data
	entry := NewEntryWithTime(key, data, int64(timestamp), valueSize == 0)
	if !entry.CompareHashes(CRU) {
		fmt.Println("Data integrity check failed for key:", entry.GetKey())
		return nil
	}
	return entry
}

func CalculateCRUFromBytes(data []byte) uint64 {
	if len(data) < 8 {
		return 0 // Not enough data to calculate CRU
	}
	cruBytes := data[:8] // Assuming CRU is stored in the first 8 bytes
	cru := utils.BytesToUint64(cruBytes)
	return cru
}

func GetDataFile(data_path string, bm *BlockManager) *BlockManagerWrap {
	file := NewBlockManagerWrap(data_path, bm)
	// Check if the file exists
	file.SetCurrentPosition(1) // Move to the second byte to check for compression
	_, oneFile := file.ReadFromCurrentPosition(1)
	file.MoveFromCurrentPosition(uvarMaxSize) // Skip the number of entries
	if oneFile[0] == 1 {
		file.SetCurrentPosition(58) // Data starts from header at position 58
		return file                 // Return the file
	} else {
		file.SetCurrentPosition(10)
		m, dataPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		dataPathLength, n := binary.Uvarint(dataPathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		m, dataPath := file.ReadFromCurrentPosition(int(dataPathLength))
		dataPathString := string(dataPath[:dataPathLength])
		return NewBlockManagerWrap(dataPathString, bm) // Return the data file
	}
}
func GetIndexFile(data_path string, bm *BlockManager) *BlockManagerWrap {
	file := NewBlockManagerWrap(data_path, bm)
	// Check if the file exists
	file.SetCurrentPosition(1) // Move to the second byte to check for compression
	_, oneFile := file.ReadFromCurrentPosition(1)
	file.SetCurrentPosition(10) // Skip the number of entries
	if oneFile[0] == 1 {
		file.ReadFromCurrentPosition(8)                             // Data
		file.ReadFromCurrentPosition(8)                             // Bloom filter
		file.ReadFromCurrentPosition(8)                             // Index summary
		_, indexBlockOffsetBytes := file.ReadFromCurrentPosition(8) // Index block offset
		file.ReadFromCurrentPosition(8)                             // Merkle tree offset
		// Read the offsets from the files

		indexBlockOffset := utils.BytesToUint64(indexBlockOffsetBytes)
		file.SetCurrentPosition(int(indexBlockOffset)) // Move to the index block offset
		return file                                    // Return the index file
		// Return the file
	} else {
		file.SetCurrentPosition(10)
		m, dataPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		dataPathLength, n := binary.Uvarint(dataPathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		file.ReadFromCurrentPosition(int(dataPathLength))
		m, bloomFilterPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		bloomFilterPathLength, n := binary.Uvarint(bloomFilterPathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		file.ReadFromCurrentPosition(int(bloomFilterPathLength))
		m, indexSummeryPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		indexSummeryPathLength, n := binary.Uvarint(indexSummeryPathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		file.ReadFromCurrentPosition(int(indexSummeryPathLength))
		m, indexBlockPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		indexBlockPathLength, n := binary.Uvarint(indexBlockPathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		m, indexBlockPath := file.ReadFromCurrentPosition(int(indexBlockPathLength))
		indexBlockPathString := string(indexBlockPath[:indexBlockPathLength])
		file.MoveFromCurrentPosition(int(n) - int(m))
		return NewBlockManagerWrap(indexBlockPathString, bm) // Return the index file
	}
}
func GetMerkleTreeFile(data_path string, bm *BlockManager) (*BlockManagerWrap, uint64) {
	file := NewBlockManagerWrap(data_path, bm)
	// Check if the file exists
	file.SetCurrentPosition(1) // Move to the second byte to check for compression
	_, oneFile := file.ReadFromCurrentPosition(1)
	// Skip the number of entries
	file.SetCurrentPosition(10)
	if oneFile[0] == 1 {

		file.ReadFromCurrentPosition(8) //data
		file.ReadFromCurrentPosition(8) // bloom filter
		file.ReadFromCurrentPosition(8) // index summary
		file.ReadFromCurrentPosition(8) //Index
		_, merkleTreeOffsetBytes := file.ReadFromCurrentPosition(8)
		// Read the offsets from the files
		merkleTreeOffset := utils.BytesToUint64(merkleTreeOffsetBytes)
		_, MerkleTreeSize := file.ReadFromCurrentPosition(uvarMaxSize)
		sizeOfMerkleTree := utils.BytesToUint64(MerkleTreeSize)
		file.SetCurrentPosition(int(merkleTreeOffset))   // Move to the merkle tree offset
		return file, sizeOfMerkleTree - merkleTreeOffset // Return the merkle tree file and its size
	} else {
		m, dataPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		dataPathLength, n := binary.Uvarint(dataPathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		file.MoveFromCurrentPosition(int(dataPathLength)) // Skip the data path
		m, bloomFilterPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		bloomFilterPathLength, n := binary.Uvarint(bloomFilterPathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		file.MoveFromCurrentPosition(int(bloomFilterPathLength)) // Skip the bloom filter path
		m, indexSummeryPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		indexSummeryPathLength, n := binary.Uvarint(indexSummeryPathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		file.MoveFromCurrentPosition(int(indexSummeryPathLength)) // Skip the index summary path
		m, indexBlockPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		indexBlockPathLength, n := binary.Uvarint(indexBlockPathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		file.MoveFromCurrentPosition(int(indexBlockPathLength)) // Skip the index block path
		m, merkleTreePathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		merkleTreePathLength, n := binary.Uvarint(merkleTreePathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		m, merkleTreePath := file.ReadFromCurrentPosition(int(merkleTreePathLength))
		merkleTreePathString := string(merkleTreePath[:merkleTreePathLength])
		newFile := NewBlockManagerWrap(merkleTreePathString, bm)
		_, MerkleTreeSize := newFile.ReadFromCurrentPosition(8)
		sizeOfMerkleTree := utils.BytesToUint64(MerkleTreeSize)
		return newFile, sizeOfMerkleTree // Return the merkle tree file and its size

		// Return the data file
	}

}
func ReadInedx(file *BlockManagerWrap) *Entry {
	// Read the index block entry
	n, buffer := file.ReadFromCurrentPosition(uvarMaxSize)
	firstKeyLenght, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	//fmt.Println("First Key Length:", firstKeyLenght)
	_, firstKeyBytes := file.ReadFromCurrentPosition(int(firstKeyLenght))
	firstKey := string(firstKeyBytes[:firstKeyLenght])
	//fmt.Println("First Key:", firstKey)
	n, buffer = file.ReadFromCurrentPosition(uvarMaxSize)
	offset, m := binary.Uvarint(buffer)
	file.MoveFromCurrentPosition(int(m) - int(n))
	//fmt.Println("Offset:", offset)
	Entry := &Entry{
		key:     firstKey,
		value:   utils.Uint64ToBytes(offset), // Value is not used in index entries
		deleted: false,                       // Tombstone is not used in index entries
		time:    0,                           // Time is not used in index entries
	}
	return Entry
}
func GetListInex(file *BlockManagerWrap, MaxEntrys uint64) *auxiliary.ListWithIterator[Entry] {
	list := auxiliary.NewListWithIterator[Entry]()
	i := uint64(0)
	for i < MaxEntrys {
		Entry := ReadInedx(file)
		list.Add(Entry)
		i++
	}
	return list
}
func RemoveSStabel(path string, bm *BlockManager) {
	file := NewBlockManagerWrap(path, bm)
	// Check if the file exists
	file.SetCurrentPosition(1)
	_, oneFile := file.ReadFromCurrentPosition(1)
	if oneFile[0] == 1 {
		// Remove the file from the block manager
		//fmt.Println("")
		file.FlushBuffer(false)
		os.Remove(path)
		// Remove the file from the filesystem
	} else {
		file.SetCurrentPosition(10)
		m, dataPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		dataPathLength, n := binary.Uvarint(dataPathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		m, dataPath := file.ReadFromCurrentPosition(int(dataPathLength))
		dataPathString := string(dataPath[:dataPathLength])
		os.Remove(dataPathString) // Remove the data file from the filesystem
		m, bloomFilterPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		bloomFilterPathLength, n := binary.Uvarint(bloomFilterPathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		m, bloomFilterPath := file.ReadFromCurrentPosition(int(bloomFilterPathLength))
		bloomFilterPathString := string(bloomFilterPath[:bloomFilterPathLength])
		os.Remove(bloomFilterPathString) // Remove the bloom filter file from the filesystem
		m, indexSummeryPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		indexSummeryPathLength, n := binary.Uvarint(indexSummeryPathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		m, indexSummeryPath := file.ReadFromCurrentPosition(int(indexSummeryPathLength))
		indexSummeryPathString := string(indexSummeryPath[:indexSummeryPathLength])
		os.Remove(indexSummeryPathString) // Remove the index summary file from the filesystem
		m, indexBlockPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		indexBlockPathLength, n := binary.Uvarint(indexBlockPathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		m, indexBlockPath := file.ReadFromCurrentPosition(int(indexBlockPathLength))
		indexBlockPathString := string(indexBlockPath[:indexBlockPathLength])
		os.Remove(indexBlockPathString) // Remove the index block file from the filesystem
		m, merkleTreePathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		merkleTreePathLength, n := binary.Uvarint(merkleTreePathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		m, merkleTreePath := file.ReadFromCurrentPosition(int(merkleTreePathLength))
		merkleTreePathString := string(merkleTreePath[:merkleTreePathLength])
		os.Remove(merkleTreePathString) // Remove the merkle tree file from the filesystem
		os.Remove(path)                 // Remove the main SSTable file from the filesystem
	}

}
func GetFirstAndLAstKeyCompresed(path string, bm *BlockManager) (string, string) {
	file := NewBlockManagerWrap(path, bm)
	var file2 *BlockManagerWrap
	// Check if the file exists
	file.SetCurrentPosition(1) // Move to the second byte to check for compression
	_, oneFile := file.ReadFromCurrentPosition(1)
	file.MoveFromCurrentPosition(uvarMaxSize) // Skip the number of entries
	if oneFile[0] == 1 {
		file.SetCurrentPosition(50)
		_, merkleTreeOffset := file.ReadFromCurrentPosition(8)
		file.SetCurrentPosition(int(utils.BytesToUint64(merkleTreeOffset)))
		file2 = file // Use the same file for reading first and last keys
		// Data starts from header at position 58
	} else {
		file.SetCurrentPosition(10)
		m, dataPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		dataPathLength, n := binary.Uvarint(dataPathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		m, _ = file.ReadFromCurrentPosition(int(dataPathLength))
		m, bloomFilterPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		bloomFilterPathLength, n := binary.Uvarint(bloomFilterPathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		m, _ = file.ReadFromCurrentPosition(int(bloomFilterPathLength))
		m, indexSummeryPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		indexSummeryPathLength, n := binary.Uvarint(indexSummeryPathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		m, _ = file.ReadFromCurrentPosition(int(indexSummeryPathLength))
		m, indexBlockPathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		indexBlockPathLength, n := binary.Uvarint(indexBlockPathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		m, _ = file.ReadFromCurrentPosition(int(indexBlockPathLength))
		m, merkleTreePathBytes := file.ReadFromCurrentPosition(uvarMaxSize)
		merkleTreePathLength, n := binary.Uvarint(merkleTreePathBytes)
		file.MoveFromCurrentPosition(int(n) - int(m))
		m, _ = file.ReadFromCurrentPosition(int(merkleTreePathLength))
		file2 = file // Use the same file for reading first and last keys

	}
	_, buffer := file2.ReadFromCurrentPosition(8)
	firstKeySize := utils.BytesToUint64(buffer)
	_, firstKeyBytes := file.ReadFromCurrentPosition(int(firstKeySize))
	firstKey := string(firstKeyBytes)
	// Move the file pointer to the next position
	// Move the file pointer to the next position
	_, buffer = file2.ReadFromCurrentPosition(8)
	lastKeySize := utils.BytesToUint64(buffer)
	_, lastKeyBytes := file.ReadFromCurrentPosition(int(lastKeySize))
	lastKey := string(lastKeyBytes)
	return firstKey, lastKey
}
func CheckSStable(path string, bm *BlockManager) bool {

	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	file := NewBlockManagerWrap(path, bm)
	file.SetCurrentPosition(1)
	_, oneFile := file.ReadFromCurrentPosition(1)
	if oneFile[0] == 1 {
		return true // If the first byte is 1, it indicates a single file SSTable
	} else {
		i := 0
		file.SetCurrentPosition(10)
		for i < 5 {
			m, fileBytesSize := file.ReadFromCurrentPosition(uvarMaxSize)
			fileSize, n := binary.Uvarint(fileBytesSize)
			file.MoveFromCurrentPosition(int(n) - int(m))
			_, filePath := file.ReadFromCurrentPosition(int(fileSize))
			filePathString := string(filePath[:fileSize])
			if _, err := os.Stat(filePathString); os.IsNotExist(err) {
				return false
			}
			i++
		}
	}
	return true
}
func ValidateSStabel(path string, bm *BlockManager) (bool, int) {
	file, size := GetMerkleTreeFile(path, bm)
	if size == 0 {
		fmt.Println("Merkle tree size is zero, SSTable is invalid.")
		return false, -1
	}
	// Read the Merkle tree from the file
	_, merkleTreeData := file.ReadFromCurrentPosition(int(size))
	// Deserialize the Merkle tree
	merkleTree := DeserializeMerkle(merkleTreeData)
	merkle2 := NewMerkleTree()
	file2 := NewBlockManagerWrap(path, bm)
	// Check if the Merkle tree is valid
	file2.SetCurrentPosition(2)
	_, NeumberOfEntrys := file2.ReadFromCurrentPosition(8)
	file3 := GetDataFile(path, bm)
	for i := uint64(0); i < utils.BytesToUint64(NeumberOfEntrys); i++ {
		entry := ReadData(file3)
		merkle2.Put(entry.toBytes())
	}
	merkle2.build()
	isOk, idx := CompareTreesWithIndex(merkleTree, merkle2)
	if !isOk {
		return false, idx
	}
	return true, 0
}
