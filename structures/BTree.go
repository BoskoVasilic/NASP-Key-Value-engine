package structures

import (
	"fmt"
	"strings"
)

// BTreeNode represents a single node in the B-Tree
type BTreeNode struct {
	Values   []uint32
	Children []*BTreeNode
	Parent   *BTreeNode
}

// Create a new free-standing BTreeNode with a single value
func NewFreeBTreeNode(value uint32) *BTreeNode {
	return &BTreeNode{
		Values:   []uint32{value},
		Children: []*BTreeNode{},
		Parent:   nil,
	}
}

// BTree represents the B-Tree
type BTree struct {
	Root   *BTreeNode
	Size   uint32
	Height uint32
	Order  uint8
}

// Create a new BTree with a specified order
func NewBTree(order uint8) *BTree {
	return &BTree{
		Root:   nil,
		Size:   0,
		Height: 0,
		Order:  order,
	}
}

// Put inserts a new value into the BTree
func (bTree *BTree) Put(value uint32) {
	if bTree.Root == nil {
		bTree.Root = NewFreeBTreeNode(value)
		bTree.Height = 1
		bTree.Size++
	} else {
		insertNode(bTree, bTree.Root, value)
	}
}

// Get searches for a value in the BTree
func (bTree *BTree) Get(value uint32) (bool, *BTreeNode) {
	current := bTree.Root
	for current != nil {
		for i, v := range current.Values {
			if value == v {
				return true, current
			} else if value < v {
				if len(current.Children) > 0 {
					current = current.Children[i]
					goto CONTINUE
				}
				return false, nil
			}
		}
		// Check rightmost child
		if len(current.Children) > 0 {
			current = current.Children[len(current.Children)-1]
		} else {
			return false, nil
		}
	CONTINUE:
	}
	return false, nil
}

// Delete removes a value from the BTree
func (bTree *BTree) Delete(value uint32) {
	if bTree.Root == nil {
		return
	}
	deleteNode(bTree, bTree.Root, value)

	// Adjust root if needed
	if len(bTree.Root.Values) == 0 && len(bTree.Root.Children) > 0 {
		bTree.Root = bTree.Root.Children[0]
		bTree.Root.Parent = nil
		bTree.Height--
	}
	if len(bTree.Root.Values) == 0 && len(bTree.Root.Children) == 0 {
		bTree.Root = nil
		bTree.Height = 0
	}
}

// ---------- Internal insertion logic ----------

func insertNode(bTree *BTree, currentNode *BTreeNode, value uint32) {
	if len(currentNode.Children) == 0 {
		currentNode.Values = insertIntoSortedArray(currentNode.Values, value)
		if len(currentNode.Values) >= int(bTree.Order) {
			splitNode(bTree, currentNode)
		}
		bTree.Size++
		return
	}
	for i, v := range currentNode.Values {
		if value < v {
			insertNode(bTree, currentNode.Children[i], value)
			return
		}
	}
	insertNode(bTree, currentNode.Children[len(currentNode.Children)-1], value)
}

func splitNode(bTree *BTree, node *BTreeNode) {
	midIndex := len(node.Values) / 2
	midValue := node.Values[midIndex]

	leftNode := &BTreeNode{
		Values: node.Values[:midIndex],
		Parent: node.Parent,
	}
	rightNode := &BTreeNode{
		Values: node.Values[midIndex+1:],
		Parent: node.Parent,
	}

	if len(node.Children) > 0 {
		leftNode.Children = make([]*BTreeNode, len(node.Children[:midIndex+1]))
		copy(leftNode.Children, node.Children[:midIndex+1])
		rightNode.Children = make([]*BTreeNode, len(node.Children[midIndex+1:]))
		copy(rightNode.Children, node.Children[midIndex+1:])
		for _, child := range leftNode.Children {
			if child != nil {
				child.Parent = leftNode
			}
		}
		for _, child := range rightNode.Children {
			if child != nil {
				child.Parent = rightNode
			}
		}
	}

	if node.Parent == nil {
		newRoot := &BTreeNode{
			Values:   []uint32{midValue},
			Children: []*BTreeNode{leftNode, rightNode},
		}
		leftNode.Parent = newRoot
		rightNode.Parent = newRoot
		bTree.Root = newRoot
		bTree.Height++
	} else {
		parent := node.Parent
		parent.Values = insertIntoSortedArray(parent.Values, midValue)
		parent.Children = replaceChild(parent.Children, node, leftNode, rightNode)
		if len(parent.Values) >= int(bTree.Order) {
			splitNode(bTree, parent)
		}
	}
}

func insertIntoSortedArray(arr []uint32, value uint32) []uint32 {
	for i, v := range arr {
		if value < v {
			return append(arr[:i], append([]uint32{value}, arr[i:]...)...)
		}
	}
	return append(arr, value)
}

func replaceChild(children []*BTreeNode, oldNode, leftNode, rightNode *BTreeNode) []*BTreeNode {
	newChildren := []*BTreeNode{}
	for _, child := range children {
		if child == oldNode {
			newChildren = append(newChildren, leftNode, rightNode)
		} else {
			newChildren = append(newChildren, child)
		}
	}
	return newChildren
}

// ---------- Internal deletion logic ----------

