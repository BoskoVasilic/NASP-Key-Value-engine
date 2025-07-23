package structures

import (
	"encoding/binary"
	"math"
)

type BlockManagerWrap struct {
	blockManager           *BlockManager
	currentPosition        int
	fileStart              int
	fileEnd                int
	filePath               string
	writePosition          int
	writeBuffer            []byte
	currentBufferUsedSpace int
	isBufferFlushed        bool
	isSpaceAvailable       bool
}

func NewBlockManagerWrap(filePath string, blockManager *BlockManager) *BlockManagerWrap {
	return &BlockManagerWrap{
		blockManager:           blockManager,
		currentPosition:        0,
		fileStart:              0,
		filePath:               filePath,
		writePosition:          blockManager.GetEndOfFile(filePath),
		writeBuffer:            make([]byte, 0),
		currentBufferUsedSpace: 0,
		fileEnd:                blockManager.GetEndOfFile(filePath) + BLOCK_SIZE,
		isBufferFlushed:        false,
		isSpaceAvailable:       true,
	}
}

func (bmw *BlockManagerWrap) GetCurrentPosition() int {
	return bmw.currentPosition
}

func (bmw *BlockManagerWrap) GetFileEnd() int {
	return bmw.fileEnd
}

func (bmw *BlockManagerWrap) GetWritePosition() int {
	return bmw.writePosition
}

func (bmw *BlockManagerWrap) setAvilableSpace(isAvailable bool) {
	bmw.isSpaceAvailable = isAvailable
}

func (bmw *BlockManagerWrap) ReadFromFile(offset int, bytesNeeded int) (int, []byte) {
	successfullyRead, data := bmw.blockManager.ReadBytesByOffset(bmw.filePath, offset, bytesNeeded)
	bmw.currentPosition = offset
	bmw.currentPosition += bytesNeeded
	return successfullyRead, data
}

func (bmw *BlockManagerWrap) MoveFromCurrentPosition(offset int) {
	bmw.currentPosition += offset
}

func (bmw *BlockManagerWrap) SetCurrentPosition(position int) {
	bmw.currentPosition = position
}

func (bmw *BlockManagerWrap) GetLastWrittenByteOffset() int {
	bmw.InitializeWriteBuffer()
	return bmw.fileEnd - BLOCK_SIZE + bmw.currentBufferUsedSpace
}

func (bmw *BlockManagerWrap) ReadFromCurrentPosition(bytesNeeded int) (int, []byte) {
	successfullyRead, data := bmw.blockManager.ReadBytesByOffset(bmw.filePath, bmw.currentPosition, bytesNeeded)
	bmw.currentPosition += bytesNeeded
	return successfullyRead, data
}

func (bmw *BlockManagerWrap) InitializeWriteBuffer() {
	if len(bmw.writeBuffer) == 0 {
		switch bmw.writePosition {
		case -1:
			bmw.writePosition = 0
			bmw.fileEnd += 1
			newBlock := make([]byte, BLOCK_SIZE)
			binary.LittleEndian.PutUint32(newBlock[BLOCK_SIZE-BLOCK_HEADER_SIZE:], uint32(0))
			bmw.writeBuffer = newBlock
			bmw.currentBufferUsedSpace = 0
		default:
			bmw.writeBuffer = bmw.blockManager.ReadBlockByOffset(bmw.filePath, bmw.writePosition)
			bmw.currentBufferUsedSpace = bmw.blockManager.GetCurrentLength(bmw.writeBuffer)
		}
	}
}

