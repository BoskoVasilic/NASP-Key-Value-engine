package utils

type Iterator[T any, H any, V any] struct {
	obj            *T
	extractAndMove func(*T, H) (V, bool, H)
	state          H
}

func NewIterator[T any, H any, V any](obj *T, state H, extract func(*T, H) (V, bool)) *Iterator[T, H, V] {

	a, b := extract(obj, state)
	return &Iterator[T, H, V]{
		obj: obj,
		extractAndMove: func(t *T, h H) (V, bool, H) {
			return a, b, state
		},
		state: state,
	}
}

func (it *Iterator[T, H, V]) Next() (V, bool) {
	var zero V
	val, success, state := it.extractAndMove(it.obj, it.state)
	it.state = state
	if !success {
		return zero, false
	}
	return val, true
}
