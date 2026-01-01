package gomerk

import (
	"fmt"
	"iter"
	"slices"
	"strings"
)

func leftChild(i int) int  { return 2*i + 1 }
func rightChild(i int) int { return 2*i + 2 }
func parent(i int) int     { return (i - 1) / 2 }
func sibling(i int) int    { return ((i + 1) ^ 1) - 1 }

func isTreeNode(n, i int) bool     { return i >= 0 && i < n }
func isInternalNode(n, i int) bool { return isTreeNode(n, leftChild(i)) }
func isLeafNode(n, i int) bool     { return isTreeNode(n, i) && !isInternalNode(n, i) }
func isValidNode(s string) bool    { _, err := HexToBytes32(s); return err == nil }

func checkLeaf(n, i int) error {
	if !isTreeNode(n, i) {
		return ErrIndexOutOfBounds
	}
	if !isLeafNode(n, i) {
		return ErrNotALeaf
	}
	return nil
}

// MakeTree builds a Merkle tree from leaves.
func MakeTree(leaves []Bytes32) ([]string, error) {
	if len(leaves) == 0 {
		return nil, ErrEmptyTree
	}
	n := 2*len(leaves) - 1
	tree := make([]string, n)
	for i, leaf := range leaves {
		tree[n-1-i] = leaf.Hex()
	}
	for i := n - 1 - len(leaves); i >= 0; i-- {
		l, _ := HexToBytes32(tree[leftChild(i)])
		r, _ := HexToBytes32(tree[rightChild(i)])
		tree[i] = HashNode(l, r).Hex()
	}
	return tree, nil
}

// GetProof returns a single proof for a leaf at index.
func GetProof(tree []string, index int) ([]string, error) {
	if err := checkLeaf(len(tree), index); err != nil {
		return nil, err
	}
	var proof []string
	for index > 0 {
		proof = append(proof, tree[sibling(index)])
		index = parent(index)
	}
	return proof, nil
}

// ProcessProof computes the root from a leaf and proof.
func ProcessProof(leaf Bytes32, proof []string) (string, error) {
	current := leaf
	for _, sib := range proof {
		s, err := HexToBytes32(sib)
		if err != nil {
			return "", err
		}
		current = HashNode(current, s)
	}
	return current.Hex(), nil
}

// MultiProof represents a proof for multiple leaves.
type MultiProof struct {
	Leaves     []string `json:"leaves"`
	Proof      []string `json:"proof"`
	ProofFlags []bool   `json:"proofFlags"`
}

// GetMultiProof generates a proof for multiple leaf indices.
func GetMultiProof(tree []string, indices []int) (*MultiProof, error) {
	for _, i := range indices {
		if err := checkLeaf(len(tree), i); err != nil {
			return nil, err
		}
	}

	sorted := slices.Clone(indices)
	slices.SortFunc(sorted, func(a, b int) int { return b - a })

	seen := make(map[int]bool)
	for _, i := range sorted {
		if seen[i] {
			return nil, ErrDuplicatedIndex
		}
		seen[i] = true
	}

	stack := slices.Clone(sorted)
	var proof []string
	var flags []bool

	for len(stack) > 0 && stack[0] > 0 {
		j := stack[0]
		stack = stack[1:]
		s := sibling(j)
		p := parent(j)

		if len(stack) > 0 && s == stack[0] {
			flags = append(flags, true)
			stack = stack[1:]
		} else {
			flags = append(flags, false)
			proof = append(proof, tree[s])
		}

		pos, _ := slices.BinarySearchFunc(stack, p, func(a, b int) int { return b - a })
		stack = slices.Insert(stack, pos, p)
	}

	if len(stack) != 1 {
		proof = append(proof, tree[0])
	}

	leaves := make([]string, len(sorted))
	for i, idx := range sorted {
		leaves[i] = tree[idx]
	}

	return &MultiProof{Leaves: leaves, Proof: proof, ProofFlags: flags}, nil
}

// ProcessMultiProof computes the root from a MultiProof.
func ProcessMultiProof(mp *MultiProof) (string, error) {
	if len(mp.Leaves)+len(mp.Proof) != len(mp.ProofFlags)+1 {
		return "", ErrInvariant
	}

	stack := make([]Bytes32, 0, len(mp.Leaves))
	for _, leaf := range mp.Leaves {
		b, err := HexToBytes32(leaf)
		if err != nil {
			return "", err
		}
		stack = append(stack, b)
	}

	proofIdx := 0
	for _, flag := range mp.ProofFlags {
		if len(stack) == 0 {
			return "", ErrInvariant
		}
		a := stack[0]
		stack = stack[1:]

		var b Bytes32
		if flag {
			if len(stack) == 0 {
				return "", ErrInvariant
			}
			b = stack[0]
			stack = stack[1:]
		} else {
			if proofIdx >= len(mp.Proof) {
				return "", ErrInvariant
			}
			var err error
			b, err = HexToBytes32(mp.Proof[proofIdx])
			if err != nil {
				return "", err
			}
			proofIdx++
		}
		stack = append(stack, HashNode(a, b))
	}

	if len(stack) == 1 {
		return stack[0].Hex(), nil
	}
	if proofIdx < len(mp.Proof) {
		return mp.Proof[proofIdx], nil
	}
	return "", ErrInvariant
}

// IsValidTree checks if tree is a valid Merkle tree.
func IsValidTree(tree []string) bool {
	if len(tree) == 0 {
		return false
	}
	for i, node := range tree {
		if !isValidNode(node) {
			return false
		}
		l, r := leftChild(i), rightChild(i)
		if r >= len(tree) {
			if l < len(tree) {
				return false
			}
			continue
		}
		left, _ := HexToBytes32(tree[l])
		right, _ := HexToBytes32(tree[r])
		nodeB, _ := HexToBytes32(node)
		if nodeB != HashNode(left, right) {
			return false
		}
	}
	return true
}

// RenderTree returns a string representation of the tree.
func RenderTree(tree []string) (string, error) {
	if len(tree) == 0 {
		return "", ErrEmptyTree
	}

	type item struct {
		idx  int
		path []int
	}
	stack := []item{{0, nil}}
	var lines []string

	for len(stack) > 0 {
		it := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		var sb strings.Builder
		for _, p := range it.path[:max(0, len(it.path)-1)] {
			sb.WriteString([2]string{"   ", "│  "}[p])
		}
		if len(it.path) > 0 {
			sb.WriteString([2]string{"└─ ", "├─ "}[it.path[len(it.path)-1]])
		}
		sb.WriteString(fmt.Sprintf("%d) %s", it.idx, tree[it.idx]))
		lines = append(lines, sb.String())

		if rightChild(it.idx) < len(tree) {
			stack = append(stack, item{rightChild(it.idx), append(slices.Clone(it.path), 0)})
			stack = append(stack, item{leftChild(it.idx), append(slices.Clone(it.path), 1)})
		}
	}
	return strings.Join(lines, "\n"), nil
}

// TreeNodes returns an iterator over tree node indices.
func TreeNodes(tree []string) iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		for i, v := range tree {
			if !yield(i, v) {
				return
			}
		}
	}
}

// TreeLeaves returns an iterator over leaf indices.
func TreeLeaves(tree []string) iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		for i, v := range tree {
			if isLeafNode(len(tree), i) {
				if !yield(i, v) {
					return
				}
			}
		}
	}
}
