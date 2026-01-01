package gomerk

import (
	"encoding/hex"
	"fmt"
	"iter"
	"math/big"
	"slices"
	"strings"
)

// StandardValue holds a leaf value and its tree index.
type StandardValue struct {
	Value     []any `json:"value"`
	TreeIndex int   `json:"treeIndex"`
}

// StandardTreeData is the serialization format for StandardMerkleTree.
type StandardTreeData struct {
	Format       string          `json:"format"`
	LeafEncoding []string        `json:"leafEncoding"`
	Tree         []string        `json:"tree"`
	Values       []StandardValue `json:"values"`
}

// StandardMerkleTree is a Merkle tree for ABI-encoded structured data.
type StandardMerkleTree struct {
	tree         []string
	values       []StandardValue
	leafEncoding []string
}

// NewStandardMerkleTree creates a new StandardMerkleTree.
func NewStandardMerkleTree(values [][]any, leafEncoding []string, sortLeaves bool) (*StandardMerkleTree, error) {
	type hashed struct {
		value []any
		hash  Bytes32
		index int
	}

	items := make([]hashed, len(values))
	for i, v := range values {
		h, err := encodeAndHash(leafEncoding, v)
		if err != nil {
			return nil, err
		}
		items[i] = hashed{v, h, i}
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

	vals := make([]StandardValue, len(items))
	for i, it := range items {
		vals[it.index] = StandardValue{
			Value:     it.value,
			TreeIndex: len(tree) - 1 - i,
		}
	}

	return &StandardMerkleTree{tree: tree, values: vals, leafEncoding: leafEncoding}, nil
}

// LoadStandardMerkleTree loads a tree from serialized data.
func LoadStandardMerkleTree(data StandardTreeData) (*StandardMerkleTree, error) {
	if data.Format != "standard-v1" {
		return nil, ErrInvalidFormat
	}
	t := &StandardMerkleTree{tree: data.Tree, values: data.Values, leafEncoding: data.LeafEncoding}
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *StandardMerkleTree) Root() string           { return t.tree[0] }
func (t *StandardMerkleTree) Len() int               { return len(t.values) }
func (t *StandardMerkleTree) LeafEncoding() []string { return t.leafEncoding }

func (t *StandardMerkleTree) At(i int) ([]any, bool) {
	if i < 0 || i >= len(t.values) {
		return nil, false
	}
	return t.values[i].Value, true
}

// All returns an iterator over all (index, value) pairs.
func (t *StandardMerkleTree) All() iter.Seq2[int, []any] {
	return func(yield func(int, []any) bool) {
		for i, v := range t.values {
			if !yield(i, v.Value) {
				return
			}
		}
	}
}

// Validate checks tree integrity.
func (t *StandardMerkleTree) Validate() error {
	for _, v := range t.values {
		h, err := encodeAndHash(t.leafEncoding, v.Value)
		if err != nil {
			return err
		}
		if t.tree[v.TreeIndex] != h.Hex() {
			return ErrInvariant
		}
	}
	if !IsValidTree(t.tree) {
		return ErrInvariant
	}
	return nil
}

func (t *StandardMerkleTree) leafIndex(leaf []any) (int, error) {
	h, err := encodeAndHash(t.leafEncoding, leaf)
	if err != nil {
		return -1, err
	}
	for i, v := range t.values {
		if t.tree[v.TreeIndex] == h.Hex() {
			return i, nil
		}
	}
	return -1, ErrLeafNotInTree
}

// GetProof returns a proof for the given leaf.
func (t *StandardMerkleTree) GetProof(leaf []any) ([]string, error) {
	i, err := t.leafIndex(leaf)
	if err != nil {
		return nil, err
	}
	return t.GetProofByIndex(i)
}

// GetProofByIndex returns a proof for the leaf at index.
func (t *StandardMerkleTree) GetProofByIndex(i int) ([]string, error) {
	if i < 0 || i >= len(t.values) {
		return nil, ErrIndexOutOfBounds
	}
	return GetProof(t.tree, t.values[i].TreeIndex)
}

// Verify checks if a leaf is in the tree using the given proof.
func (t *StandardMerkleTree) Verify(leaf []any, proof []string) (bool, error) {
	h, err := encodeAndHash(t.leafEncoding, leaf)
	if err != nil {
		return false, err
	}
	root, err := ProcessProof(h, proof)
	if err != nil {
		return false, err
	}
	return root == t.Root(), nil
}

// GetMultiProofByIndices returns a proof for leaves at the given indices.
func (t *StandardMerkleTree) GetMultiProofByIndices(indices []int) (*MultiProof, error) {
	for _, i := range indices {
		if i < 0 || i >= len(t.values) {
			return nil, ErrIndexOutOfBounds
		}
	}
	treeIndices := make([]int, len(indices))
	for i, idx := range indices {
		treeIndices[i] = t.values[idx].TreeIndex
	}
	return GetMultiProof(t.tree, treeIndices)
}

// VerifyMultiProof checks a multi-proof.
func (t *StandardMerkleTree) VerifyMultiProof(mp *MultiProof) (bool, error) {
	root, err := ProcessMultiProof(mp)
	if err != nil {
		return false, err
	}
	return root == t.Root(), nil
}

// Dump serializes the tree.
func (t *StandardMerkleTree) Dump() StandardTreeData {
	return StandardTreeData{
		Format:       "standard-v1",
		LeafEncoding: t.leafEncoding,
		Tree:         t.tree,
		Values:       t.values,
	}
}

// Render returns a string representation.
func (t *StandardMerkleTree) Render() (string, error) { return RenderTree(t.tree) }

// VerifyStandard is a static verification function.
func VerifyStandard(root string, leafEncoding []string, leaf []any, proof []string) (bool, error) {
	h, err := encodeAndHash(leafEncoding, leaf)
	if err != nil {
		return false, err
	}
	r, err := ProcessProof(h, proof)
	if err != nil {
		return false, err
	}
	return r == root, nil
}

// ABI encoding helpers

func encodeAndHash(types []string, values []any) (Bytes32, error) {
	if len(types) != len(values) {
		return Bytes32{}, ErrMismatchedCount
	}
	var buf []byte
	for i, typ := range types {
		b, err := encodeValue(typ, values[i])
		if err != nil {
			return Bytes32{}, err
		}
		buf = append(buf, b...)
	}
	return HashLeaf(buf), nil
}

func encodeValue(typ string, val any) ([]byte, error) {
	out := make([]byte, 32)

	switch {
	case typ == "address":
		return encodeAddress(val)
	case typ == "bytes32":
		return encodeBytes32(val)
	case strings.HasPrefix(typ, "uint"):
		return encodeUint(val)
	case strings.HasPrefix(typ, "int"):
		return encodeInt(val)
	case typ == "bool":
		if b, ok := val.(bool); ok {
			if b {
				out[31] = 1
			}
			return out, nil
		}
		return nil, ErrAbiEncode
	case typ == "string":
		if s, ok := val.(string); ok {
			h := Keccak256([]byte(s))
			return h[:], nil
		}
		return nil, ErrAbiEncode
	case typ == "bytes":
		return encodeBytes(val)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedType, typ)
	}
}

