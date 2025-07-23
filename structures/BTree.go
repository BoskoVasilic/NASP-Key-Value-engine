package structures

import (
	"fmt"
	"strings"

	utils "KVengine/utils"
)

// TODO: Check if updating something in PUT also updates the size and height, it should not do so,
// Also, allow deleting something that is not in the tree, it should add that new entry with a tombstone flag

// BTreeNode holds sorted entries
type BTreeNode struct {
	Entries  []Entry
	Children []*BTreeNode
	Parent   *BTreeNode
}

// Entry represents a key-value pair with a tombstone flag
// Root is a node with no parent,
// size is the number of entries,
// height is the tree height,
// order is the maximum number of entries per node
type BTree struct {
	Root   *BTreeNode
	Size   uint32
	Height uint32
	Order  uint8
}

// Initialize a new BTree with a given order
func NewBTree(order uint8) *BTree {
	return &BTree{Order: order}
}

// Put inserts or updates a key–value pair
func (b *BTree) put(key string, value []byte, tombstone bool) {
	if b.Root == nil {
		b.Root = &BTreeNode{
			Entries: []Entry{*NewEntry(key, value, tombstone)},
		}
		b.Height = 1
		b.Size++
	} else {
		insertEntry(b, b.Root, *NewEntry(key, value, tombstone))
	}
}

func (b *BTree) Put(key string, value []byte) {
	b.put(key, value, false)
}

// Get returns the value for a key
func (b *BTree) Get(key string) (Entry, bool) {
	node := b.Root
	for node != nil {
		for i, e := range node.Entries {
			if key == e.key {
				return e, true
			} else if key < e.key {
				if len(node.Children) > 0 {
					node = node.Children[i]
					goto NEXT
				}
				return Entry{}, false
			}
		}
		if len(node.Children) > 0 {
			node = node.Children[len(node.Children)-1]
		} else {
			return Entry{}, false
		}
	NEXT:
	}
	return Entry{}, false
}

// Delete marks a key (an entry) as logically deleted
func (b *BTree) Delete(key string) {
	if b.Root == nil {
		// tree empty, just add a tombstone
		b.Root = &BTreeNode{
			Entries: []Entry{*NewEntry(key, nil, true)},
		}
		b.Height = 1
		b.Size++
		return
	}
	if !markEntrydeleted(b.Root, key) {
		// not found, insert a new tombstone entry
		b.put(key, nil, true)
	}
}

func markEntrydeleted(node *BTreeNode, key string) bool {
	for {
		for i, e := range node.Entries {
			if key == e.key {
				node.Entries[i].deleted = true
				return true
			} else if key < e.key {
				if len(node.Children) > 0 {
					node = node.Children[i]
					goto NEXT
				}
				return false
			}
		}
		if len(node.Children) > 0 {
			node = node.Children[len(node.Children)-1]
		} else {
			return false
		}
	NEXT:
	}
}

// helper functions

func insertEntry(b *BTree, node *BTreeNode, entry Entry) {
	// Leaf
	if len(node.Children) == 0 {
		idx := findInsertIndex(node.Entries, entry.key)
		if idx < len(node.Entries) && node.Entries[idx].key == entry.key {
			node.Entries[idx].value = entry.value
			node.Entries[idx].deleted = entry.deleted
			return
		}
		node.Entries = insertEntryAt(node.Entries, idx, entry)
		b.Size++
		if len(node.Entries) >= int(b.Order) {
			splitNode(b, node)
		}
		return
	}
	// Internal
	idx := findInsertIndex(node.Entries, entry.key)
	if idx < len(node.Entries) && node.Entries[idx].key == entry.key {
		node.Entries[idx].value = entry.value
		node.Entries[idx].deleted = entry.deleted
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
		idx := findInsertIndex(parent.Entries, midEntry.key)
		parent.Entries = insertEntryAt(parent.Entries, idx, midEntry)
		parent.Children = replaceChild(parent.Children, node, left, right)
		if len(parent.Entries) >= int(b.Order) {
			splitNode(b, parent)
		}
	}
}

