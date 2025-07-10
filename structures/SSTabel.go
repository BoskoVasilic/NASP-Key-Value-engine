package structures

import (
	utils "KVengine/utils"
	"encoding/binary"
	"fmt"
	"os"
)

type SSTable[T any, H any] struct {
	Podaci           *utils.Iterator[T, H, *Entry] // Iterator over the data in the SSTable
	Filter           *BloomFilter
	numberOfEnterys  int
	oneFile          bool // Indicates if the SSTable is stored in a single file
	SumeeryIndex     uint64
	BlockManager     *BlockManager
	globalDictionary *GlobalDictionary
}

func NewSSTable[T any, H any](podaci *utils.Iterator[T, H, *Entry], NumberOfEnyerys int, SumeeryIndex uint64, oneFile bool, BlocMenager *BlockManager, dictinery *GlobalDictionary) *SSTable[T, H] {
	return &SSTable[T, H]{
		Podaci:           podaci,
		Filter:           NewBloomFilterWithParams(NumberOfEnyerys, 0.01), // Example parameters for the Bloom filter
		SumeeryIndex:     SumeeryIndex,
		numberOfEnterys:  NumberOfEnyerys,
		oneFile:          oneFile,
		BlockManager:     BlocMenager,
		globalDictionary: dictinery,
	}
}

func (s *SSTable[T, H]) GenerateSSTabel() {
	// Create a file to store the SSTable data

	// Ovde cu da radim pretvaranje)
	data := AddItems(s.globalDictionary, s.Podaci)
	podatak, ok := data.Next()
	firstKey := -1 // Initialize firstKey to -1 to indicate it hasn't been set yet
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
		s.Filter.Put(utils.Uint64ToBytes(podatak.GetKey()))
		key := podatak.GetKey()
		if firstKey == -1 {
			firstKey = int(podatak.GetKey()) // Set firstKey to the first key encountered
		}
		//Create an block
		IndexBlockEntry := make([]byte, 0)
		IndexBlockEntry = binary.AppendUvarint(IndexBlockEntry, key)
		IndexBlockEntry = binary.AppendUvarint(IndexBlockEntry, uint64(offset))
		fmt.Println(IndexBlockEntry)
		index += 1
		IndexBlock = append(IndexBlock, IndexBlockEntry...)
		offset += podatak.GetSize()
		// Creating IndexSummery
		if index%int(s.SumeeryIndex) == 0 {
			// Create an index entry
			indexEntry := make([]byte, 0)
			indexEntry = binary.AppendUvarint(indexEntry, uint64(firstKey))
			indexEntry = binary.AppendUvarint(indexEntry, key)
			indexEntry = binary.AppendUvarint(indexEntry, uint64(indexoffset))
			IndexSummery = append(IndexSummery, indexEntry...)
			indexoffset = uint64(len(IndexBlock))
			firstKey = -1 // Reset firstKey for the next block
		}
		podatak, ok = data.Next()
		if podatak != nil {
			lastKey = podatak.GetKey()
		}
		if !ok && firstKey != -1 {
			// If we reach the end of the data and firstKey is still set, add the last entry
			indexEntryS := make([]byte, 0)
			indexEntryS = binary.AppendUvarint(indexEntryS, uint64(firstKey))
			indexEntryS = binary.AppendUvarint(indexEntryS, uint64(lastKey))
			indexEntryS = binary.AppendUvarint(indexEntryS, uint64(indexoffset))
			IndexSummery = append(IndexSummery, indexEntryS...)
		}

	}

	// Write the data block to the file
	if s.oneFile {
		file := NewBlockManagerWrap("SSTable.bin", s.BlockManager)
		defer file.FlushBuffer()
		offsetFile := uint64(0)
		// Write in file
		file.WriteToFile([]byte{1})
		n, _ := file.WriteToFile(binary.AppendUvarint(nil, uint64(firtKEyEver)))
		offsetFile += uint64(n)
		n, _ = file.WriteToFile(binary.AppendUvarint(nil, uint64(lastKey)))
		offsetFile += uint64(n)
		BFilter := s.Filter.SerializeBloomF()
		offsetFile += uint64(8 * 4)
		offsetFile += uint64(1)
		file.WriteToFile(utils.Uint64ToBytes(offsetFile))
		file.WriteToFile(utils.Uint64ToBytes(offsetFile + uint64(len(BFilter))))
		file.WriteToFile(utils.Uint64ToBytes(offsetFile + uint64(len(BFilter)) + uint64(len(IndexSummery))))
		file.WriteToFile(utils.Uint64ToBytes(offsetFile + uint64(len(BFilter)) + uint64(len(IndexSummery)) + uint64(len(IndexBlock))))
		//file.WriteToFile(BFilter)
		fmt.Println("Index Sumery:", IndexSummery)
		fmt.Println("Index Block:", IndexBlock)
		file.WriteToFile(IndexSummery)
		file.WriteToFile(IndexBlock)
		// Placeholder for the data block
		s.Podaci.Reset()
		for true {
			podatak, _ := s.Podaci.Next()
			if podatak == nil {
				break
			}
			file.WriteToFile(podatak.toBytes())
		}
	} else {
		// Handle the case where oneFile is false (if needed)
		/*file, err := os.Create("SSTableTOC.bin")
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
		}*/
	}
	// Reset the iterator to the beginning
}
func WriteDataBlockCompresed[T any, H any](file *BlockManagerWrap, itter *utils.Iterator[T, H, *ExportEntry]) {

}
func WriteDataBlock() {

}
func ReadSStabel(pathtoToc string, key string, BlockManager *BlockManager, globalDictionary *GlobalDictionary) {

	file := NewBlockManagerWrap(pathtoToc, BlockManager)
	// Read the first byte to determine if it's a single file or multiple files
	// Read the first byte to determine if it's a single file or multiple files
	_, fileTypeData := file.ReadFromCurrentPosition(1)
	fmt.Println("File Type Data:", fileTypeData)
	i := 0
	for i < 110 {
		n, fileTypeData := file.ReadFromCurrentPosition(1)
		if n == 0 {
			break // End of file reached
		}
		fmt.Println("File Type Data:", fileTypeData, " ")
		i++
	}
	if fileTypeData[0] == 1 {
		// Handle single file case
		fmt.Println("Single file SSTable detected.")
		/*
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
		*/
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
