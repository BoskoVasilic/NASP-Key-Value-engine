package utils

type Iterator[T any] struct {
	obj     *T
	extract func(*T, int) (string, uint64, []byte, bool)
	pos     int
}

func NewIterator[T any](obj *T, pos int, extract func(*T, int) (string, uint64, []byte, bool)) *Iterator[T] {
	return &Iterator[T]{
		obj:     obj,
		extract: extract,
		pos:     pos,
	}
}

func (it *Iterator[T]) Next() (string, uint64, []byte, bool) {
	key, val_size, val, success := it.extract(it.obj, it.pos)
	if !success {
		return "", 0, nil, false
	}
	it.pos++
	return key, val_size, val, true
}
