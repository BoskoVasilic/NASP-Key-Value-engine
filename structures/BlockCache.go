package structures

type BlockLocation struct {
	filePath string
	offset   uint32
}

type BlockCache struct {
	// Cache of blocks
	cache *Cache[BlockLocation, []byte]
}

func NewBlockCache(maxSize int) *BlockCache {
	return &BlockCache{
		cache: NewCache[BlockLocation, []byte](maxSize),
	}
}

func (bc *BlockCache) Put(filePath string, offset uint32, value []byte) {
	bl := &BlockLocation{filePath: filePath, offset: offset}

	bc.cache.Put(*bl, value)

}

func (bc *BlockCache) Get(filePath string, offset uint32) ([]byte, bool) {
	bl := &BlockLocation{filePath: filePath, offset: offset}

	value, ok := bc.cache.Get(*bl)
	if !ok {
		return nil, false
	}
	return value, true
}
