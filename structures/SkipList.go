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
		},
	}
}

func (s *SkipList) findPath(key string) (bool, []*Node) {
	path := make([]*Node, s.currentHeight+1)
	currentNode := s.head
	height := 0
	for currentNode.key != key {
		if currentNode.next != nil && currentNode.next.key < key {
			currentNode = currentNode.next
		} else if currentNode.down != nil {
			fmt.Println(currentNode, height)
			path[height] = currentNode
			height++
			currentNode = currentNode.down
		} else {
			fmt.Println(currentNode, height)
			path[height] = currentNode
			return false, path
		}
	}
	return true, path
}

func (s *SkipList) Insert(key string, value []byte) {
	found, path := s.findPath(key)
	if found {
		path[s.currentHeight].value = value
		return
	}
	newNode := &Node{
		key:   key,
		value: value,
		next:  nil,
		down:  nil,
	}
	roll := s.roll()
	Added := min(roll, s.currentHeight)
	fmt.Println(path)
	for i := 0; i <= Added; i++ {
		newNode.next = path[s.currentHeight-i].next
		path[s.currentHeight-i].next = newNode
		if i != Added {
			newNode = &Node{
				key:   key,
				value: value,
				next:  nil,
				down:  newNode,
			}
		}
	}
	if roll > s.currentHeight {
		newNode := &Node{
			key:   key,
			value: value,
			next:  nil,
			down:  newNode,
		}
		s.currentHeight = roll
		headNode := &Node{
			key:  "",
			next: newNode,
			down: s.head,
		}
		s.head = headNode
	}
}
func (s *SkipList) Get(key string) []byte {
	found, path := s.findPath(key)
	if found {
		return path[s.currentHeight].value
	}
	return nil
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
