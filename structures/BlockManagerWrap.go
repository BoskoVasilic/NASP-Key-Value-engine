package structures

import (
	"encoding/binary"
)

type BlockManagerWrap struct {
	blockManager           *BlockManager
	currentPosition        int
	fileStart              int
	filePath               string
	writePostion           int
	writeBuffer            []byte
	currentBufferUsedSpace int
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
	}
}

func (bmw *BlockManagerWrap) ReadFromFile(offset int, bytesNeeded int) []byte {
	data := bmw.blockManager.ReadBytesByOffset(bmw.filePath, offset, bytesNeeded)
	bmw.currentPosition = offset
	bmw.currentPosition += bytesNeeded
	return data
}

func (bmw *BlockManagerWrap) MoveFromCurrentPosition(offset int) {
	bmw.currentPosition += offset
}

func (bmw *BlockManagerWrap) SetCurrentPosition(position int) {
	bmw.currentPosition = position
}

func (bmw *BlockManagerWrap) ReadFromCurrentPosition(bytesNeeded int) []byte {
	data := bmw.blockManager.ReadBytesByOffset(bmw.filePath, bmw.currentPosition, bytesNeeded)
	bmw.currentPosition += bytesNeeded
	return data
}

func (bmw *BlockManagerWrap) WriteToFile(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	if len(bmw.writeBuffer) == 0 {
		switch bmw.writePostion {
		case -1:
			bmw.writePostion = 0
			bmw.writeBuffer = bmw.blockManager.ReadBlockByOffset(bmw.filePath, bmw.writePostion)
			bmw.currentBufferUsedSpace = bmw.blockManager.GetCurrentLength(bmw.writeBuffer)
		case 0:
			newBlock := make([]byte, BLOCK_SIZE)
			binary.LittleEndian.PutUint32(newBlock[:BLOCK_HEADER_SIZE], uint32(BLOCK_HEADER_SIZE))
			bmw.writeBuffer = newBlock
			bmw.currentBufferUsedSpace = BLOCK_HEADER_SIZE
		default:
			bmw.writeBuffer = bmw.blockManager.ReadBlockByOffset(bmw.filePath, bmw.writePostion)
			bmw.currentBufferUsedSpace = bmw.blockManager.GetCurrentLength(bmw.writeBuffer)
		}
	}

	availableBlockSpace := BLOCK_SIZE - bmw.currentBufferUsedSpace - RECORD_TYPE_SIZE

	if len(data) > availableBlockSpace {
		dataToWrite := make([]byte, 0)
		recievedData, lastBlock := bmw.blockManager.AddDataWithBiggerSizeToBlock(bmw.currentBufferUsedSpace, availableBlockSpace, data, bmw.writeBuffer)
		dataToWrite = append(dataToWrite, recievedData...)
		err := bmw.blockManager.WriteToDisk(bmw.filePath, bmw.writePostion/BLOCK_SIZE, dataToWrite[:len(dataToWrite)-BLOCK_SIZE])
		if err != nil {
			return err
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
			return err
		}
		bmw.writePostion += BLOCK_SIZE
		bmw.writeBuffer = bmw.writeBuffer[:0]
		bmw.writeBuffer = newBlock
		bmw.currentBufferUsedSpace = BLOCK_HEADER_SIZE
	}

	return nil
}
