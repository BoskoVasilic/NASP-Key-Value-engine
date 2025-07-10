package structures

import (
	"encoding/binary"
	"log"
	"math"
	"os"
)

const (
	BLOCK_SIZE         = 8000 // specify the block size in config file
	BLOCK_HEADER_SIZE  = 4
	RECORD_TYPE_SIZE   = 1
	BLOCK_USABLE_SPACE = BLOCK_SIZE - BLOCK_HEADER_SIZE - RECORD_TYPE_SIZE
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

func NewBlockManager() *BlockManager {
	return &BlockManager{
		cache: NewBlockCache(BLOCK_SIZE * 10),
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

func (bm *BlockManager) ReadBlock(filePath string, blockNumber int) []byte {

	// check if block is in BufferPool
	// if yes, return the block from BufferPool

	if data, ok := bm.cache.Get(filePath, uint32(blockNumber*BLOCK_SIZE)); ok {
		return data
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
		log.Fatal(err)
	}

	// add block to BufferPool
	bm.cache.Put(filePath, uint32(blockNumber*BLOCK_SIZE), data)

	return data
}

func (bm *BlockManager) WriteBlock(filePath string, blockNumber int, data []byte) error {
	block := bm.ReadBlock(filePath, blockNumber)

	currentLength := bm.GetCurrentLength(block)
	remainingSpace := BLOCK_SIZE - currentLength - RECORD_TYPE_SIZE

	if currentLength+len(data) > remainingSpace {
		block, _ = bm.AddDataWithBiggerSizeToBlock(currentLength, remainingSpace, data, block)
	} else if currentLength+len(data) <= remainingSpace {
		block, _ = bm.AddDataWithCorrectSizeToBlock(currentLength, data, block)
	}

	err := bm.WriteToDisk(filePath, blockNumber, block)

	return err
}

func (bm *BlockManager) GetCurrentLength(data []byte) int {
	return int(binary.LittleEndian.Uint32(data[:BLOCK_HEADER_SIZE]))
}

func (bm *BlockManager) AddDataWithCorrectSizeToBlock(currentLength int, data []byte, block []byte) ([]byte, int) {
	block[currentLength] = byte(FULL)
	copy(block[currentLength+RECORD_TYPE_SIZE:], data)
	newLength := currentLength + len(data) + RECORD_TYPE_SIZE
	binary.LittleEndian.PutUint32(block[:BLOCK_HEADER_SIZE], uint32(newLength))
	return block, newLength
}

func (bm *BlockManager) AddDataWithEqualSizeToBlock(currentLength int, data []byte, block []byte) ([]byte, []byte) {
	block[currentLength] = byte(FULL)
	copy(block[currentLength+RECORD_TYPE_SIZE:], data)
	newLength := currentLength + len(data) + RECORD_TYPE_SIZE
	binary.LittleEndian.PutUint32(block[:BLOCK_HEADER_SIZE], uint32(newLength))
	newBlock := make([]byte, BLOCK_SIZE)
	binary.LittleEndian.PutUint32(newBlock[:BLOCK_HEADER_SIZE], uint32(BLOCK_HEADER_SIZE))
	return block, newBlock
}

func (bm *BlockManager) AddDataWithBiggerSizeToBlock(currentLength int, remainingSpace int, data []byte, block []byte) ([]byte, []byte) {
	newBlock := make([]byte, BLOCK_SIZE)
	copy(block[currentLength:], make([]byte, remainingSpace+1))
	newLength := currentLength + remainingSpace + 1
	binary.LittleEndian.PutUint32(block[:BLOCK_HEADER_SIZE], uint32(newLength))
	allBlocksToWrite := make([]byte, 0)
	allBlocksToWrite = append(allBlocksToWrite, block...)
	lastBlock := make([]byte, 0)
	if len(data) > BLOCK_USABLE_SPACE {
		numBlocks := math.Ceil(float64(len(data)) / float64(BLOCK_USABLE_SPACE))
		for i := 0; i < int(numBlocks); i++ {
			start := i * BLOCK_USABLE_SPACE
			end := start + BLOCK_USABLE_SPACE
			if end > len(data) {
				end = len(data)
			}
			blockData := data[start:end]

			if i == 0 {
				newBlock[BLOCK_HEADER_SIZE] = byte(FIRST)
			} else if i == int(numBlocks)-1 {
				newBlock[BLOCK_HEADER_SIZE] = byte(LAST)
			} else {
				newBlock[BLOCK_HEADER_SIZE] = byte(MIDDLE)
			}

			copy(newBlock[BLOCK_HEADER_SIZE+RECORD_TYPE_SIZE:], blockData)
			newLength := len(blockData) + RECORD_TYPE_SIZE + BLOCK_HEADER_SIZE
			binary.LittleEndian.PutUint32(newBlock[:BLOCK_HEADER_SIZE], uint32(newLength))

			if i == int(numBlocks)-1 {
				lastBlock = newBlock
			}

			allBlocksToWrite = append(allBlocksToWrite, newBlock...)
		}
		block = allBlocksToWrite
	} else {
		newBlock[BLOCK_HEADER_SIZE] = byte(FULL)
		copy(newBlock[BLOCK_HEADER_SIZE+RECORD_TYPE_SIZE:], data)
		newLength := len(data) + RECORD_TYPE_SIZE + BLOCK_HEADER_SIZE
		binary.LittleEndian.PutUint32(newBlock[:BLOCK_HEADER_SIZE], uint32(newLength))
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

func (bm *BlockManager) ReadBytesByOffset(filePath string, offset int, bytesNeeded int) []byte {
	blockNumber := offset / BLOCK_SIZE
	startInBlock := offset % BLOCK_SIZE

	allBlocksToReturn := make([]byte, 0)
	runWhile := bytesNeeded

	for runWhile > 0 {
		block := bm.ReadBlock(filePath, blockNumber)
		blockNumber++
		allBlocksToReturn = append(allBlocksToReturn, block...)
		runWhile -= BLOCK_SIZE
	}

	return allBlocksToReturn[startInBlock : startInBlock+bytesNeeded]
}

func (bm *BlockManager) ReadBlockByOffset(filePath string, offset int) []byte {
	blockNumber := offset / BLOCK_SIZE

	block := bm.ReadBlock(filePath, blockNumber)

	return block
}

func (bm *BlockManager) WriteBytesByOffset(filePath string, offset int, data []byte) error {
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
