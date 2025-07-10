package structures

import (
	"encoding/binary"
	"fmt"
	"math"
)

type BlockManagerWrap struct {
	blockManager           *BlockManager
	currentPosition        int
	fileStart              int
	filePath               string
	writePostion           int
	writeBuffer            []byte
	currentBufferUsedSpace int
	fileEnd                int
}

func NewBlockManagerWrap(filePath string, blockManager *BlockManager) *BlockManagerWrap {
	return &BlockManagerWrap{
		blockManager:           blockManager,
		currentPosition:        0,
		fileStart:              0,
		filePath:               filePath,
		writePostion:           blockManager.GetEndOfFile(filePath),
		writeBuffer:            make([]byte, 0),
		currentBufferUsedSpace: 0,
		fileEnd:                blockManager.GetEndOfFile(filePath) + BLOCK_SIZE,
	}
}

func (bmw *BlockManagerWrap) GetCurrentPosition() int {
	return bmw.currentPosition
}

func (bmw *BlockManagerWrap) GetFileEnd() int {
	return bmw.fileEnd
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

func (bmw *BlockManagerWrap) ReadFromCurrentPosition(bytesNeeded int) (int, []byte) {
	successfullyRead, data := bmw.blockManager.ReadBytesByOffset(bmw.filePath, bmw.currentPosition, bytesNeeded)
	bmw.currentPosition += bytesNeeded
	return successfullyRead, data
}

func (bmw *BlockManagerWrap) WriteToFile(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}

	if len(bmw.writeBuffer) == 0 {
		switch bmw.writePostion {
		case -1:
			bmw.writePostion = 0
			bmw.writeBuffer = bmw.blockManager.ReadBlockByOffset(bmw.filePath, bmw.writePostion)
			bmw.currentBufferUsedSpace = bmw.blockManager.GetCurrentLength(bmw.writeBuffer)
		case 0:
			newBlock := make([]byte, BLOCK_SIZE)
			binary.LittleEndian.PutUint32(newBlock[BLOCK_SIZE-BLOCK_HEADER_SIZE:], uint32(0))
			bmw.writeBuffer = newBlock
			bmw.currentBufferUsedSpace = 0
		default:
			bmw.writeBuffer = bmw.blockManager.ReadBlockByOffset(bmw.filePath, bmw.writePostion)
			bmw.currentBufferUsedSpace = bmw.blockManager.GetCurrentLength(bmw.writeBuffer)
		}
	}

	availableBlockSpace := BLOCK_SIZE - bmw.currentBufferUsedSpace

	if len(data) > availableBlockSpace {
		dataToWrite := make([]byte, 0)
		recievedData, lastBlock := bmw.blockManager.AddDataWithBiggerSizeToBlock(bmw.currentBufferUsedSpace, availableBlockSpace, data, bmw.writeBuffer)
		dataToWrite = append(dataToWrite, recievedData...)
		err := bmw.blockManager.WriteToDisk(bmw.filePath, bmw.writePostion/BLOCK_SIZE, dataToWrite[:len(dataToWrite)-BLOCK_SIZE])
		if err != nil {
			return 0, err
		}
		bmw.writePostion += len(dataToWrite) - BLOCK_SIZE
		bmw.writeBuffer = bmw.writeBuffer[:0]
		bmw.writeBuffer = lastBlock
		bmw.currentBufferUsedSpace = bmw.blockManager.GetCurrentLength(bmw.writeBuffer)
	} else if len(data) < availableBlockSpace {
		bmw.writeBuffer, bmw.currentBufferUsedSpace = bmw.blockManager.AddDataWithCorrectSizeToBlock(bmw.currentBufferUsedSpace, data, bmw.writeBuffer)
	} else if len(data) == availableBlockSpace {
		var newBlock []byte
		bmw.writeBuffer, newBlock = bmw.blockManager.AddDataWithEqualSizeToBlock(bmw.currentBufferUsedSpace, data, bmw.writeBuffer)
		err := bmw.blockManager.WriteToDisk(bmw.filePath, bmw.writePostion/BLOCK_SIZE, bmw.writeBuffer)
		if err != nil {
			return 0, err
		}
		bmw.writePostion += BLOCK_SIZE
		bmw.writeBuffer = bmw.writeBuffer[:0]
		bmw.writeBuffer = newBlock
		bmw.currentBufferUsedSpace = 0
	}

	bmw.currentPosition = bmw.writePostion
	bmw.fileEnd = bmw.writePostion

	return len(data), nil
}

