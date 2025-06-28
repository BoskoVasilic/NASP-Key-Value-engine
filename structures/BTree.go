package structures

import (
	"fmt"
	"sort"
	"strings"

	utils "KVengine/utils"
)

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
func (b *BTree) Put(key string, value []byte, tombstone bool) {
	if b.Root == nil {
		b.Root = &BTreeNode{
			Entries: []Entry{{key: key, value: value, deleted: tombstone}},
		}
		b.Height = 1
		b.Size++
	} else {
		insertEntry(b, b.Root, Entry{key: key, value: value, deleted: tombstone})
	}
}

// Get returns the value for a key.
func (b *BTree) Get(key string) ([]byte, bool) {
	node := b.Root
	for node != nil {
		for i, e := range node.Entries {
			if key == e.key {
				if e.deleted {
					return nil, false
				}
				return e.value, true
			} else if key < e.key {
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

// Delete marks a key as logically deleted.
func (b *BTree) Delete(key string) {
	if b.Root == nil {
		return
	}
	markEntrydeleted(b.Root, key)
}

func markEntrydeleted(node *BTreeNode, key string) {
	for {
		for i, e := range node.Entries {
			if key == e.key {
				node.Entries[i].deleted = true
				return
			} else if key < e.key {
				if len(node.Children) > 0 {
					node = node.Children[i]
					goto NEXT
				}
				return
			}
		}
		if len(node.Children) > 0 {
			node = node.Children[len(node.Children)-1]
		} else {
			return
		}
	NEXT:
	}
}

// ---------------- internal helpers -------------------

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
		if len(node.Entries) >= int(b.Order) {
			splitNode(b, node)
		}
		b.Size++
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

// ---------------- iterator ----------------

func (b *BTree) GetIter() *utils.Iterator[[]*Entry, []byte] {
	allEntries := collectEntries(b.Root)

	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].key < allEntries[j].key
	})

	return utils.NewIterator(&allEntries, 0, func(entries *[]*Entry, pos int) ([]byte, bool) {
		if pos >= len(*entries) {
			return nil, false
		}
		e := (*entries)[pos]
		return e.toBytes(), true
	})
}

func collectEntries(node *BTreeNode) []*Entry {
	if node == nil {
		return nil
	}
	var result []*Entry
	for i := range node.Entries {
		result = append(result, &node.Entries[i])
	}
	for _, c := range node.Children {
		result = append(result, collectEntries(c)...)
	}
	return result
}

// ---------------- serialization ----------------

func (b *BTree) Serialize() []byte {
	buf := make([]byte, 0)

	sizeBytes := utils.Uint32ToBytes(b.Size)
	buf = append(buf, sizeBytes...)

	entries := collectEntries(b.Root)

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].key < entries[j].key
	})

	for _, e := range entries {
		buf = append(buf, e.toBytes()...)
	}

	return buf
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

// ---------------- minimal test ----------------

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
