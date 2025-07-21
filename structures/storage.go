package structures

import "KVengine/utils"

type Storage interface {
	Get(key string) (value []byte, found bool)
	Put(key string, value []byte)
	Delete(key string)
	GetSize() int
	GetIter() utils.Iterator[any, any, Entry]
	Iterate(func(uint64, string, []byte) bool)
}
