package structures

import (
	"encoding/binary"
	"io"
	"log"
	"math"
	"os"
)

const (
	BLOCK_SIZE        = 8000 // specify the block size in config file
	BLOCK_HEADER_SIZE = 4
)

type RecordType uint8

const (
	FULL RecordType = iota
	FIRST
	MIDDLE
	LAST
)

type BlockManager struct {
	cache *BlockCache
}

func NewBlockManager(cacheSize int) *BlockManager {
	return &BlockManager{
		cache: NewBlockCache(cacheSize),
	}
}

func openFile(filePath string, openFlag int) *os.File {
	file, err := os.OpenFile(filePath, openFlag, 0666)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
}

func (bm *BlockManager) ReadBlock(filePath string, blockNumber int) ([]byte, error) {
	if data, ok := bm.cache.Get(filePath, uint32(blockNumber*BLOCK_SIZE)); ok {
		return data, nil
	}

	file := openFile(filePath, os.O_RDONLY)
	defer closeFile(file)

	_, err := file.Seek(int64(blockNumber*BLOCK_SIZE), 0)
	if err != nil {
		log.Fatal(err)
	}

	data := make([]byte, BLOCK_SIZE)

	_, err = file.Read(data)
	if err != nil {
		log.Printf("Error reading file: %v", err)
	}

	bm.cache.Put(filePath, uint32(blockNumber*BLOCK_SIZE), data)

	return data, err
}

func (bm *BlockManager) WriteBlock(filePath string, blockNumber int, data []byte) ([]byte, error) {
	block, _ := bm.ReadBlock(filePath, blockNumber)

	currentLength := bm.GetCurrentLength(block)
	remainingSpace := BLOCK_SIZE - currentLength

	lastBlock := make([]byte, 0)

	if currentLength+len(data) > remainingSpace {
		block, lastBlock = bm.AddDataWithBiggerSizeToBlock(currentLength, remainingSpace, data, block)
	} else if currentLength+len(data) <= remainingSpace {
		block, _ = bm.AddDataWithCorrectSizeToBlock(currentLength, data, block)
	}

	err := bm.WriteToDisk(filePath, blockNumber, block)

	return lastBlock, err
}

func (bm *BlockManager) GetCurrentLength(data []byte) int {
	return int(binary.LittleEndian.Uint32(data[BLOCK_SIZE-BLOCK_HEADER_SIZE:]))
}

func (bm *BlockManager) AddDataWithCorrectSizeToBlock(currentLength int, data []byte, block []byte) ([]byte, int) {
	copy(block[currentLength:], data)
	newLength := currentLength + len(data)
	var additionalBlock []byte
	if BLOCK_SIZE >= newLength+BLOCK_HEADER_SIZE {
		binary.LittleEndian.PutUint32(block[BLOCK_SIZE-BLOCK_HEADER_SIZE:], uint32(newLength))
	} else {
		additionalBlock = make([]byte, BLOCK_SIZE)
		binary.LittleEndian.PutUint32(additionalBlock[BLOCK_SIZE-BLOCK_HEADER_SIZE:], uint32(0))
	}
	return block, newLength
}

func (bm *BlockManager) AddDataWithEqualSizeToBlock(currentLength int, data []byte, block []byte) ([]byte, []byte) {
	copy(block[currentLength:], data)
	newBlock := make([]byte, BLOCK_SIZE)
	binary.LittleEndian.PutUint32(newBlock[BLOCK_SIZE-BLOCK_HEADER_SIZE:], uint32(0))
	return block, newBlock
}