func findInsertIndex(entries []Entry, key string) int {
	for i, e := range entries {
		if key <= e.key {
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

// GetIter returns an iterator over the BTree entries in sorted (in‑order) order.
// It uses an explicit stack (slice) for traversal and returns Entry values directly.
func (b *BTree) GetIter() any {
	state := btreeIterState{
		nodeStack:  []*BTreeNode{},
		indexStack: []int{},
		current:    b.Root,
	}
	return utils.NewIterator(b, &state, nextBTreeEntry)
}

func (b *BTree) GetSize() int {
	return int(b.Size)
}

// internal state for iteration
type btreeIterState struct {
	nodeStack  []*BTreeNode
	indexStack []int
	current    *BTreeNode
}

// nextBTreeEntry performs one step of in‑order traversal using state
func nextBTreeEntry(tree *BTree, state *btreeIterState) (*Entry, bool) {
	var zero Entry

	// Push leftmost path of current node
	for state.current != nil {
		state.nodeStack = append(state.nodeStack, state.current)
		state.indexStack = append(state.indexStack, 0)
		if len(state.current.Children) > 0 {
			state.current = state.current.Children[0]
		} else {
			state.current = nil
		}
	}

	for len(state.nodeStack) > 0 {
		// Look at top of stack
		n := state.nodeStack[len(state.nodeStack)-1]
		i := state.indexStack[len(state.indexStack)-1]

		if i < len(n.Entries) {
			// Update index for next time
			state.indexStack[len(state.indexStack)-1]++

			// If there’s a right child after this entry, set current to it
			if len(n.Children) > i+1 {
				state.current = n.Children[i+1]
			}

			return &n.Entries[i], true
		} else {
			// Pop finished node
			state.nodeStack = state.nodeStack[:len(state.nodeStack)-1]
			state.indexStack = state.indexStack[:len(state.indexStack)-1]
		}
	}

	return &zero, false
}

// serilization function

func (b *BTree) Serialize() []byte {
	buf := make([]byte, 0)

	// prepend size
	sizeBytes := utils.Uint32ToBytes(b.Size)
	buf = append(buf, sizeBytes...)

	// use iterator to walk entries in sorted order
	iter := b.GetIter().(*utils.Iterator[BTree, btreeIterState, *Entry])
	for {
		entry, ok := iter.Next()
		if !ok {
			break
		}
		buf = append(buf, entry.toBytes()...)
	}

	return buf
}

// for debugging: prints the BTree
func PrintTree(node *BTreeNode, level int) {
	if node == nil {
		return
	}
	fmt.Printf("%sLevel %d: %v\n", strings.Repeat("  ", level), level, node.Entries)
	for _, child := range node.Children {
		PrintTree(child, level+1)
	}
}

func (b *BTree) GetTypeInt() int {
	return 1 // BTree is represented by 1
}

func (b *BTree) GetNew() Storage {
	return NewBTree(b.Order)

}

// test function for BTree
/*
func TestBTree() {
	fmt.Println("\nBTree test with key–value pairs and tombstones:")
	tree := NewBTree(4)

	// Insert
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
	}

	for _, p := range pairs {
		fmt.Printf("Putting: %s -> %s\n", p.key, p.value)
		tree.Put(p.key, []byte(p.value), false)
		PrintTree(tree.Root, 0)
	}

	// logical deletion
	fmt.Printf("\nDeleting key: %s (logical tombstone)\n", "k15")
	tree.Delete("k15")
	PrintTree(tree.Root, 0)

	// retrieval
	key := "k15"
	value, found := tree.Get(key)
	if found {
		fmt.Printf("Get %s: %s\n", key, string(value))
	} else {
		fmt.Printf("Get %s: tombstoned or not found\n", key)
	}

	// iterator test
	iter := tree.GetIter()
	fmt.Println("\nIterating BTree entries:")
	for {
		bin, ok := iter.Next()
		if !ok {
			break
		}
		fmt.Printf("Serialized entry: %v\n", bin)
	}

	// serialize test
	serialized := tree.Serialize()
	fmt.Printf("\nSerialized tree: %v\n", serialized)
}
*/
