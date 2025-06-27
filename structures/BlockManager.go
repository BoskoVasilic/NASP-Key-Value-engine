package structures

import (
	"encoding/binary"
	"log"
	"math"
	"os"
)

const (
	BLOCK_SIZE         = 8192 // specify the block size in config file
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
}

func NewBlockManager() *BlockManager {
	return &BlockManager{}
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

	return data
}

func (bm *BlockManager) WriteBlock(filePath string, blockNumber int, data []byte) error {
	block := bm.ReadBlock(filePath, blockNumber)

	currentLength := int(binary.LittleEndian.Uint32(block[:BLOCK_HEADER_SIZE]))
	remainingSpace := BLOCK_SIZE - currentLength - RECORD_TYPE_SIZE

	if currentLength+len(data) > remainingSpace {
		newBlock := make([]byte, BLOCK_SIZE)
		newLength := currentLength + remainingSpace + 1
		binary.LittleEndian.PutUint32(block[:BLOCK_HEADER_SIZE], uint32(newLength))
		allBlocksToWrite := make([]byte, 0)
		allBlocksToWrite = append(allBlocksToWrite, block...)
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
		}
	} else if currentLength+len(data) <= remainingSpace {
		block[currentLength] = byte(FULL)
		copy(block[currentLength+RECORD_TYPE_SIZE:], data)
		newLength := currentLength + len(data) + RECORD_TYPE_SIZE
		binary.LittleEndian.PutUint32(block[:BLOCK_HEADER_SIZE], uint32(newLength))
	}

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

	// update or add a block to BufferPool

	return nil
}
