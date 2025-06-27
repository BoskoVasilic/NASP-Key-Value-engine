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
			data: nil,
		},
	}
}

func (s *SkipList) findPath(key string) (bool, []*Node, error) {
	path := make([]*Node, s.currentHeight+1)
	currentNode := s.head
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
func (s *SkipList) Put(key string, value []byte, tombstone bool) error {
	err := s.insert(key, value, tombstone)
	return err
}
func (s *SkipList) insert(key string, value []byte, tombstone bool) error {
	found, path, err := s.findPath(key)
	if found {
		path[s.currentHeight].data.value = value
		path[s.currentHeight].data.deleted = tombstone
		return err
	}
	newData := &Entry{
		key:     key,
		value:   value,
		deleted: tombstone,
	}
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
			data: nil,
		}
		s.head = headNode
	}
	return err
}
func (s *SkipList) getEntry(key string) (Entry, bool, error) {
	var zero Entry
	found, path, err := s.findPath(key)
	fmt.Println("found:", found, "path:", path, len(path), "currentHeight:",
		path[s.currentHeight].data.value, "deleted:", path[s.currentHeight].data.deleted)
	if found {
		if path[s.currentHeight].data.deleted {
			return zero, false, err
		}
		return *path[s.currentHeight].data, true, err
	}
	return zero, false, err
}
func (s *SkipList) Get(key string) (Entry, error) {
	var zero Entry
	value, found, err := s.getEntry(key)
	if err != nil {
		return zero, err
	}
	if !found {
		return zero, nil
	}
	return value, nil
}
func (s *SkipList) Print() {
	curentNode := s.head
	for curentNode != nil {
		printNode := curentNode
		for printNode != nil {
			fmt.Print(printNode.data.key, " -> ")
			printNode = printNode.next
		}
		fmt.Println()
		curentNode = curentNode.down
	}

}
func (s *SkipList) Delete(key string) error {
	err := s.insert(key, nil, true)
	return err
}
func (s *SkipList) GetIter() *utils.Iterator[Node, *utils.ExportEntry] {
	curentNode := s.head
	for curentNode.down != nil {
		curentNode = curentNode.down
	}
	curentNode = curentNode.next
	return utils.NewIterator(curentNode, 0, func(obj *Node, pos int) (*utils.ExportEntry, bool) {
		if obj == nil {
			return nil, false
		}
		Entery := utils.NewExportEntry(obj.data.key, uint64(len(obj.data.value)), obj.data.value)
		obj = obj.next
		return Entery, true
	})
}
