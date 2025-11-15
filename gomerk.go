package gomerk

import (
	"crypto/sha256"
	"errors"
	"slices"
)

// Node represents a node in the Merkle Tree, holding the hash value.
type Node []byte

// MerkleTree represents the entire tree structure.
type MerkleTree struct {
	root  Node
	layer [][]Node // Bottom-up layers for efficient proof generation
}

// NewMerkleTree constructs a Merkle Tree from a slice of leaf data blocks.
// Leaves are hashed individually. If the number of leaves is odd, the last one is duplicated.
func NewMerkleTree(leaves [][]byte) (*MerkleTree, error) {
	if len(leaves) == 0 {
		return nil, errors.New("merkle: no leaves provided")
	}

	// Hash leaves
	layer := make([]Node, len(leaves))
	for i, leaf := range leaves {
		layer[i] = HashLeaf(leaf)
	}

	// Build layers bottom-up
	layers := [][]Node{layer}
	for len(layer) > 1 {
		nextLayer := make([]Node, (len(layer)+1)/2)
		for i := 0; i < len(layer); i += 2 {
			left := layer[i]
			right := left // Duplicate if odd
			if i+1 < len(layer) {
				right = layer[i+1]
			}
			nextLayer[i/2] = HashChildren(left, right)
		}
		layer = nextLayer
		layers = append(layers, layer)
	}

	return &MerkleTree{
		root:  layer[0],
		layer: layers,
	}, nil
}

// RootHash returns the hash of the root node.
func (t *MerkleTree) RootHash() Node {
	return t.root
}

// GenerateProof returns the Merkle proof for the leaf at the given index.
// The proof is a slice of sibling hashes needed to reconstruct the root.
func (t *MerkleTree) GenerateProof(index int) ([]Node, error) {
	if index < 0 || index >= len(t.layer[0]) {
		return nil, errors.New("merkle: index out of range")
	}

	var proof []Node
	current := index
	for i := 0; i < len(t.layer)-1; i++ {
		sibling := current ^ 1 // XOR for sibling index
		if sibling < len(t.layer[i]) {
			proof = append(proof, t.layer[i][sibling])
		} else {
			// If odd, sibling is duplicate (itself), but already handled in build
			proof = append(proof, t.layer[i][current])
		}
		current /= 2
	}
	return proof, nil
}

// VerifyProof verifies if the given leaf and proof reconstruct the root hash.
// dir indicates if the sibling is left (false) or right (true), but here we infer from structure.
func VerifyProof(proof []Node, root, leaf Node, index int) bool {
	computed := leaf
	current := index
	for _, sibling := range proof {
		if current%2 == 0 {
			computed = HashChildren(computed, sibling)
		} else {
			computed = HashChildren(sibling, computed)
		}
		current /= 2
	}
	return slices.Equal(computed, root)
}

// HashLeaf hashes a single leaf: SHA-256(0x00 || leaf)
func HashLeaf(leaf []byte) Node {
	h := sha256.New()
	h.Write([]byte{0}) // Prefix for leaves (per RFC 6962)
	h.Write(leaf)
	return h.Sum(nil)
}

// HashChildren hashes two child nodes: SHA-256(0x01 || left || right)
func HashChildren(left, right Node) Node {
	h := sha256.New()
	h.Write([]byte{1}) // Prefix for internal nodes
	h.Write(left)
	h.Write(right)
	return h.Sum(nil)
}
