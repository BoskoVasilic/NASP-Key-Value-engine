package structures

type Storage interface {
	Get(key string) (value []byte, found bool)
	Put(key string, value []byte)
	Delete(key string)
	Iterate(func(uint64, string, []byte) bool)
}
