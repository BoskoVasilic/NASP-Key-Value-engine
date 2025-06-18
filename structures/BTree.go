package structures

import (
	"fmt"
	"strings"
)

// Entry represents a key–value pair.
type Entry struct {
	Key   string
	Value []byte
}

// BTreeNode holds sorted entries.
type BTreeNode struct {
	Entries  []Entry
	Children []*BTreeNode
	Parent   *BTreeNode
}

// BTree represents the B-Tree.
type BTree struct {
	Root   *BTreeNode
	Size   uint32
	Height uint32
	Order  uint8
}

// NewBTree creates a B-Tree of given order.
func NewBTree(order uint8) *BTree {
	return &BTree{Order: order}
}

// Put inserts or updates a key–value pair.
func (b *BTree) Put(key string, value []byte) {
	if b.Root == nil {
		b.Root = &BTreeNode{
			Entries: []Entry{{Key: key, Value: value}},
		}
		b.Height = 1
		b.Size++
	} else {
		insertEntry(b, b.Root, Entry{Key: key, Value: value})
	}
}

// Get returns the value for a key.
func (b *BTree) Get(key string) ([]byte, bool) {
	node := b.Root
	for node != nil {
		for i, e := range node.Entries {
			if key == e.Key {
				return e.Value, true
			} else if key < e.Key {
				if len(node.Children) > 0 {
					node = node.Children[i]
					goto NEXT
				}
				return nil, false
			}
		}
		if len(node.Children) > 0 {
			node = node.Children[len(node.Children)-1]
		} else {
			return nil, false
		}
	NEXT:
	}
	return nil, false
}

// Delete removes a key from the B-Tree.
func (b *BTree) Delete(key string) {
	if b.Root == nil {
		return
	}
	deleteEntry(b, b.Root, key)
	if len(b.Root.Entries) == 0 && len(b.Root.Children) > 0 {
		b.Root = b.Root.Children[0]
		b.Root.Parent = nil
		b.Height--
	}
	if len(b.Root.Entries) == 0 && len(b.Root.Children) == 0 {
		b.Root = nil
		b.Height = 0
	}
}

// ---------------- internal helpers -------------------

func insertEntry(b *BTree, node *BTreeNode, entry Entry) {
	// Leaf
	if len(node.Children) == 0 {
		idx := findInsertIndex(node.Entries, entry.Key)
		if idx < len(node.Entries) && node.Entries[idx].Key == entry.Key {
			node.Entries[idx].Value = entry.Value
			return
		}
		node.Entries = insertEntryAt(node.Entries, idx, entry)
		if len(node.Entries) >= int(b.Order) {
			splitNode(b, node)
		}
		b.Size++
		return
	}
	// Internal
	idx := findInsertIndex(node.Entries, entry.Key)
	if idx < len(node.Entries) && node.Entries[idx].Key == entry.Key {
		node.Entries[idx].Value = entry.Value
		return
	}
	insertEntry(b, node.Children[idx], entry)
}

func splitNode(b *BTree, node *BTreeNode) {
	mid := len(node.Entries) / 2
	midEntry := node.Entries[mid]

	left := &BTreeNode{
		Entries: append([]Entry(nil), node.Entries[:mid]...),
		Parent:  node.Parent,
	}
	right := &BTreeNode{
		Entries: append([]Entry(nil), node.Entries[mid+1:]...),
		Parent:  node.Parent,
	}

	if len(node.Children) > 0 {
		left.Children = append([]*BTreeNode(nil), node.Children[:mid+1]...)
		right.Children = append([]*BTreeNode(nil), node.Children[mid+1:]...)
		for _, c := range left.Children {
			c.Parent = left
		}
		for _, c := range right.Children {
			c.Parent = right
		}
	}

	if node.Parent == nil {
		b.Root = &BTreeNode{
			Entries:  []Entry{midEntry},
			Children: []*BTreeNode{left, right},
		}
		left.Parent = b.Root
		right.Parent = b.Root
		b.Height++
	} else {
		parent := node.Parent
		idx := findInsertIndex(parent.Entries, midEntry.Key)
		parent.Entries = insertEntryAt(parent.Entries, idx, midEntry)
		parent.Children = replaceChild(parent.Children, node, left, right)
		if len(parent.Entries) >= int(b.Order) {
			splitNode(b, parent)
		}
	}
}

func deleteEntry(b *BTree, node *BTreeNode, key string) {
	idx := findInsertIndex(node.Entries, key)
	if idx < len(node.Entries) && node.Entries[idx].Key == key {
		if len(node.Children) == 0 {
			node.Entries = append(node.Entries[:idx], node.Entries[idx+1:]...)
			fixUnderflow(b, node)
		} else {
			pre := findPredecessorEntry(node, idx)
			node.Entries[idx] = pre
			deleteEntry(b, node.Children[idx], pre.Key)
		}
	} else if len(node.Children) > 0 {
		deleteEntry(b, node.Children[idx], key)
	}
}

func fixUnderflow(b *BTree, node *BTreeNode) {
	minKeys := int(b.Order)/2 - 1
	if node == b.Root || len(node.Entries) >= minKeys {
		return
	}

	parent := node.Parent
	if parent == nil {
		return
	}

	idx := childIndex(parent, node)

	if idx > 0 && len(parent.Children[idx-1].Entries) > minKeys {
		borrowFromLeft(parent, idx)
	} else if idx < len(parent.Children)-1 && len(parent.Children[idx+1].Entries) > minKeys {
		borrowFromRight(parent, idx)
	} else {
		if idx > 0 {
			mergeWithLeft(b, parent, idx)
		} else {
			mergeWithRight(b, parent, idx)
		}
	}
}

