package structures

import (
	"fmt"
	"strings"
)

// BTreeNode represents a single node in the B-Tree
type BTreeNode struct {
	Values   []uint32     // Holds values (keys) in sorted order
	Children []*BTreeNode // Holds pointers to child nodes
	Parent   *BTreeNode   // Pointer to parent node
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

// Add a new value to the BTree
func AddNode(bTree *BTree, value uint32) *BTree {
	if bTree.Root == nil {
		// If the tree is empty, create a new root node
		bTree.Root = NewFreeBTreeNode(value)
		bTree.Height = 1
		bTree.Size++
	} else {
		// Insert the value into the appropriate position
		insertNode(bTree, bTree.Root, value)
	}

	return bTree
}

// Recursive function to insert a value into the BTree
func insertNode(bTree *BTree, currentNode *BTreeNode, value uint32) {
	// If the current node is a leaf, insert the value
	if len(currentNode.Children) == 0 {
		currentNode.Values = insertIntoSortedArray(currentNode.Values, value)

		// If the node overflows, split it
		if len(currentNode.Values) >= int(bTree.Order) {
			splitNode(bTree, currentNode)
		}
		bTree.Size++
		return
	}

	// Traverse to the correct child node
	for i, v := range currentNode.Values {
		if value < v {
			insertNode(bTree, currentNode.Children[i], value)
			return
		}
	}

	// Traverse to the last child if value is greater than all keys
	insertNode(bTree, currentNode.Children[len(currentNode.Children)-1], value)
}

// Helper function to split a node
func splitNode(bTree *BTree, node *BTreeNode) {
	midIndex := len(node.Values) / 2
	midValue := node.Values[midIndex]

	leftNode := &BTreeNode{
		Values:   append([]uint32{}, node.Values[:midIndex]...),
		Parent:   node.Parent,
	}
	rightNode := &BTreeNode{
		Values:   append([]uint32{}, node.Values[midIndex+1:]...),
		Parent:   node.Parent,
	}

	// Only split children if they exist (i.e., node is internal)
	if len(node.Children) > 0 {
		leftNode.Children = make([]*BTreeNode, len(node.Children[:midIndex+1]))
		copy(leftNode.Children, node.Children[:midIndex+1])

		rightNode.Children = make([]*BTreeNode, len(node.Children[midIndex+1:]))
		copy(rightNode.Children, node.Children[midIndex+1:])

		// Update parent for children
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
		// New root case
		newRoot := &BTreeNode{
			Values:   []uint32{midValue},
			Children: []*BTreeNode{leftNode, rightNode},
		}
		leftNode.Parent = newRoot
		rightNode.Parent = newRoot
		bTree.Root = newRoot
		bTree.Height++
	} else {
		// Promote to parent
		parent := node.Parent
		parent.Values = insertIntoSortedArray(parent.Values, midValue)
		parent.Children = replaceChild(parent.Children, node, leftNode, rightNode)

		// Recurse up if parent overflows
		if len(parent.Values) >= int(bTree.Order) {
			splitNode(bTree, parent)
		}
	}
}


// Helper function to insert a value into a sorted slice
func insertIntoSortedArray(arr []uint32, value uint32) []uint32 {
	for i, v := range arr {
		if value < v {
			return append(arr[:i], append([]uint32{value}, arr[i:]...)...)
		}
	}
	return append(arr, value)
}

// Helper function to replace a child node during splitting
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

// DeleteNode removes a value from the BTree
func DeleteNode(bTree *BTree, value uint32) *BTree {
	if bTree.Root == nil {
		return bTree // nothing to do
	}

	deleteNode(bTree, bTree.Root, value)

	// If root has become empty and has children, promote the only child
	if len(bTree.Root.Values) == 0 && len(bTree.Root.Children) > 0 {
		bTree.Root = bTree.Root.Children[0]
		bTree.Root.Parent = nil
		bTree.Height--
	}

	// If root has no values and no children, tree is empty
	if len(bTree.Root.Values) == 0 && len(bTree.Root.Children) == 0 {
		bTree.Root = nil
		bTree.Height = 0
	}

	return bTree
}

// deleteNode recursively deletes a value from the tree rooted at 'node'
func deleteNode(bTree *BTree, node *BTreeNode, value uint32) {
	idx := indexOf(node.Values, value)

	if idx != -1 {
		// Value found in this node
		if len(node.Children) == 0 {
			// Case 1: leaf node
			node.Values = append(node.Values[:idx], node.Values[idx+1:]...)
			fixUnderflow(bTree, node)
		} else {
			// Case 2: internal node
			// Replace with predecessor or successor
			predecessor := findPredecessor(node, idx)
			node.Values[idx] = predecessor
			deleteNode(bTree, node.Children[idx], predecessor)
		}
	} else {
		// Value not found in this node: descend
		if len(node.Children) == 0 {
			// Not found and no children: nothing to do
			return
		}

		// Find child to descend
		childIdx := 0
		for childIdx < len(node.Values) && value > node.Values[childIdx] {
			childIdx++
		}

		deleteNode(bTree, node.Children[childIdx], value)
	}
}

// fixUnderflow ensures the node still has enough keys
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

	// Try left sibling
	if idx > 0 && len(parent.Children[idx-1].Values) > minKeys {
		borrowFromLeft(parent, idx)
	} else if idx < len(parent.Children)-1 && len(parent.Children[idx+1].Values) > minKeys {
		// Try right sibling
		borrowFromRight(parent, idx)
	} else {
		// Merge with sibling
		if idx > 0 {
			mergeWithLeft(bTree, parent, idx)
		} else {
			mergeWithRight(bTree, parent, idx)
		}
	}
}