func (bmw *BlockManagerWrap) WriteToFile(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}

	bmw.InitializeWriteBuffer()

	availableBlockSpace := BLOCK_SIZE - bmw.currentBufferUsedSpace

	if len(data) > availableBlockSpace {
		dataToWrite := make([]byte, 0)
		recievedData, lastBlock := bmw.blockManager.AddDataWithBiggerSizeToBlock(bmw.currentBufferUsedSpace, availableBlockSpace, data, bmw.writeBuffer)
		dataToWrite = append(dataToWrite, recievedData...)
		err := bmw.blockManager.WriteToDisk(bmw.filePath, bmw.writePosition/BLOCK_SIZE, dataToWrite[:len(dataToWrite)-BLOCK_SIZE])
		if err != nil {
			return 0, err
		}
		bmw.writePosition += len(dataToWrite) - BLOCK_SIZE
		bmw.writeBuffer = bmw.writeBuffer[:0]
		bmw.writeBuffer = lastBlock
		bmw.currentBufferUsedSpace = bmw.blockManager.GetCurrentLength(bmw.writeBuffer)
		bmw.isBufferFlushed = false
		bmw.currentPosition = bmw.writePosition
	} else if len(data) < availableBlockSpace {
		var newBlock []byte
		allBlocksToWrite := make([]byte, 0)
		bmw.writeBuffer, bmw.currentBufferUsedSpace, newBlock = bmw.blockManager.AddDataWithCorrectSizeToBlock(bmw.currentBufferUsedSpace, data, bmw.writeBuffer)
		bmw.currentPosition = bmw.writePosition + bmw.currentBufferUsedSpace
		bmw.isBufferFlushed = true
		if len(newBlock) > 0 {
			allBlocksToWrite = append(allBlocksToWrite, bmw.writeBuffer...)
			if bmw.isSpaceAvailable {
				allBlocksToWrite = append(allBlocksToWrite, newBlock...)
			}
			err := bmw.blockManager.WriteToDisk(bmw.filePath, bmw.writePosition/BLOCK_SIZE, allBlocksToWrite)
			if err != nil {
				return 0, err
			}
			bmw.writePosition += BLOCK_SIZE
			bmw.writeBuffer = bmw.writeBuffer[:0]
			bmw.writeBuffer = newBlock
			bmw.currentBufferUsedSpace = 0
			bmw.currentPosition = bmw.writePosition
			bmw.isBufferFlushed = true
			if bmw.isSpaceAvailable {
				bmw.fileEnd += BLOCK_SIZE
			}
		}
	} else if len(data) == availableBlockSpace {
		var newBlock []byte
		allBlocksToWrite := make([]byte, 0)
		bmw.writeBuffer, newBlock = bmw.blockManager.AddDataWithEqualSizeToBlock(bmw.currentBufferUsedSpace, data, bmw.writeBuffer)
		allBlocksToWrite = append(allBlocksToWrite, bmw.writeBuffer...)
		if bmw.isSpaceAvailable {
			allBlocksToWrite = append(allBlocksToWrite, newBlock...)
		}
		err := bmw.blockManager.WriteToDisk(bmw.filePath, bmw.writePosition/BLOCK_SIZE, allBlocksToWrite)
		if err != nil {
			return 0, err
		}

		if bmw.isSpaceAvailable {
			bmw.fileEnd += BLOCK_SIZE
		}
		bmw.writePosition += BLOCK_SIZE
		bmw.writeBuffer = bmw.writeBuffer[:0]
		bmw.writeBuffer = newBlock
		bmw.currentBufferUsedSpace = 0
		bmw.isBufferFlushed = true
		bmw.currentPosition = bmw.writePosition
	}

	if bmw.writePosition > bmw.fileEnd {
		bmw.fileEnd = bmw.writePosition
	}

	return len(data), nil
}

func (bmw *BlockManagerWrap) FlushBuffer(isFromOverwrite bool) error {
	bmw.InitializeWriteBuffer()

	if len(bmw.writeBuffer) == 0 {
		return nil
	}

	if bmw.currentBufferUsedSpace > 0 {
		availableBlockSpace := BLOCK_SIZE - bmw.currentBufferUsedSpace
		copy(bmw.writeBuffer[bmw.currentBufferUsedSpace:BLOCK_SIZE-BLOCK_HEADER_SIZE], make([]byte, availableBlockSpace))
		err := bmw.blockManager.WriteToDisk(bmw.filePath, bmw.writePosition/BLOCK_SIZE, bmw.writeBuffer)
		if err != nil {
			return err
		}
	}

	if !bmw.isBufferFlushed {
		if bmw.writePosition != 0 && bmw.currentBufferUsedSpace > 0 {
			bmw.fileEnd += BLOCK_SIZE
		}
		bmw.isBufferFlushed = true
	}

	if !isFromOverwrite {
		bmw.currentPosition = bmw.writePosition + bmw.currentBufferUsedSpace
	}

	return nil
}

