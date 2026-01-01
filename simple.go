package gomerk

import (
	"iter"
	"slices"
)

// SimpleValue holds a leaf value and its tree index.
type SimpleValue struct {
	Value     string `json:"value"`
	TreeIndex int    `json:"treeIndex"`
}

// SimpleTreeData is the serialization format for SimpleMerkleTree.
type SimpleTreeData struct {
	Format string        `json:"format"`
	Tree   []string      `json:"tree"`
	Values []SimpleValue `json:"values"`
}

// SimpleMerkleTree is a Merkle tree for Bytes32 values.
type SimpleMerkleTree struct {
	tree   []string
	values []SimpleValue
}

// NewSimpleMerkleTree creates a new SimpleMerkleTree from values.
func NewSimpleMerkleTree(values []Bytes32, sortLeaves bool) (*SimpleMerkleTree, error) {
	type hashed struct {
		value Bytes32
		hash  Bytes32
		index int
	}

	items := make([]hashed, len(values))
	for i, v := range values {
		items[i] = hashed{v, HashLeaf(v[:]), i}
	}

	if sortLeaves {
		slices.SortFunc(items, func(a, b hashed) int { return a.hash.Compare(b.hash) })
	}

	leaves := make([]Bytes32, len(items))
	for i, it := range items {
		leaves[i] = it.hash
	}

	tree, err := MakeTree(leaves)
	if err != nil {
		return nil, err
	}

	vals := make([]SimpleValue, len(items))
	for i, it := range items {
		vals[it.index] = SimpleValue{
			Value:     it.value.Hex(),
			TreeIndex: len(tree) - 1 - i,
		}
	}

	return &SimpleMerkleTree{tree: tree, values: vals}, nil
}

// LoadSimpleMerkleTree loads a tree from serialized data.
func LoadSimpleMerkleTree(data SimpleTreeData) (*SimpleMerkleTree, error) {
	if data.Format != "simple-v1" {
		return nil, ErrInvalidFormat
	}
	t := &SimpleMerkleTree{tree: data.Tree, values: data.Values}
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *SimpleMerkleTree) Root() string { return t.tree[0] }
func (t *SimpleMerkleTree) Len() int     { return len(t.values) }

func (t *SimpleMerkleTree) At(i int) (string, bool) {
	if i < 0 || i >= len(t.values) {
		return "", false
	}
	return t.values[i].Value, true
}

// All returns an iterator over all (index, value) pairs.
func (t *SimpleMerkleTree) All() iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		for i, v := range t.values {
			if !yield(i, v.Value) {
				return
			}
		}
	}
}

// Validate checks tree integrity.
func (t *SimpleMerkleTree) Validate() error {
	for _, v := range t.values {
		leaf, err := HexToBytes32(v.Value)
		if err != nil {
			return err
		}
		if t.tree[v.TreeIndex] != HashLeaf(leaf[:]).Hex() {
			return ErrInvariant
		}
	}
	if !IsValidTree(t.tree) {
		return ErrInvariant
	}
	return nil
}

func (t *SimpleMerkleTree) leafIndex(leaf Bytes32) (int, error) {
	h := HashLeaf(leaf[:]).Hex()
	for i, v := range t.values {
		if t.tree[v.TreeIndex] == h {
			vb, _ := HexToBytes32(v.Value)
			if vb == leaf {
				return i, nil
			}
		}
	}
	return -1, ErrLeafNotInTree
}

// GetProof returns a proof for the given leaf.
func (t *SimpleMerkleTree) GetProof(leaf Bytes32) ([]string, error) {
	i, err := t.leafIndex(leaf)
	if err != nil {
		return nil, err
	}
	return t.GetProofByIndex(i)
}

// GetProofByIndex returns a proof for the leaf at index.
func (t *SimpleMerkleTree) GetProofByIndex(i int) ([]string, error) {
	if i < 0 || i >= len(t.values) {
		return nil, ErrIndexOutOfBounds
	}
	return GetProof(t.tree, t.values[i].TreeIndex)
}

// Verify checks if a leaf is in the tree using the given proof.
func (t *SimpleMerkleTree) Verify(leaf Bytes32, proof []string) (bool, error) {
	root, err := ProcessProof(HashLeaf(leaf[:]), proof)
	if err != nil {
		return false, err
	}
	return root == t.Root(), nil
}

// GetMultiProof returns a proof for multiple leaves.
func (t *SimpleMerkleTree) GetMultiProof(leaves []Bytes32) (*MultiProof, error) {
	indices := make([]int, len(leaves))
	for i, leaf := range leaves {
		idx, err := t.leafIndex(leaf)
		if err != nil {
			return nil, err
		}
		indices[i] = idx
	}
	return t.GetMultiProofByIndices(indices)
}

// GetMultiProofByIndices returns a proof for leaves at the given indices.
func (t *SimpleMerkleTree) GetMultiProofByIndices(indices []int) (*MultiProof, error) {
	for _, i := range indices {
		if i < 0 || i >= len(t.values) {
			return nil, ErrIndexOutOfBounds
		}
	}
	treeIndices := make([]int, len(indices))
	for i, idx := range indices {
		treeIndices[i] = t.values[idx].TreeIndex
	}
	mp, err := GetMultiProof(t.tree, treeIndices)
	if err != nil {
		return nil, err
	}
	// Replace hashed leaves with original values
	mp.Leaves = make([]string, len(indices))
	for i, idx := range indices {
		mp.Leaves[i] = t.values[idx].Value
	}
	return mp, nil
}

// VerifyMultiProof checks a multi-proof.
func (t *SimpleMerkleTree) VerifyMultiProof(mp *MultiProof) (bool, error) {
	hashed := make([]string, len(mp.Leaves))
	for i, leaf := range mp.Leaves {
		b, err := HexToBytes32(leaf)
		if err != nil {
			return false, err
		}
		hashed[i] = HashLeaf(b[:]).Hex()
	}
	root, err := ProcessMultiProof(&MultiProof{
		Leaves:     hashed,
		Proof:      mp.Proof,
		ProofFlags: mp.ProofFlags,
	})
	if err != nil {
		return false, err
	}
	return root == t.Root(), nil
}

// Dump serializes the tree.
func (t *SimpleMerkleTree) Dump() SimpleTreeData {
	return SimpleTreeData{Format: "simple-v1", Tree: t.tree, Values: t.values}
}

// Render returns a string representation.
func (t *SimpleMerkleTree) Render() (string, error) { return RenderTree(t.tree) }

// VerifySimple is a static verification function.
func VerifySimple(root string, leaf Bytes32, proof []string) (bool, error) {
	r, err := ProcessProof(HashLeaf(leaf[:]), proof)
	if err != nil {
		return false, err
	}
	return r == root, nil
}