func encodeAddress(val any) ([]byte, error) {
	s, ok := val.(string)
	if !ok {
		return nil, ErrAbiEncode
	}
	s = strings.TrimPrefix(s, "0x")
	data, err := hex.DecodeString(s)
	if err != nil || len(data) != 20 {
		return nil, ErrAbiEncode
	}
	out := make([]byte, 32)
	copy(out[12:], data)
	return out, nil
}

func encodeBytes32(val any) ([]byte, error) {
	switch v := val.(type) {
	case string:
		b, err := HexToBytes32(v)
		return b[:], err
	case []byte:
		if len(v) != 32 {
			return nil, ErrInvalidNodeLength
		}
		return v, nil
	default:
		return nil, ErrAbiEncode
	}
}

func encodeUint(val any) ([]byte, error) {
	n, err := toBigInt(val)
	if err != nil {
		return nil, err
	}
	if n.Sign() < 0 {
		return nil, ErrAbiEncode
	}
	out := make([]byte, 32)
	b := n.Bytes()
	if len(b) > 32 {
		return nil, ErrAbiEncode
	}
	copy(out[32-len(b):], b)
	return out, nil
}

func encodeInt(val any) ([]byte, error) {
	n, err := toBigInt(val)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 32)
	if n.Sign() >= 0 {
		b := n.Bytes()
		copy(out[32-len(b):], b)
	} else {
		tc := new(big.Int).Add(n, new(big.Int).Lsh(big.NewInt(1), 256))
		b := tc.Bytes()
		for i := range out {
			out[i] = 0xff
		}
		copy(out[32-len(b):], b)
	}
	return out, nil
}

func encodeBytes(val any) ([]byte, error) {
	var data []byte
	switch v := val.(type) {
	case string:
		var err error
		data, err = hex.DecodeString(strings.TrimPrefix(v, "0x"))
		if err != nil {
			return nil, ErrAbiEncode
		}
	case []byte:
		data = v
	default:
		return nil, ErrAbiEncode
	}
	h := Keccak256(data)
	return h[:], nil
}

func toBigInt(val any) (*big.Int, error) {
	n := new(big.Int)
	switch v := val.(type) {
	case int:
		n.SetInt64(int64(v))
	case int64:
		n.SetInt64(v)
	case uint64:
		n.SetUint64(v)
	case float64:
		n.SetInt64(int64(v))
	case string:
		s := strings.TrimPrefix(v, "0x")
		base := 10
		if strings.HasPrefix(v, "0x") {
			base = 16
		}
		if _, ok := n.SetString(s, base); !ok {
			return nil, ErrAbiEncode
		}
	case *big.Int:
		n.Set(v)
	default:
		return nil, ErrAbiEncode
	}
	return n, nil
}