func (bmw *BlockManagerWrap) OverwriteDataInFile(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}

	if bmw.currentPosition+len(data) > bmw.fileEnd-BLOCK_HEADER_SIZE {
		err := bmw.FlushBuffer(true)
		if err != nil {
			return 0, err
		}
		cutOffForBuffer := data[((bmw.fileEnd - BLOCK_SIZE) - bmw.currentPosition):]
		if bmw.currentPosition+len(data) > bmw.fileEnd-BLOCK_HEADER_SIZE {
			err := bmw.blockManager.OverwriteOnDisk(bmw.filePath, bmw.currentPosition, data[:bmw.fileEnd-bmw.currentPosition])
			if err != nil {
				return 0, err
			}

			startPosition := bmw.currentPosition
			bmw.currentPosition += len(data[:bmw.fileEnd-bmw.currentPosition])

			var lastBlock []byte
			addFileSize := false
			lastBlock, addFileSize, err = bmw.blockManager.WriteBlock(bmw.filePath, bmw.currentPosition/BLOCK_SIZE, data[bmw.fileEnd-startPosition:])
			if err != nil {
				return len(data[:bmw.fileEnd-bmw.currentPosition]), err
			}
			cutOfData := data[bmw.fileEnd-startPosition:]

			bmw.fileEnd += int(math.Ceil(float64(len(cutOfData))/float64(BLOCK_SIZE))) * BLOCK_SIZE
			bmw.currentPosition += len(cutOfData)
			if len(lastBlock) > 0 {
				if addFileSize {
					bmw.fileEnd += BLOCK_SIZE
				}
				bmw.writePosition = bmw.fileEnd - BLOCK_SIZE
				bmw.writeBuffer = bmw.writeBuffer[:0]
				bmw.writeBuffer = lastBlock
				bmw.currentBufferUsedSpace = bmw.blockManager.GetCurrentLength(bmw.writeBuffer)
			} else {
				bmw.writePosition = bmw.fileEnd - BLOCK_SIZE
				bmw.writeBuffer = bmw.writeBuffer[:0]
				bmw.writeBuffer = make([]byte, BLOCK_SIZE)
				copy(bmw.writeBuffer, cutOfData)
				bmw.currentBufferUsedSpace = len(cutOfData)
				binary.LittleEndian.PutUint32(bmw.writeBuffer[BLOCK_SIZE-BLOCK_HEADER_SIZE:], uint32(len(cutOfData)))
			}
		} else {
			err := bmw.blockManager.OverwriteOnDisk(bmw.filePath, bmw.currentPosition, data)
			if err != nil {
				return 0, err
			}
			bmw.writeBuffer = bmw.writeBuffer[:0]
			bmw.writeBuffer = make([]byte, BLOCK_SIZE)
			copy(bmw.writeBuffer, cutOffForBuffer)
			bmw.currentBufferUsedSpace = len(cutOffForBuffer)
			binary.LittleEndian.PutUint32(bmw.writeBuffer[BLOCK_SIZE-BLOCK_HEADER_SIZE:], uint32(len(cutOffForBuffer)))
			bmw.currentPosition += len(data)
		}
	} else {
		err := bmw.blockManager.OverwriteOnDisk(bmw.filePath, bmw.currentPosition, data)
		if err != nil {
			return 0, err
		}
		bmw.currentPosition += len(data)
		if bmw.writePosition < bmw.currentPosition {
			copy(bmw.writeBuffer[bmw.currentPosition-len(data):], data)
		}

	}

	return len(data), nil
}
