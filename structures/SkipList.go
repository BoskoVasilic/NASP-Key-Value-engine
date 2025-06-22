package structures

import (
	"fmt"
	"math/rand"
)

type Node struct {
	key   string
	value []byte
	next  *Node
	down  *Node
	size  uint64
}

func (n *Node) GetKey() string {
	return n.key
}
func (n *Node) GetValue() []byte {
	return n.value
}
func (n *Node) GetSize() uint64 {
	return n.size
}

type SkipList struct {
	currentHeight int
	head          *Node
}

func (s *SkipList) GetHeight() int {
	return s.currentHeight
}

func (s *SkipList) roll() int {
	level := 0
	// possible ret values from rand are 0 and 1
	// we stop shen we get a 0
	for ; rand.Int31n(2) == 1; level++ {
		if level > s.currentHeight {
			return level
		}
	}
	return level
}
func NewSkipList() *SkipList {
	return &SkipList{
		currentHeight: 0,
		head: &Node{
			key:   "",
			value: nil,
			next:  nil,
			down:  nil,
			size:  0,
		},
	}
}

func (s *SkipList) findPath(key string) (bool, []*Node, error) {
	path := make([]*Node, s.currentHeight+1)
	currentNode := s.head
	height := 0
	for currentNode.key != key {
		if currentNode.next != nil && currentNode.next.key <= key {
			currentNode = currentNode.next
		} else if currentNode.down != nil {
			path[height] = currentNode
			height++
			currentNode = currentNode.down
		} else {
			path[height] = currentNode
			return false, path, nil
		}
	}
	for height < len(path) {
		path[height] = currentNode
		height++
		currentNode = currentNode.down
	}

	return true, path, nil
}
func (s *SkipList) Put(key string, value []byte) error {
	err := s.insert(key, value, uint64(len(value)))
	return err
}
func (s *SkipList) insert(key string, value []byte, size uint64) error {
	found, path, err := s.findPath(key)
	if found {
		path[s.currentHeight].size = size
		path[s.currentHeight].value = value
		return err
	}
	newNode := &Node{
		key:   key,
		value: value,
		next:  nil,
		down:  nil,
		size:  size,
	}
	roll := s.roll()
	Added := min(roll, s.currentHeight)
	for i := 0; i <= Added; i++ {
		newNode.next = path[s.currentHeight-i].next
		path[s.currentHeight-i].next = newNode
		if i != Added {
			newNode = &Node{
				key:   key,
				value: value,
				next:  nil,
				down:  newNode,
				size:  size,
			}
		}
	}
	if roll > s.currentHeight {
		newNode := &Node{
			key:   key,
			value: value,
			next:  nil,
			down:  newNode,
			size:  size,
		}
		s.currentHeight = roll
		headNode := &Node{
			key:  "",
			next: newNode,
			down: s.head,
			size: 0,
		}
		s.head = headNode
	}
	return err
}
func (s *SkipList) getValue(key string) ([]byte, error) {
	found, path, err := s.findPath(key)
	fmt.Println("found:", found, "path:", path, len(path), "currentHeight:", path[s.currentHeight].value, "size:", path[s.currentHeight].size)
	if found {
		if path[s.currentHeight].size == ^uint64(0) {
			return []byte("Deleted"), err
		}
		return path[s.currentHeight].value, err
	}
	return nil, err
}
func (s *SkipList) Get(key string) (string, error) {
	value, err := s.getValue(key)
	if err != nil {
		return "", err
	}
	if value == nil {
		return "", nil
	}
	return string(value), nil
}
func (s *SkipList) Print() {
	curentNode := s.head
	for curentNode != nil {
		printNode := curentNode
		for printNode != nil {
			fmt.Print(printNode.key, " -> ")
			printNode = printNode.next
		}
		fmt.Println()
		curentNode = curentNode.down
	}

}
func (s *SkipList) Delete(key string) error {
	err := s.insert(key, nil, ^uint64(0))
	return err
}
func (s *SkipList) Itterate(yield func(n *Node) bool) {
	currentNode := s.head
	for currentNode.down != nil {
		currentNode = currentNode.down
	}

	currentNode = currentNode.next
	for currentNode != nil {
		if !yield(currentNode) {
			return
		}
		currentNode = currentNode.next
	}
}
