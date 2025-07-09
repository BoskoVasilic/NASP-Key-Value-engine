package structures

import (
	utils "KVengine/utils"
	"encoding/binary"
	"fmt"
	"os"
)

type SSTable[T any] struct {
	Podaci          *utils.Iterator[T, *Entry] // Iterator over the data in the SSTable
	Filter          *BloomFilter
	numberOfEnterys int
	oneFile         bool // Indicates if the SSTable is stored in a single file
	SumeeryIndex    uint64
	BlockManager    *BlockManager
}

func NewSSTable[T any](podaci *utils.Iterator[T, *Entry], NumberOfEnyerys int, SumeeryIndex uint64, oneFile bool, BlocMenager *BlockManager) *SSTable[T] {
	return &SSTable[T]{
		Podaci:          podaci,
		Filter:          NewBloomFilterWithParams(NumberOfEnyerys, 0.01), // Example parameters for the Bloom filter
		SumeeryIndex:    SumeeryIndex,
		numberOfEnterys: NumberOfEnyerys,
		oneFile:         oneFile,
		BlockManager:    BlocMenager,
	}
}

func (s *SSTable[T]) GenerateSSTabel() {
	// Create a file to store the SSTable data

	// Ovde cu da radim pretvaranje
	podatak, ok := s.Podaci.Next()
	firstKey := "NoKey"
	IndexBlock := make([]byte, 0, 0)
	IndexSummery := make([]byte, 0, 0)
	offset := uint64(0)
	indexoffset := uint64(0)
	index := 0
	firtKEyEver := podatak.GetKey()
	lastKey := firtKEyEver // Initialize lastKey with the first key
	for ok {
		if podatak == nil {
			break
		}
		s.Filter.Put([]byte(podatak.GetKey()))
		key := podatak.GetKey()
		if firstKey == "NoKey" {
			firstKey = key
		}
		//Create an block
		IndexBlockEntry := make([]byte, 0)
		IndexBlockEntry = binary.AppendUvarint(IndexBlockEntry, uint64(len(key)))
		IndexBlockEntry = append(IndexBlockEntry, []byte(key)...)
		IndexBlockEntry = binary.AppendUvarint(IndexBlockEntry, uint64(offset))
		fmt.Println(IndexBlockEntry)
		index += 1
		IndexBlock = append(IndexBlock, IndexBlockEntry...)
		offset += podatak.GetSize()
		// Creating IndexSummery
		if index%int(s.SumeeryIndex) == 0 {
			// Create an index entry
			indexEntry := make([]byte, 0)
			indexEntry = binary.AppendUvarint(indexEntry, uint64(len(firstKey)))
			indexEntry = append(indexEntry, []byte(firstKey)...)
			indexEntry = binary.AppendUvarint(indexEntry, uint64(len(lastKey)))
			indexEntry = append(indexEntry, []byte(key)...)
			indexEntry = binary.AppendUvarint(indexEntry, uint64(indexoffset))
			IndexSummery = append(IndexSummery, indexEntry...)
			indexoffset = uint64(len(IndexBlock))
			firstKey = "NoKey" // Reset firstKey for the next block
		}
		podatak, ok = s.Podaci.Next()
		if podatak != nil {
			lastKey = podatak.GetKey()
		}
		if !ok && firstKey != "NoKey" {
			// If we reach the end of the data and firstKey is still set, add the last entry
			indexEntryS := make([]byte, 0)
			indexEntryS = binary.AppendUvarint(indexEntryS, uint64(len(firstKey)))
			indexEntryS = append(indexEntryS, []byte(firstKey)...)
			indexEntryS = binary.AppendUvarint(indexEntryS, uint64(len(lastKey)))
			indexEntryS = append(indexEntryS, []byte(lastKey)...)
			indexEntryS = binary.AppendUvarint(indexEntryS, uint64(indexoffset))
			IndexSummery = append(IndexSummery, indexEntryS...)
		}

	}

	// Write the data block to the file
	if s.oneFile {
		file := NewBlockManagerWrap("SSTable.bin", s.BlockManager)
		defer file.Close()
		offsetFile := uint64(0)
		// Write in file
		_, err = file.Write([]byte{1})
		n, err := file.Write(binary.AppendUvarint(nil, uint64(len(firtKEyEver))))
		offsetFile += uint64(n)
		n, err = file.Write([]byte(firtKEyEver))
		offsetFile += uint64(n)
		n, err = file.Write(binary.AppendUvarint(nil, uint64(len(lastKey))))
		offsetFile += uint64(n)
		n, err = file.Write([]byte(lastKey))
		offsetFile += uint64(n)
		BFilter := s.Filter.SerializeBloomF()
		offsetFile += uint64(8 * 4)
		offsetFile += uint64(1)
		_, err = file.Write(utils.Uint64ToBytes(offsetFile))
		_, err = file.Write(utils.Uint64ToBytes(offsetFile + uint64(len(BFilter))))
		_, err = file.Write(utils.Uint64ToBytes(offsetFile + uint64(len(BFilter)) + uint64(len(IndexSummery))))
		_, err = file.Write(utils.Uint64ToBytes(offsetFile + uint64(len(BFilter)) + uint64(len(IndexSummery)) + uint64(len(IndexBlock))))
		_, err = file.Write(BFilter)
		fmt.Println("Index Sumery:", IndexSummery)
		fmt.Println("Index Block:", IndexBlock)
		_, err = file.Write(IndexSummery)
		_, err = file.Write(IndexBlock)
		// Placeholder for the data block
		s.Podaci.Reset()
		for true {
			podatak, _ := s.Podaci.Next()
			if podatak == nil {
				break
			}
			file.Write(podatak.toBytes())
		}
		file.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0})
	} else {
		// Handle the case where oneFile is false (if needed)
		file, err := os.Create("SSTableTOC.bin")
		if err != nil {
			panic(err)
		}
		defer file.Close()
		fileFilterPath := "SSTableFilter.bin"
		fileFilter, err := os.Create(fileFilterPath)
		if err != nil {
			panic(err)
		}
		defer fileFilter.Close()
		FileIndexSumeeryPath := "SSTableIndexSummery.bin"
		fileIndexSummery, err := os.Create(FileIndexSumeeryPath)
		if err != nil {
			panic(err)
		}
		defer fileIndexSummery.Close()
		fileIndexPath := "SSTableIndex.bin"
		fileIndex, err := os.Create(fileIndexPath)
		if err != nil {
			panic(err)
		}
		defer fileIndex.Close()
		fileDataPath := "SSTableData.bin"
		fileData, err := os.Create(fileDataPath)
		if err != nil {
			panic(err)
		}
		defer fileData.Close()
		offsetFile := uint64(0)
		// Write in file
		_, err = file.Write([]byte{0})
		_, err = file.Write(utils.Uint64ToBytes(uint64(len(firstKey))))
		offsetFile += uint64(len(firstKey)) + 8 // +1 for the type byte
		_, err = file.Write([]byte(firstKey))
		_, err = file.Write(utils.Uint64ToBytes(uint64(len(lastKey))))
		_, err = file.Write([]byte(lastKey))
		offsetFile += uint64(len(lastKey)) + 8 // +8 for the length of the last key
		_, err = file.Write(utils.Uint64ToBytes(uint64(len(fileFilterPath))))
		_, err = file.Write([]byte(fileFilterPath))
		_, err = file.Write(utils.Uint64ToBytes(uint64(len(FileIndexSumeeryPath))))
		_, err = file.Write([]byte(FileIndexSumeeryPath))
		_, err = file.Write(utils.Uint64ToBytes(uint64(len(fileIndexPath))))
		_, err = file.Write([]byte(fileIndexPath))
		_, err = file.Write(utils.Uint64ToBytes(uint64(len(fileDataPath))))
		_, err = file.Write([]byte(fileDataPath))

		BFilter := s.Filter.SerializeBloomF()
		_, err = fileFilter.Write(BFilter)
		_, err = fileIndexSummery.Write(IndexSummery)
		_, err = fileIndex.Write(IndexBlock)
		for true {
			podatak, _ := s.Podaci.Next()
			if podatak == nil {
				break
			}
			fileData.Write(podatak.toBytes())
		}
	}
	// Reset the iterator to the beginning
}
func ReadSStabel(pathtoToc string, key string) {
	file, err := os.Open(pathtoToc)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Read the first byte to determine if it's a single file or multiple files
	var fileType byte
	err = binary.Read(file, binary.LittleEndian, &fileType)
	if err != nil {
		panic(err)
	}

	if fileType == 1 {
		// Handle single file case
		fmt.Println("Single file SSTable detected.")

		data := make([]byte, 8)
		_, err := file.Read(data)
		if err != nil {
			panic(err)
		}
		firstKeyLength, n := binary.Uvarint(data)
		file.Seek(int64(n)-8, 1) // Move the file pointer to the next position
		fmt.Println("First Key Length:", firstKeyLength)
		data = make([]byte, firstKeyLength)
		file.Read(data)
		firstKey := string(data)
		fmt.Println("First Key:", firstKey)
		data = make([]byte, 8)
		_, err = file.Read(data)
		if err != nil {
			panic(err)
		}
		lastKeyLength, n := binary.Uvarint(data)
		file.Seek(int64(n)-8, 1) // Move the file pointer to the next position
		fmt.Println("Last Key Length:", lastKeyLength)
		data = make([]byte, lastKeyLength)
		file.Read(data)
		lastKey := string(data)
		fmt.Println("Last Key:", lastKey)
		// Read the Bloom filter
		data = make([]byte, 8)
		_, err = file.Read(data)
		if err != nil {
			panic(err)
		}
		OffsetBlumFilte := utils.BytesToUint64(data) // Move the file pointer to the next position
		fmt.Println("Offset Bloom Filter:", OffsetBlumFilte)
		data = make([]byte, 8)
		_, err = file.Read(data)
		if err != nil {
			panic(err)
		}
		OffsetIndexSummery := utils.BytesToUint64(data) // Move the file pointer to the next position
		fmt.Println("Offset Index Summery:", OffsetIndexSummery)
		data = make([]byte, 8)
		_, err = file.Read(data)
		if err != nil {
			panic(err)
		}
		OffsetIndex := utils.BytesToUint64(data) // Move the file pointer to the next position
		fmt.Println("Offset Index:", OffsetIndex)
		data = make([]byte, 8)
		_, err = file.Read(data)
		if err != nil {
			panic(err)
		}
		OffsetData := utils.BytesToUint64(data) // Move the file pointer to the next position
		fmt.Println("Offset Data:", OffsetData)
		// Read the Bloom filter data
		BytesBlum := make([]byte, OffsetIndexSummery-OffsetBlumFilte)
		file.Read(BytesBlum)
		blumFilter := DeserializeBloomF(BytesBlum)
		found := blumFilter.Check([]byte(key))
		fmt.Println("File found", found)
		//Loading Index Summery
		file.Seek(int64(OffsetIndexSummery), 0) // Move to the index summary offset
		FirstKey, SecondKey, IndexOffset := GetInexSumeeryEntry(file)
		for key > SecondKey {
			// Read the next index summary entry
			FirstKey, SecondKey, IndexOffset = GetInexSumeeryEntry(file)
			fmt.Println("First Key:", FirstKey, "Second Key:", SecondKey, "Index Offset:", IndexOffset)
		}
		if key < FirstKey {
			fmt.Println("Key is less than the first key in the index summary.")
		}
		// Read the index block
		fmt.Println(IndexOffset)

		KeyData, DataOFfset := GetIndexBlockEntry(file)
		KeyData, DataOFfset = GetIndexBlockEntry(file)
		KeyData, DataOFfset = GetIndexBlockEntry(file)
		KeyData, DataOFfset = GetIndexBlockEntry(file)
		KeyData, DataOFfset = GetIndexBlockEntry(file)

		fmt.Println("Key Data:", KeyData, "Data Offset:", DataOFfset)
	} else {
		// Handle multiple files case
		fmt.Println("Multiple files SSTable detected.")
	}

}
func GetInexSumeeryEntry(file *os.File) (string, string, uint64) {
	// Implement logic to read the index summary entry from the file
	// This function should return the first key, last key, and index offset
	data := make([]byte, 8)
	_, err := file.Read(data)
	if err != nil {
		panic(err)
	}
	firstKeyLength, n := binary.Uvarint(data)
	file.Seek(int64(n)-8, 1) // Move the file pointer to the next position\
	firstKeyData := make([]byte, firstKeyLength)
	_, err = file.Read(firstKeyData)
	if err != nil {
		panic(err)
	}
	firstKey := string(firstKeyData)
	data = make([]byte, 8)
	_, err = file.Read(data)
	if err != nil {
		panic(err)
	}
	lastKeyLength, n := binary.Uvarint(data)
	file.Seek(int64(n)-8, 1) // Move the file pointer to the next position
	lastKeyData := make([]byte, lastKeyLength)
	_, err = file.Read(lastKeyData)
	if err != nil {
		panic(err)
	}
	lastKey := string(lastKeyData)
	data = make([]byte, 8)
	_, err = file.Read(data)
	if err != nil {
		panic(err)
	}
	indexOffset, n := binary.Uvarint(data)
	file.Seek(int64(n)-8, 1) // Move the file pointer to the next position
	return firstKey, lastKey, indexOffset
}
func GetIndexBlockEntry(file *os.File) (string, uint64) {
	data := make([]byte, 8)
	_, err := file.Read(data)
	if err != nil {
		panic(err)
	}
	//fmt.Println("Data Read:", data)
	firstKeyLength, n := binary.Uvarint(data)
	//fmt.Print("First Key Length:", firstKeyLength, " ", int64(n)-8, " ")
	// Ovaj f.Seek samo prestaje da radi
	file.Seek(int64(n)-8, 1) // Move the file pointer to the next position
	file.Read(data)
	//fmt.Println("Data Length:", len(data), " ", data)
	file.Seek(-8, 1)

	firstKeyData := make([]byte, firstKeyLength)
	_, err = file.Read(firstKeyData)
	if err != nil {
		panic(err)
	}
	firstKey := string(firstKeyData)

	data = make([]byte, 8)
	_, err = file.Read(data)
	if err != nil {
		panic(err)
	}
	offset, n := binary.Uvarint(data)
	//fmt.Println("Offset:", offset, " ", n-8)
	file.Seek(int64(n)-8, 1) // Move the file pointer to the next position

	return firstKey, offset
}
func GetDataBlock(file *os.File) string {
	return ""
}
