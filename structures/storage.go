package structures

type Storage interface {
	Get(key string) (entry Entry, found bool)
	Put(key string, value []byte)
	Delete(key string)
	GetSize() int
	GetTypeInt() int
	GetNew() Storage
	GetIter() any 
}
