package structures

import (
	utils "KVengine/utils"
	"fmt"
	"math/rand"
)

type Node struct {
	next *Node
	down *Node
	data *Entry
}

type SkipList struct {
	currentHeight int
	head          *Node
	size          int
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
			next: nil,
			down: nil,
			data: &Entry{
				key:     "",
				value:   nil,
				deleted: false,
			},
		},
	}
}

func (s *SkipList) findPath(key string) (bool, []*Node, error) {
	fmt.Println("Finding " + key)
	path := make([]*Node, s.currentHeight+1)
	currentNode := s.head
	fmt.Println(s)
	height := 0
	for currentNode.data.key != key {
		if currentNode.next != nil && currentNode.next.data.key <= key {
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
func (s *SkipList) Put(key string, value []byte) {
	s.insert(key, value, false)
}
func (s *SkipList) insert(key string, value []byte, tombstone bool) error {
	found, path, err := s.findPath(key)
	if found {
		path[s.currentHeight].data.value = value
		path[s.currentHeight].data.deleted = tombstone
		return err
	}
	newData := NewEntry(key, value, tombstone)
	newNode := &Node{
		next: nil,
		down: nil,
		data: newData,
	}
	roll := s.roll()
	Added := min(roll, s.currentHeight)
	for i := 0; i <= Added; i++ {
		newNode.next = path[s.currentHeight-i].next
		path[s.currentHeight-i].next = newNode
		if i != Added {
			newNode = &Node{
				next: nil,
				down: newNode,
				data: newData,
			}
		}
	}
	if roll > s.currentHeight {
		newNode := &Node{
			next: nil,
			down: newNode,
			data: newData,
		}
		s.currentHeight = roll
		headNode := &Node{
			next: newNode,
			down: s.head,
			data: &Entry{
				key:     "",
				value:   nil,
				deleted: false,
			},
		}
		s.head = headNode
	}

	if err != nil {
		s.size++
	}
	return err
}
func (s *SkipList) getEntry(key string) (Entry, bool, error) {
	var zero Entry
	found, path, err := s.findPath(key)
	fmt.Println("found:", found, "path:", path, len(path), "currentHeight:",
		path[s.currentHeight].data.value, "deleted:", path[s.currentHeight].data.deleted)
	if found {
		return *path[s.currentHeight].data, true, err
	}
	return zero, false, err
}
func (s *SkipList) Get(key string) (Entry, bool) {
	var zero Entry
	value, found, err := s.getEntry(key)
	if err != nil {
		return zero, false
	}
	if !found {
		return zero, false
	}
	return value, true
}

func (s *SkipList) Print() {
	curentNode := s.head
	for curentNode != nil {
		printNode := curentNode
		for printNode != nil {
			fmt.Print(printNode.data.key, ":", printNode.data.value, "->")
			printNode = printNode.next
		}
		fmt.Println()
		curentNode = curentNode.down
	}

}
func (s *SkipList) delete(key string) error {
	err := s.insert(key, nil, true)
	return err
}

func (s *SkipList) Delete(key string) {
	err := s.delete(key)
	if err != nil {
		panic(err)
	}
}

func (s *SkipList) GetIter() any {

	curentNode := s.head
	for curentNode.down != nil {
		curentNode = curentNode.down
	}
	curentNode = curentNode.next
	currentNodePointer := &curentNode
	return utils.NewIterator(curentNode, currentNodePointer, func(obj *Node, current **Node) (*Entry, bool) {
		if *current == nil {
			return nil, false
		}

		data := (*current).data
		*current = (*current).next
		return data, true
	})
}

func (s *SkipList) RangeScan(first, last string) *utils.Iterator[string, *Node, *Entry] {
	_, path, _ := s.findPath(first)
	currentNode := path[s.currentHeight]
	if currentNode.data.key < first {
		currentNode = currentNode.next
	}
	return utils.NewIterator(&last, &currentNode, func(obj *string, current **Node) (*Entry, bool) {
		if *current == nil || (*current).data.key > *obj {
			return nil, false
		}
		for (*current).data.deleted {
			*current = (*current).next
			if *current == nil || (*current).data.key > *obj {
				return nil, false
			}
		}
		data := (*current).data
		*current = (*current).next
		return data, true
	})
}

func (s *SkipList) GetSize() int {
	return s.size
}

func (s *SkipList) GetTypeInt() int {
	return 2 // SkipList is represented by 2
}

func (s *SkipList) GetNew() Storage {
	return NewSkipList()
}