func deleteNode(bTree *BTree, node *BTreeNode, value uint32) {
	idx := indexOf(node.Values, value)

	if idx != -1 {
		if len(node.Children) == 0 {
			node.Values = append(node.Values[:idx], node.Values[idx+1:]...)
			fixUnderflow(bTree, node)
		} else {
			predecessor := findPredecessor(node, idx)
			node.Values[idx] = predecessor
			deleteNode(bTree, node.Children[idx], predecessor)
		}
	} else {
		if len(node.Children) == 0 {
			return
		}
		childIdx := 0
		for childIdx < len(node.Values) && value > node.Values[childIdx] {
			childIdx++
		}
		deleteNode(bTree, node.Children[childIdx], value)
	}
}

func fixUnderflow(bTree *BTree, node *BTreeNode) {
	minKeys := int(bTree.Order)/2 - 1
	if node == bTree.Root || len(node.Values) >= minKeys {
		return
	}

	parent := node.Parent
	if parent == nil {
		return
	}

	idx := childIndex(parent, node)

	if idx > 0 && len(parent.Children[idx-1].Values) > minKeys {
		borrowFromLeft(parent, idx)
	} else if idx < len(parent.Children)-1 && len(parent.Children[idx+1].Values) > minKeys {
		borrowFromRight(parent, idx)
	} else {
		if idx > 0 {
			mergeWithLeft(bTree, parent, idx)
		} else {
			mergeWithRight(bTree, parent, idx)
		}
	}
}

func findPredecessor(node *BTreeNode, idx int) uint32 {
	curr := node.Children[idx]
	for len(curr.Children) > 0 {
		curr = curr.Children[len(curr.Children)-1]
	}
	return curr.Values[len(curr.Values)-1]
}

func childIndex(parent *BTreeNode, child *BTreeNode) int {
	for i, c := range parent.Children {
		if c == child {
			return i
		}
	}
	return -1
}

func borrowFromLeft(parent *BTreeNode, idx int) {
	left := parent.Children[idx-1]
	child := parent.Children[idx]
	child.Values = append([]uint32{parent.Values[idx-1]}, child.Values...)
	parent.Values[idx-1] = left.Values[len(left.Values)-1]
	left.Values = left.Values[:len(left.Values)-1]
	if len(left.Children) > 0 {
		child.Children = append([]*BTreeNode{left.Children[len(left.Children)-1]}, child.Children...)
		left.Children = left.Children[:len(left.Children)-1]
		child.Children[0].Parent = child
	}
}

func borrowFromRight(parent *BTreeNode, idx int) {
	right := parent.Children[idx+1]
	child := parent.Children[idx]
	child.Values = append(child.Values, parent.Values[idx])
	parent.Values[idx] = right.Values[0]
	right.Values = right.Values[1:]
	if len(right.Children) > 0 {
		child.Children = append(child.Children, right.Children[0])
		right.Children[0].Parent = child
		right.Children = right.Children[1:]
	}
}

func mergeWithLeft(bTree *BTree, parent *BTreeNode, idx int) {
	left := parent.Children[idx-1]
	child := parent.Children[idx]
	left.Values = append(left.Values, parent.Values[idx-1])
	left.Values = append(left.Values, child.Values...)
	if len(child.Children) > 0 {
		left.Children = append(left.Children, child.Children...)
		for _, c := range child.Children {
			c.Parent = left
		}
	}
	parent.Values = append(parent.Values[:idx-1], parent.Values[idx:]...)
	parent.Children = append(parent.Children[:idx], parent.Children[idx+1:]...)
	fixUnderflow(bTree, parent)
}

func mergeWithRight(bTree *BTree, parent *BTreeNode, idx int) {
	child := parent.Children[idx]
	right := parent.Children[idx+1]
	child.Values = append(child.Values, parent.Values[idx])
	child.Values = append(child.Values, right.Values...)
	if len(right.Children) > 0 {
		child.Children = append(child.Children, right.Children...)
		for _, c := range right.Children {
			c.Parent = child
		}
	}
	parent.Values = append(parent.Values[:idx], parent.Values[idx+1:]...)
	parent.Children = append(parent.Children[:idx+1], parent.Children[idx+2:]...)
	fixUnderflow(bTree, parent)
}

func indexOf(arr []uint32, value uint32) int {
	for i, v := range arr {
		if v == value {
			return i
		}
	}
	return -1
}

func PrintTree(node *BTreeNode, level int) {
	if node == nil {
		return
	}
	fmt.Printf("%sLevel %d: %v\n", strings.Repeat(" ", level*2), level, node.Values)
	for _, child := range node.Children {
		PrintTree(child, level+1)
	}
}

func TestBTree() {
	fmt.Println("\nBTree test:\n")
	tree := NewBTree(4)
	values := []uint32{10, 20, 5, 15, 25, 30, 35, 40, 45}
	for _, value := range values {
		fmt.Printf("Putting value: %d\n", value)
		tree.Put(value)
		PrintTree(tree.Root, 0)
		fmt.Println()
	}

	fmt.Printf("Deleting value: %d\n", 15)
	tree.Delete(15)
	PrintTree(tree.Root, 0)
	fmt.Println()

	fmt.Printf("Deleting value: %d\n", 30)
	tree.Delete(30)
	PrintTree(tree.Root, 0)
	fmt.Println()

	// Try Get:
	found, node := tree.Get(20)
	fmt.Println("Get 20:", found, node)
	found2, node2 := tree.Get(50)
	fmt.Println("Get 50:", found2, node2)
}
