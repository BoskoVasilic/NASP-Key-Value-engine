package structures

import (
	"crypto/sha256"
	"fmt"
)

// MerkleNode represents a single node in the Merkle Tree
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Hash  []byte
}

// MerkleTree represents the Merkle Tree itself
type MerkleTree struct {
	Root *MerkleNode
}

// NewMerkleNode creates a new Merkle node given left and right children and/or data
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := &MerkleNode{}

	// If no children, hash the data directly
	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		node.Hash = hash[:]
	} else {
		// Concatenate left and right hashes and hash them
		prevHashes := append(left.Hash, right.Hash...)
		hash := sha256.Sum256(prevHashes)
		node.Hash = hash[:]
	}

	node.Left = left
	node.Right = right

	return node
}

// NewMerkleTree builds a Merkle tree from a list of data blocks
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	// Create leaf nodes
	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}

	// Build tree upwards
	for len(nodes) > 1 {
		var newLevel []MerkleNode

		for i := 0; i < len(nodes); i += 2 {
			var left = &nodes[i]
			var right *MerkleNode

			// If odd number of nodes, duplicate last
			if i+1 < len(nodes) {
				right = &nodes[i+1]
			} else {
				right = &nodes[i]
			}

			parent := NewMerkleNode(left, right, nil)
			newLevel = append(newLevel, *parent)
		}

		nodes = newLevel
	}

	tree := &MerkleTree{&nodes[0]}
	return tree
}

func main() {
	// Example data
	data := [][]byte{
		[]byte("Block A"),
		[]byte("Block B"),
		[]byte("Block C"),
		[]byte("Block D"),
	}

	tree := NewMerkleTree(data)

	fmt.Printf("Root Hash: %x\n", tree.Root.Hash)

	// Optional: print tree
	printNode(tree.Root, 0)
}

func printNode(node *MerkleNode, level int) {
	if node == nil {
		return
	}
	fmt.Printf("%sLevel %d: %x\n", indent(level), level, node.Hash)
	printNode(node.Left, level+1)
	printNode(node.Right, level+1)
}

func indent(level int) string {
	return fmt.Sprintf("%s", string(make([]byte, level*2)))
}
