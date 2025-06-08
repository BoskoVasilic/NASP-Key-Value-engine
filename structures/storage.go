package structures

type Storage interface {
	Get(key string) (value string, err error)
	Put(key string, value string) error
	Delete(key string) error
}