func (bm *BlockManager) AddDataWithBiggerSizeToBlock(currentLength int, remainingSpace int, data []byte, block []byte) ([]byte, []byte) {
	newBlock := make([]byte, BLOCK_SIZE)
	copy(block[currentLength:], make([]byte, remainingSpace))
	allBlocksToWrite := make([]byte, 0)
	additionalBlock := make([]byte, 0)
	allBlocksToWrite = append(allBlocksToWrite, block...)
	lastBlock := make([]byte, 0)
	if len(data) > BLOCK_SIZE {
		numBlocks := math.Ceil(float64(len(data)) / float64(BLOCK_SIZE))
		for i := 0; i < int(numBlocks); i++ {
			start := i * BLOCK_SIZE
			end := start + BLOCK_SIZE
			if end > len(data) {
				end = len(data)
			}
			blockData := data[start:end]

			copy(newBlock, blockData)

			if i == int(numBlocks)-1 {
				newLength := len(blockData)
				if newLength+BLOCK_HEADER_SIZE <= BLOCK_SIZE {
					binary.LittleEndian.PutUint32(newBlock[BLOCK_SIZE-BLOCK_HEADER_SIZE:], uint32(newLength))
					lastBlock = newBlock
				} else {
					additionalBlock = make([]byte, BLOCK_SIZE)
					binary.LittleEndian.PutUint32(newBlock[BLOCK_SIZE-BLOCK_HEADER_SIZE:], uint32(0))
				}
			}

			allBlocksToWrite = append(allBlocksToWrite, newBlock...)
			if len(additionalBlock) > 0 {
				allBlocksToWrite = append(allBlocksToWrite, additionalBlock...)
			}

		}
		block = allBlocksToWrite
	} else {
		copy(newBlock, data)
		newLength := len(data)
		binary.LittleEndian.PutUint32(newBlock[BLOCK_SIZE-BLOCK_HEADER_SIZE:], uint32(newLength))
		allBlocksToWrite = append(allBlocksToWrite, newBlock...)
		block = allBlocksToWrite
		lastBlock = newBlock
	}
	return block, lastBlock
}

func (bm *BlockManager) WriteToDisk(filePath string, blockNumber int, block []byte) error {
	file := openFile(filePath, os.O_RDWR)
	defer closeFile(file)

	_, err := file.Seek(int64(blockNumber*BLOCK_SIZE), 0)
	if err != nil {
		log.Fatal(err)
	}

	_, err = file.Write(block)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(block); i += BLOCK_SIZE {
		blockData := block[i : i+BLOCK_SIZE]
		bm.cache.Put(filePath, uint32((blockNumber+i)*BLOCK_SIZE), blockData)
	}

	return nil
}

func (bm *BlockManager) OverwriteOnDisk(filePath string, offset int, data []byte) error {
	file := openFile(filePath, os.O_RDWR)
	defer closeFile(file)

	_, err := file.Seek(int64(offset), 0)
	if err != nil {
		log.Fatal(err)
	}

	_, err = file.Write(data)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (bm *BlockManager) ReadBytesByOffset(filePath string, offset int, bytesNeeded int) (int, []byte) {
	blockNumber := offset / BLOCK_SIZE
	startInBlock := offset % BLOCK_SIZE

	allBlocksToReturn := make([]byte, 0)
	runWhile := bytesNeeded

	successfullyRead := 0

	for runWhile > 0 {
		block, err := bm.ReadBlock(filePath, blockNumber)
		if err != nil {
			if err == io.EOF {
				successfullyRead = len(allBlocksToReturn[startInBlock:])
				allBlocksToReturn = append(allBlocksToReturn, make([]byte, bytesNeeded-successfullyRead)...)
				break
			}
		}
		blockNumber++
		allBlocksToReturn = append(allBlocksToReturn, block...)
		runWhile -= BLOCK_SIZE
	}

	if successfullyRead == 0 {
		successfullyRead = bytesNeeded
	}

	return successfullyRead, allBlocksToReturn[startInBlock : startInBlock+bytesNeeded]
}

func (bm *BlockManager) ReadBlockByOffset(filePath string, offset int) []byte {
	blockNumber := offset / BLOCK_SIZE

	block, _ := bm.ReadBlock(filePath, blockNumber)

	return block
}

func (bm *BlockManager) WriteBytesByOffset(filePath string, offset int, data []byte) ([]byte, error) {
	blockNumber := offset / BLOCK_SIZE

	return bm.WriteBlock(filePath, blockNumber, data)
}

func (bm *BlockManager) GetEndOfFile(filePath string) int {
	file := openFile(filePath, os.O_CREATE)

	defer closeFile(file)

	info, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	if info.Size() == 0 || info.Size() < BLOCK_SIZE {
		return 0
	} else if info.Size() < BLOCK_SIZE {
		return -1
	}

	return int(info.Size()) - BLOCK_SIZE
}
