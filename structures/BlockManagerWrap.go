package structures

type BlockManagerWrap struct {
	blockManager    *BlockManager
	currentPosition int
	fileStart       int
	filePath        string
}

func NewBlockManagerWrap(filePath string, blockManager *BlockManager) *BlockManagerWrap {
	return &BlockManagerWrap{
		blockManager:    blockManager,
		currentPosition: 0,
		fileStart:       0,
		filePath:        filePath,
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
