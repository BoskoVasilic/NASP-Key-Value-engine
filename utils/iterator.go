package utils

type Iterator[T any, H any, V any] struct {
	obj            *T
	extractAndMove func(*T, *H) (V, bool)
	state          *H
}

func NewIterator[T any, H any, V any](obj *T, state *H, extract func(*T, *H) (V, bool)) *Iterator[T, H, V] {

	return &Iterator[T, H, V]{
		obj:            obj,
		extractAndMove: extract,

		state: state,
	}
}

func (it *Iterator[T, H, V]) Next() (V, bool) {
	var zero V
	val, success := it.extractAndMove(it.obj, it.state)
	if !success {
		return zero, false
	}
	return val, true
}
