package structures

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	utils "KVengine/utils"
)

// MerkleNode is a tree node
type MerkleNode struct {
	Hash  []byte
	Left  *MerkleNode
	Right *MerkleNode
}

// MerkleTree holds the root and the leaves
type MerkleTree struct {
	Root   *MerkleNode
	Leaves []*MerkleNode
}

// NewMerkleTree returns an empty MerkleTree, like an empty constructor
func NewMerkleTree() *MerkleTree {
	return &MerkleTree{}
}

// Put adds a new leaf node with the hash of data
func (m *MerkleTree) Put(data []byte) {
	hash := sha256.Sum256(data)
	node := &MerkleNode{
		Hash: hash[:],
	}
	m.Leaves = append(m.Leaves, node)
	m.build()
}

// build constructs the tree from current leaves
func (m *MerkleTree) build() {
	nodes := m.Leaves
	if len(nodes) == 0 {
		m.Root = nil
		return
	}
	for len(nodes) > 1 {
		var newLevel []*MerkleNode
		for i := 0; i < len(nodes); i += 2 {
			if i+1 < len(nodes) {
				combined := append(nodes[i].Hash, nodes[i+1].Hash...)
				h := sha256.Sum256(combined)
				parent := &MerkleNode{
					Hash:  h[:],
					Left:  nodes[i],
					Right: nodes[i+1],
				}
				newLevel = append(newLevel, parent)
			} else {
				// odd node promoted
				newLevel = append(newLevel, nodes[i])
			}
		}
		nodes = newLevel
	}
	m.Root = nodes[0]
}

// RootHash returns the hex string of the root hash
func (m *MerkleTree) RootHash() string {
	if m.Root == nil {
		return ""
	}
	return hex.EncodeToString(m.Root.Hash)
}

// CompareTreesWithIndex compares two Merkle trees.
// Returns (true, -1) if trees are equal.
// Returns (false, index) where the first difference in leaves occurs.
func CompareTreesWithIndex(a, b *MerkleTree) (bool, int) {
	if a == nil || b == nil {
		return false, -1
	}

	// Check number of leaves first
	if len(a.Leaves) != len(b.Leaves) {
		return false, -1
	}

	// Check leaf hashes one by one
	for i := 0; i < len(a.Leaves); i++ {
		if !equalHashes(a.Leaves[i].Hash, b.Leaves[i].Hash) {
			return false, i // Mismatch at index i
		}
	}

	// Rebuild and compare roots
	a.build()
	b.build()

	if !equalHashes(a.Root.Hash, b.Root.Hash) {
		return false, -1 // Same leaves but different tree structure
	}

	return true, -1
}

// Helper function to compare two byte slices
func equalHashes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Serialize returns the leaves hashes serialized
func (m *MerkleTree) Serialize() []byte {
	buf := make([]byte, 0)
	leafCount := utils.Uint32ToBytes(uint32(len(m.Leaves)))
	buf = append(buf, leafCount...)
	for _, leaf := range m.Leaves {
		buf = append(buf, leaf.Hash...) // each hash is 32 bytes now
	}
	return buf
}

// Deserialize creates a MerkleTree from serialized data
func DeserializeMerkleTree(data []byte) *MerkleTree {
	if len(data) < 4 {
		return nil // not enough data for leaf count
	}
	leafCount := utils.BytesToUint32(data[:4])
	offset := 4

	const hashSize = 32
	expectedLength := 4 + int(leafCount)*hashSize
	if len(data) < expectedLength {
		return nil // data too short
	}

	m := NewMerkleTree()
	for i := uint32(0); i < leafCount; i++ {
		hashBytes := data[offset : offset+hashSize]
		offset += hashSize
		node := &MerkleNode{
			Hash: append([]byte(nil), hashBytes...), // copy slice
		}
		m.Leaves = append(m.Leaves, node)
	}

	m.build()
	return m
}

// GetIter returns an iterator over the leaf hashes
/*
func (m *MerkleTree) GetIter() *utils.Iterator[[]*MerkleNode, []byte] {
	return utils.NewIterator(&m.Leaves, 0, func(obj *[]*MerkleNode, pos int) ([]byte, bool) {
		if pos >= len(*obj) {
			return nil, false
		}
		return (*obj)[pos].Hash, true
	})
}
*/
// Prints the tree for debugging
func (m *MerkleTree) Print() {
	fmt.Println("Merkle Tree:")
	printMerkle(m.Root, 0)
}

func printMerkle(node *MerkleNode, level int) {
	if node == nil {
		return
	}
	fmt.Printf("%sHash: %x\n", indent(level), node.Hash)
	printMerkle(node.Left, level+1)
	printMerkle(node.Right, level+1)
}

func indent(level int) string {
	s := ""
	for i := 0; i < level; i++ {
		s += "  "
	}
	return s
}

// Tests all MerkleTree operations
/*
func TestMerkle() {
	m1 := NewMerkleTree()
	m2 := NewMerkleTree()

	data := [][]byte{
		[]byte("alpha"),
		[]byte("beta"),
		[]byte("gamma"),
		[]byte("delta"),
	}

	for _, d := range data {
		m1.Put(d)
		m2.Put(d)
	}

	m1.Print()
	fmt.Println("RootHash 1:", m1.RootHash())

	iter := m1.GetIter()
	fmt.Println("Iterating leaves:")
	for {
		val, ok := iter.Next()
		if !ok {
			break
		}
		fmt.Printf("Leaf hash: %x\n", val)
	}

	serialized := m1.Serialize()
	fmt.Printf("Serialized leaves: %x\n", serialized)

	same, index := CompareTreesWithIndex(m1, m2)
	if same {
		fmt.Println("Merkle trees are identical.")
	} else {
		if index == -1 {
			fmt.Println("Trees differ in structure or size.")
		} else {
			fmt.Printf("Trees differ at leaf index %d\n", index)
		}
	}
}
*/