// Helper: find predecessor (max value in left subtree)
func findPredecessor(node *BTreeNode, idx int) uint32 {
	curr := node.Children[idx]
	for len(curr.Children) > 0 {
		curr = curr.Children[len(curr.Children)-1]
	}
	return curr.Values[len(curr.Values)-1]
}

// Helper: get child index in parent
func childIndex(parent *BTreeNode, child *BTreeNode) int {
	for i, c := range parent.Children {
		if c == child {
			return i
		}
	}
	return -1 // should never happen
}

// Borrow a key from left sibling
func borrowFromLeft(parent *BTreeNode, idx int) {
	left := parent.Children[idx-1]
	child := parent.Children[idx]

	// Pull key from parent down
	child.Values = append([]uint32{parent.Values[idx-1]}, child.Values...)
	// Move sibling's last key up to parent
	parent.Values[idx-1] = left.Values[len(left.Values)-1]
	left.Values = left.Values[:len(left.Values)-1]

	// If there are children, reassign
	if len(left.Children) > 0 {
		child.Children = append([]*BTreeNode{left.Children[len(left.Children)-1]}, child.Children...)
		left.Children = left.Children[:len(left.Children)-1]
		child.Children[0].Parent = child
	}
}

// Borrow a key from right sibling
func borrowFromRight(parent *BTreeNode, idx int) {
	right := parent.Children[idx+1]
	child := parent.Children[idx]

	// Pull key from parent down
	child.Values = append(child.Values, parent.Values[idx])
	// Move sibling's first key up to parent
	parent.Values[idx] = right.Values[0]
	right.Values = right.Values[1:]

	// If there are children, reassign
	if len(right.Children) > 0 {
		child.Children = append(child.Children, right.Children[0])
		right.Children[0].Parent = child
		right.Children = right.Children[1:]
	}
}

// Merge child with left sibling
func mergeWithLeft(bTree *BTree, parent *BTreeNode, idx int) {
	left := parent.Children[idx-1]
	child := parent.Children[idx]

	// Bring down parent key between left and child
	left.Values = append(left.Values, parent.Values[idx-1])
	left.Values = append(left.Values, child.Values...)

	if len(child.Children) > 0 {
		left.Children = append(left.Children, child.Children...)
		for _, c := range child.Children {
			c.Parent = left
		}
	}

	// Remove key from parent and child from parent's children
	parent.Values = append(parent.Values[:idx-1], parent.Values[idx:]...)
	parent.Children = append(parent.Children[:idx], parent.Children[idx+1:]...)

	fixUnderflow(bTree, parent)
}

// Merge child with right sibling
func mergeWithRight(bTree *BTree, parent *BTreeNode, idx int) {
	child := parent.Children[idx]
	right := parent.Children[idx+1]

	// Bring down parent key between child and right
	child.Values = append(child.Values, parent.Values[idx])
	child.Values = append(child.Values, right.Values...)

	if len(right.Children) > 0 {
		child.Children = append(child.Children, right.Children...)
		for _, c := range right.Children {
			c.Parent = child
		}
	}

	// Remove key from parent and right from parent's children
	parent.Values = append(parent.Values[:idx], parent.Values[idx+1:]...)
	parent.Children = append(parent.Children[:idx+1], parent.Children[idx+2:]...)

	fixUnderflow(bTree, parent)
}

// Helper: find index of value in slice
func indexOf(arr []uint32, value uint32) int {
	for i, v := range arr {
		if v == value {
			return i
		}
	}
	return -1
}


// Print the tree structure (for debugging)
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

	fmt.Println()
	fmt.Println("BTree test:")
	fmt.Println()

	// Create a new BTree with order 4 (2-4 tree)
	tree := NewBTree(4)

	// Add values to the tree
	values := []uint32{10, 20, 5, 15, 25, 30, 35, 40, 45}
	for _, value := range values {
		fmt.Printf("Adding value: %d\n", value)
		tree = AddNode(tree, value)
		PrintTree(tree.Root, 0)
		fmt.Println()
	}

	fmt.Printf("Deleting value: %d\n", 15)
	tree = DeleteNode(tree, 15)
	PrintTree(tree.Root, 0)	
	fmt.Println()

	fmt.Printf("Deleting value: %d\n", 30)
	tree = DeleteNode(tree, 30)
	PrintTree(tree.Root, 0)
}
