package auxiliary

type DLList[T any] struct {
	Head *Node[T]
	Tail *Node[T]
}

type Node[T any] struct {
	Prev  *Node[T]
	Next  *Node[T]
	Value T
}

func NewDLList[T any]() *DLList[T] {
	return &DLList[T]{
		Head: nil,
		Tail: nil,
	}
}

func (dl *DLList[T]) PushFront(value *T) *Node[T] {
	node := &Node[T]{Value: *value}
	if dl.Head == nil {
		dl.Head = node
		dl.Tail = node
	} else {
		node.Next = dl.Head
		dl.Head.Prev = node
		dl.Head = node
	}
	return node
}

func (dl *DLList[T]) PushBack(value T) *Node[T] {
	node := &Node[T]{Value: value}
	if dl.Tail == nil {
		dl.Head = node
		dl.Tail = node
	} else {
		node.Prev = dl.Tail
		dl.Tail.Next = node
		dl.Tail = node
	}
	return node
}

func (dl *DLList[T]) PopFront() (T, bool) {
	var zero T
	if dl.Head == nil {
		return zero, false
	}
	value := dl.Head.Value
	dl.Head = dl.Head.Next
	if dl.Head == nil {
		dl.Tail = nil
	} else {
		dl.Head.Prev = nil
	}
	return value, true
}

func (dl *DLList[T]) PopBack() (T, bool) {
	var zero T
	if dl.Tail == nil {
		return zero, false
	}
	value := dl.Tail.Value
	dl.Tail = dl.Tail.Prev
	if dl.Tail == nil {
		dl.Head = nil
	} else {
		dl.Tail.Next = nil
	}
	return value, true
}

func (dl *DLList[T]) MoveToFront(node *Node[T]) int {
	if node == nil || node == dl.Head {
		return 0
	}
	node.Prev.Next = node.Next
	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		dl.Tail = node.Prev
	}
	node.Prev = nil
	node.Next = dl.Head
	dl.Head.Prev = node
	dl.Head = node
	return 1
}

func (dl *DLList[T]) Delete(node *Node[T]) {
	if node == nil {
		return
	}
	if node.Prev != nil {
		node.Prev.Next = node.Next
	} else {
		dl.Head = node.Next
	}
	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		dl.Tail = node.Prev
	}
	node.Prev = nil
	node.Next = nil
}
