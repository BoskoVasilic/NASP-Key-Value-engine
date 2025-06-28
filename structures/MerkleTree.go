package structures

import (
	"crypto/md5"
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

// NewMerkleTree returns an empty MerkleTree
func NewMerkleTree() *MerkleTree {
	return &MerkleTree{}
}

// Put adds a new leaf node with the hash of data
func (m *MerkleTree) Put(data []byte) {
	hash := md5.Sum(data)
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
				h := md5.Sum(combined)
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

// Serialize returns the leaves hashes serialized
func (m *MerkleTree) Serialize() []byte {
	buf := make([]byte, 0)
	leafCount := utils.Uint32ToBytes(uint32(len(m.Leaves)))
	buf = append(buf, leafCount...)
	for _, leaf := range m.Leaves {
		buf = append(buf, leaf.Hash...)
	}
	return buf
}

func Deserialize(data []byte) *MerkleTree {
	if len(data) < 4 {
		return nil // not enough data for leaf count
	}
	leafCount := utils.BytesToUint32(data[:4])
	offset := 4

	expectedLength := 4 + int(leafCount)*16 // 16 bytes per MD5 hash
	if len(data) < expectedLength {
		return nil // data too short
	}

	m := NewMerkleTree()
	for i := uint32(0); i < leafCount; i++ {
		hashBytes := data[offset : offset+16]
		offset += 16
		node := &MerkleNode{
			Hash: append([]byte(nil), hashBytes...), // copy slice
		}
		m.Leaves = append(m.Leaves, node)
	}

	m.build()
	return m
}


// GetIter returns an iterator over the leaf hashes
func (m *MerkleTree) GetIter() *utils.Iterator[[]*MerkleNode, []byte] {
	return utils.NewIterator(&m.Leaves, 0, func(obj *[]*MerkleNode, pos int) ([]byte, bool) {
		if pos >= len(*obj) {
			return nil, false
		}
		return (*obj)[pos].Hash, true
	})
}

// Print for debugging
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


func TestMerkle() {
	m := NewMerkleTree()
	data := [][]byte{
		[]byte("alpha"),
		[]byte("beta"),
		[]byte("gamma"),
		[]byte("delta"),
	}

	for _, d := range data {
		m.Put(d)
	}

	m.Print()
	fmt.Println("RootHash:", m.RootHash())

	iter := m.GetIter()
	fmt.Println("Iterating leaves:")
	for {
		val, ok := iter.Next()
		if !ok {
			break
		}
		fmt.Printf("Leaf hash: %x\n", val)
	}

	serialized := m.Serialize()
	fmt.Printf("Serialized leaves: %x\n", serialized)
}
