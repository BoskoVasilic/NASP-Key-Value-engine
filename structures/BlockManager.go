package structures

import (
	"encoding/binary"
	"io"
	"log"
	"math"
	"os"
)

var BLOCK_SIZE int

const (
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
	BLOCK_SIZE = GlobalConfig.BlockManager.BlockSizeKB * 1000
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
	BlockCacheLock.Lock()
	if data, ok := bm.cache.Get(filePath, uint32(blockNumber*BLOCK_SIZE)); ok {
		BlockCacheLock.Unlock()
		return data, nil
	}
	BlockCacheLock.Unlock()

	BlockManagerLock.Lock()
	file := openFile(filePath, os.O_RDONLY)

	_, err := file.Seek(int64(blockNumber*BLOCK_SIZE), 0)
	if err != nil {
		log.Fatal(err)
	}

	data := make([]byte, BLOCK_SIZE)

	file.Read(data)

	closeFile(file)
	BlockManagerLock.Unlock()

	BlockCacheLock.Lock()
	bm.cache.Put(filePath, uint32(blockNumber*BLOCK_SIZE), data)
	BlockCacheLock.Unlock()

	return data, err
}

func (bm *BlockManager) WriteBlock(filePath string, blockNumber int, data []byte) ([]byte, bool, error) {
	block, _ := bm.ReadBlock(filePath, blockNumber)

	currentLength := bm.GetCurrentLength(block)
	remainingSpace := BLOCK_SIZE - currentLength

	lastBlock := make([]byte, 0)
	newBlock := make([]byte, 0)
	additionalBlock := make([]byte, 0)
	addFileSize := false

	if currentLength+len(data) > remainingSpace {
		block, lastBlock = bm.AddDataWithBiggerSizeToBlock(currentLength, remainingSpace, data, block)
	} else if currentLength+len(data) < remainingSpace {
		block, _, newBlock = bm.AddDataWithCorrectSizeToBlock(currentLength, data, block)
	} else if currentLength+len(data) == remainingSpace {
		block, newBlock = bm.AddDataWithEqualSizeToBlock(currentLength, data, block)
	}

	additionalBlock = append(additionalBlock, block...)

	if len(newBlock) > 0 {
		additionalBlock = append(additionalBlock, newBlock...)
		addFileSize = true
	}

	err := bm.WriteToDisk(filePath, blockNumber, additionalBlock)

	if len(newBlock) > 0 {
		return newBlock, addFileSize, err
	}

	return lastBlock, addFileSize, err
}

func (bm *BlockManager) GetCurrentLength(data []byte) int {
	return int(binary.LittleEndian.Uint32(data[BLOCK_SIZE-BLOCK_HEADER_SIZE:]))
}

func (bm *BlockManager) AddDataWithCorrectSizeToBlock(currentLength int, data []byte, block []byte) ([]byte, int, []byte) {
	copy(block[currentLength:], data)
	newLength := currentLength + len(data)
	additionalBlock := make([]byte, 0)
	if BLOCK_SIZE >= newLength+BLOCK_HEADER_SIZE {
		binary.LittleEndian.PutUint32(block[BLOCK_SIZE-BLOCK_HEADER_SIZE:], uint32(newLength))
	} else {
		additionalBlock = make([]byte, BLOCK_SIZE)
		binary.LittleEndian.PutUint32(additionalBlock[BLOCK_SIZE-BLOCK_HEADER_SIZE:], uint32(0))
	}
	return block, newLength, additionalBlock
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
		if len(data) > BLOCK_SIZE-BLOCK_HEADER_SIZE {
			additionalBlock = make([]byte, BLOCK_SIZE)
			binary.LittleEndian.PutUint32(additionalBlock[BLOCK_SIZE-BLOCK_HEADER_SIZE:], uint32(0))
			allBlocksToWrite = append(allBlocksToWrite, newBlock...)
			allBlocksToWrite = append(allBlocksToWrite, additionalBlock...)
			lastBlock = additionalBlock
		} else {
			binary.LittleEndian.PutUint32(newBlock[BLOCK_SIZE-BLOCK_HEADER_SIZE:], uint32(newLength))
			allBlocksToWrite = append(allBlocksToWrite, newBlock...)
			lastBlock = newBlock
		}
		block = allBlocksToWrite
	}
	return block, lastBlock
}

func (bm *BlockManager) WriteToDisk(filePath string, blockNumber int, block []byte) error {
	BlockManagerLock.Lock()
	file := openFile(filePath, os.O_RDWR)

	_, err := file.Seek(int64(blockNumber*BLOCK_SIZE), 0)
	if err != nil {
		log.Fatal(err)
	}

	_, err = file.Write(block)
	if err != nil {
		log.Fatal(err)
	}
	closeFile(file)
	BlockManagerLock.Unlock()

	BlockCacheLock.Lock()
	for i := 0; i < len(block); i += BLOCK_SIZE {
		blockData := block[i : i+BLOCK_SIZE]
		bm.cache.Put(filePath, uint32((blockNumber+i)*BLOCK_SIZE), blockData)
	}
	BlockCacheLock.Unlock()

	return nil
}

func (bm *BlockManager) OverwriteOnDisk(filePath string, offset int, data []byte) error {
	BlockManagerLock.Lock()
	file := openFile(filePath, os.O_RDWR)

	_, err := file.Seek(int64(offset), 0)
	if err != nil {
		log.Fatal(err)
	}

	_, err = file.Write(data)
	if err != nil {
		log.Fatal(err)
	}
	closeFile(file)
	BlockManagerLock.Unlock()

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
				if len(allBlocksToReturn) == 0 {
					allBlocksToReturn = make([]byte, bytesNeeded)
					return 0, allBlocksToReturn
				} else {
					successfullyRead = len(allBlocksToReturn[startInBlock:])
					allBlocksToReturn = append(allBlocksToReturn, make([]byte, bytesNeeded-successfullyRead)...)
					break
				}
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

func (bm *BlockManager) WriteBytesByOffset(filePath string, offset int, data []byte) ([]byte, bool, error) {
	blockNumber := offset / BLOCK_SIZE

	return bm.WriteBlock(filePath, blockNumber, data)
}

func (bm *BlockManager) GetEndOfFile(filePath string) int {
	BlockManagerLock.Lock()
	file := openFile(filePath, os.O_CREATE)

	info, err := file.Stat()

	closeFile(file)
	BlockManagerLock.Unlock()
	if err != nil {
		log.Fatal(err)
	}

	if info.Size() == 0 {
		return -1
	}

	return int(info.Size()) - BLOCK_SIZE
}
