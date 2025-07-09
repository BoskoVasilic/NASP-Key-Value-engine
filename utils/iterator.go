package utils

type Iterator[T any, V any] struct {
	first   T
	obj     *T
	extract func(*T, int) (V, bool)
	pos     int
}

func NewIterator[T any, V any](obj *T, pos int, extract func(*T, int) (V, bool)) *Iterator[T, V] {
	return &Iterator[T, V]{
		first:   *obj,
		obj:     obj,
		extract: extract,
		pos:     pos,
	}
}

func (it *Iterator[T, V]) Next() (V, bool) {
	var zero V
	val, success := it.extract(it.obj, it.pos)
	if !success {
		return zero, false
	}
	it.pos++
	return val, true
}
func (it *Iterator[T, V]) Reset() {
	it.obj = &it.first
	it.pos = 0
}
