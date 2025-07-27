package structures

type PriorityQueue[T any] struct {
	data []*T
	less func(a, b *T) bool
}

func (pq *PriorityQueue[T]) Len() int {
	return len(pq.data)
}
func (pq *PriorityQueue[T]) Less(i, j int) bool {
	return pq.less(pq.data[i], pq.data[j])
}
func (pq *PriorityQueue[T]) Swap(i, j int) {
	pq.data[i], pq.data[j] = pq.data[j], pq.data[i]
}
func (pq *PriorityQueue[T]) Push(x *T) {
	if x == nil {
		return
	}
	// Perform heap insertion logic
	idx := len(pq.data)
	pq.data = append(pq.data, x)
	for idx > 0 {
		parent := (idx - 1) / 2
		if pq.less(pq.data[idx], pq.data[parent]) {
			pq.data[idx], pq.data[parent] = pq.data[parent], pq.data[idx]
			idx = parent
		} else {
			break
		}
	}

}
func (pq *PriorityQueue[T]) Pop() *T {
	if len(pq.data) == 0 {
		return nil
	}
	top := pq.data[0]
	last := pq.data[len(pq.data)-1]
	pq.data = pq.data[:len(pq.data)-1]
	if len(pq.data) == 0 {
		return top
	}
	pq.data[0] = last
	idx := 0
	for {
		left := 2*idx + 1
		right := 2*idx + 2
		smallest := idx

		if left < len(pq.data) && pq.less(pq.data[left], pq.data[smallest]) {
			smallest = left
		}
		if right < len(pq.data) && pq.less(pq.data[right], pq.data[smallest]) {
			smallest = right
		}
		if smallest == idx {
			break
		}
		pq.data[idx], pq.data[smallest] = pq.data[smallest], pq.data[idx]
		idx = smallest
	}
	return top
}
func NewPriorityQueue[T any](less func(a *T, b *T) bool) *PriorityQueue[T] {
	return &PriorityQueue[T]{
		data: make([]*T, 0),
		less: less,
	}
}
func (pq *PriorityQueue[T]) Peek() *T {
	if len(pq.data) == 0 {
		return nil // Return nil if the queue is empty
	}
	return pq.data[0]
}
func (pq *PriorityQueue[T]) IsEmpty() bool {
	return len(pq.data) == 0
}