func borrowFromLeft(parent *BTreeNode, idx int) {
	left := parent.Children[idx-1]
	child := parent.Children[idx]
	child.Entries = append([]Entry{parent.Entries[idx-1]}, child.Entries...)
	parent.Entries[idx-1] = left.Entries[len(left.Entries)-1]
	left.Entries = left.Entries[:len(left.Entries)-1]
	if len(left.Children) > 0 {
		child.Children = append([]*BTreeNode{left.Children[len(left.Children)-1]}, child.Children...)
		left.Children = left.Children[:len(left.Children)-1]
		child.Children[0].Parent = child
	}
}

func borrowFromRight(parent *BTreeNode, idx int) {
	right := parent.Children[idx+1]
	child := parent.Children[idx]
	child.Entries = append(child.Entries, parent.Entries[idx])
	parent.Entries[idx] = right.Entries[0]
	right.Entries = right.Entries[1:]
	if len(right.Children) > 0 {
		child.Children = append(child.Children, right.Children[0])
		right.Children[0].Parent = child
		right.Children = right.Children[1:]
	}
}

func mergeWithLeft(b *BTree, parent *BTreeNode, idx int) {
	left := parent.Children[idx-1]
	child := parent.Children[idx]
	left.Entries = append(left.Entries, parent.Entries[idx-1])
	left.Entries = append(left.Entries, child.Entries...)
	if len(child.Children) > 0 {
		left.Children = append(left.Children, child.Children...)
		for _, c := range child.Children {
			c.Parent = left
		}
	}
	parent.Entries = append(parent.Entries[:idx-1], parent.Entries[idx:]...)
	parent.Children = append(parent.Children[:idx], parent.Children[idx+1:]...)
	fixUnderflow(b, parent)
}

func mergeWithRight(b *BTree, parent *BTreeNode, idx int) {
	child := parent.Children[idx]
	right := parent.Children[idx+1]
	child.Entries = append(child.Entries, parent.Entries[idx])
	child.Entries = append(child.Entries, right.Entries...)
	if len(right.Children) > 0 {
		child.Children = append(child.Children, right.Children...)
		for _, c := range right.Children {
			c.Parent = child
		}
	}
	parent.Entries = append(parent.Entries[:idx], parent.Entries[idx+1:]...)
	parent.Children = append(parent.Children[:idx+1], parent.Children[idx+2:]...)
	fixUnderflow(b, parent)
}

func findPredecessorEntry(node *BTreeNode, idx int) Entry {
	curr := node.Children[idx]
	for len(curr.Children) > 0 {
		curr = curr.Children[len(curr.Children)-1]
	}
	return curr.Entries[len(curr.Entries)-1]
}

func findInsertIndex(entries []Entry, key string) int {
	for i, e := range entries {
		if key <= e.Key {
			return i
		}
	}
	return len(entries)
}

func insertEntryAt(entries []Entry, idx int, e Entry) []Entry {
	entries = append(entries, Entry{})
	copy(entries[idx+1:], entries[idx:])
	entries[idx] = e
	return entries
}

func replaceChild(children []*BTreeNode, old, left, right *BTreeNode) []*BTreeNode {
	newChildren := []*BTreeNode{}
	for _, c := range children {
		if c == old {
			newChildren = append(newChildren, left, right)
		} else {
			newChildren = append(newChildren, c)
		}
	}
	return newChildren
}

func childIndex(parent *BTreeNode, child *BTreeNode) int {
	for i, c := range parent.Children {
		if c == child {
			return i
		}
	}
	return -1
}

// ---------------- debug ----------------

func PrintTree(node *BTreeNode, level int) {
	if node == nil {
		return
	}
	fmt.Printf("%sLevel %d: %v\n", strings.Repeat("  ", level), level, node.Entries)
	for _, child := range node.Children {
		PrintTree(child, level+1)
	}
}

func TestBTree() {
	fmt.Println("\nBTree test with key–value pairs:\n")
	tree := NewBTree(4)

	pairs := []struct {
		key   string
		value string
	}{
		{"k10", "apple"},
		{"k20", "banana"},
		{"k05", "grape"},
		{"k15", "peach"},
		{"k25", "melon"},
		{"k30", "pear"},
		{"k35", "plum"},
		{"k40", "kiwi"},
		{"k45", "lime"},
	}

	for _, p := range pairs {
		fmt.Printf("Putting: %s -> %s\n", p.key, p.value)
		tree.Put(p.key, []byte(p.value))
		PrintTree(tree.Root, 0)
		fmt.Println()
	}

	key := "k20"
	value, found := tree.Get(key)
	if found {
		fmt.Printf("Get %s: %s\n", key, string(value))
	} else {
		fmt.Printf("Get %s: not found\n", key)
	}

	fmt.Printf("\nDeleting key: %s\n", "k15")
	tree.Delete("k15")
	PrintTree(tree.Root, 0)

	fmt.Printf("\nDeleting key: %s\n", "k30")
	tree.Delete("k30")
	PrintTree(tree.Root, 0)

	key = "k15"
	value, found = tree.Get(key)
	if found {
		fmt.Printf("Get %s: %s\n", key, string(value))
	} else {
		fmt.Printf("Get %s: not found\n", key)
	}
}