func (bmw *BlockManagerWrap) FlushBuffer() error {
	if len(bmw.writeBuffer) == 0 {
		return nil
	}

	if bmw.currentBufferUsedSpace > 0 {
		availableBlockSpace := BLOCK_SIZE - bmw.currentBufferUsedSpace
		copy(bmw.writeBuffer[bmw.currentBufferUsedSpace:], make([]byte, availableBlockSpace))
		err := bmw.blockManager.WriteToDisk(bmw.filePath, bmw.writePostion/BLOCK_SIZE, bmw.writeBuffer)
		if err != nil {
			return err
		}
	}

	bmw.fileEnd += BLOCK_SIZE
	//bmw.currentPosition += bmw.currentBufferUsedSpace

	return nil
}

func (bmw *BlockManagerWrap) OverwriteDataInFile(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}

	fmt.Printf("Current position: %d, file end: %d\n", bmw.currentPosition, bmw.fileEnd)

	if bmw.currentPosition+len(data) > bmw.fileEnd-BLOCK_HEADER_SIZE {
		err := bmw.FlushBuffer()
		if err != nil {
			return 0, err
		}
		fmt.Printf("usao")
		cutOffForBuffer := data[((bmw.fileEnd - BLOCK_SIZE) - bmw.currentPosition):]
		if bmw.currentPosition+len(data) > bmw.fileEnd-BLOCK_HEADER_SIZE {
			err := bmw.blockManager.OverwriteOnDisk(bmw.filePath, bmw.currentPosition, data[:bmw.fileEnd-bmw.currentPosition])
			if err != nil {
				return 0, err
			}

			startPosition := bmw.currentPosition
			bmw.currentPosition += len(data[:bmw.fileEnd-bmw.currentPosition])

			lastBlock := make([]byte, 0)

			lastBlock, err = bmw.blockManager.WriteBlock(bmw.filePath, bmw.currentPosition/BLOCK_SIZE, data[bmw.fileEnd-startPosition:])
			if err != nil {
				return len(data[:bmw.fileEnd]), err
			}
			cutOfData := data[bmw.fileEnd-startPosition:]
			fmt.Printf("Cut of data: %s\n", string(cutOfData))

			bmw.fileEnd += int((math.Ceil(float64(len(cutOfData)) / float64(BLOCK_SIZE))) * BLOCK_SIZE)
			//bmw.currentPosition += len(data)
			if len(lastBlock) > 0 {
				bmw.writePostion = bmw.fileEnd - BLOCK_SIZE
				bmw.writeBuffer = bmw.writeBuffer[:0]
				bmw.writeBuffer = lastBlock
				bmw.currentBufferUsedSpace = bmw.blockManager.GetCurrentLength(bmw.writeBuffer)
				fmt.Printf("New write position: %d, current buffer data\n", bmw.writePostion)
				fmt.Printf("Write buffer last: %s\n", string(bmw.writeBuffer))
			} else {
				bmw.writePostion = bmw.fileEnd - BLOCK_SIZE
				bmw.writeBuffer = bmw.writeBuffer[:0]
				bmw.writeBuffer = make([]byte, BLOCK_SIZE)
				copy(bmw.writeBuffer, cutOfData)
				bmw.currentBufferUsedSpace = len(cutOfData)
				binary.LittleEndian.PutUint32(bmw.writeBuffer[BLOCK_SIZE-BLOCK_HEADER_SIZE:], uint32(len(cutOfData)))
				fmt.Printf("New write position: %d, current buffer data\n", bmw.writePostion)
				fmt.Printf("Write buffer: %s\n", string(bmw.writeBuffer))
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
	}

	return len(data), nil
}
