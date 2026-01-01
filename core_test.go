package gomerk_test

import (
	"strings"
	"testing"

	"github.com/pyroth/gomerk"
)

func testLeaves(n int) []gomerk.Bytes32 {
	out := make([]gomerk.Bytes32, n)
	for i := range out {
		out[i] = gomerk.Keccak256([]byte{byte(i)})
	}
	return out
}

func TestMakeTree(t *testing.T) {
	tests := []int{1, 2, 3, 4, 5, 7, 8, 15, 16, 31, 32}
	for _, n := range tests {
		tree, err := gomerk.MakeTree(testLeaves(n))
		if err != nil {
			t.Fatalf("n=%d: %v", n, err)
		}
		if len(tree) != 2*n-1 {
			t.Errorf("n=%d: got len %d, want %d", n, len(tree), 2*n-1)
		}
		if !gomerk.IsValidTree(tree) {
			t.Errorf("n=%d: tree invalid", n)
		}
	}
}

func TestMakeTreeEmpty(t *testing.T) {
	_, err := gomerk.MakeTree(nil)
	if err != gomerk.ErrEmptyTree {
		t.Errorf("got %v, want ErrEmptyTree", err)
	}
}

func TestGetProof(t *testing.T) {
	leaves := testLeaves(8)
	tree, _ := gomerk.MakeTree(leaves)

	firstLeaf := len(tree) - len(leaves)
	for i := firstLeaf; i < len(tree); i++ {
		proof, err := gomerk.GetProof(tree, i)
		if err != nil {
			t.Fatalf("i=%d: %v", i, err)
		}
		leaf, _ := gomerk.HexToBytes32(tree[i])
		root, _ := gomerk.ProcessProof(leaf, proof)
		if root != tree[0] {
			t.Errorf("i=%d: proof failed", i)
		}
	}
}

func TestGetProofInternalNode(t *testing.T) {
	tree, _ := gomerk.MakeTree(testLeaves(4))
	_, err := gomerk.GetProof(tree, 0)
	if err != gomerk.ErrNotALeaf {
		t.Errorf("got %v, want ErrNotALeaf", err)
	}
}

func TestGetProofOutOfBounds(t *testing.T) {
	tree, _ := gomerk.MakeTree(testLeaves(4))
	_, err := gomerk.GetProof(tree, 100)
	if err != gomerk.ErrIndexOutOfBounds {
		t.Errorf("got %v, want ErrIndexOutOfBounds", err)
	}
}

func TestProcessProofInvalidHex(t *testing.T) {
	_, err := gomerk.ProcessProof(gomerk.Bytes32{}, []string{"invalid"})
	if err == nil {
		t.Error("expected error for invalid hex")
	}
}

func TestMultiProof(t *testing.T) {
	leaves := testLeaves(8)
	tree, _ := gomerk.MakeTree(leaves)
	n := len(tree)

	indices := []int{n - 1, n - 3, n - 5}
	mp, err := gomerk.GetMultiProof(tree, indices)
	if err != nil {
		t.Fatal(err)
	}
	if len(mp.Leaves) != len(indices) {
		t.Errorf("got %d leaves, want %d", len(mp.Leaves), len(indices))
	}

	root, err := gomerk.ProcessMultiProof(mp)
	if err != nil {
		t.Fatal(err)
	}
	if root != tree[0] {
		t.Error("multiproof root mismatch")
	}
}

func TestMultiProofAllLeaves(t *testing.T) {
	leaves := testLeaves(4)
	tree, _ := gomerk.MakeTree(leaves)
	n := len(tree)

	indices := make([]int, len(leaves))
	for i := range indices {
		indices[i] = n - 1 - i
	}

	mp, _ := gomerk.GetMultiProof(tree, indices)
	root, _ := gomerk.ProcessMultiProof(mp)
	if root != tree[0] {
		t.Error("multiproof all leaves failed")
	}
}

func TestMultiProofEmpty(t *testing.T) {
	tree, _ := gomerk.MakeTree(testLeaves(4))
	mp, err := gomerk.GetMultiProof(tree, []int{})
	if err != nil {
		t.Fatal(err)
	}
	if len(mp.Leaves) != 0 {
		t.Error("expected empty leaves")
	}
}

func TestMultiProofDuplicate(t *testing.T) {
	tree, _ := gomerk.MakeTree(testLeaves(4))
	n := len(tree)
	_, err := gomerk.GetMultiProof(tree, []int{n - 1, n - 1})
	if err != gomerk.ErrDuplicatedIndex {
		t.Errorf("got %v, want ErrDuplicatedIndex", err)
	}
}

func TestMultiProofBadFormat(t *testing.T) {
	zero := "0x0000000000000000000000000000000000000000000000000000000000000000"
	mp := &gomerk.MultiProof{
		Leaves:     []string{zero, zero},
		Proof:      []string{zero, zero},
		ProofFlags: []bool{true, true, false},
	}
	_, err := gomerk.ProcessMultiProof(mp)
	if err != gomerk.ErrInvariant {
		t.Errorf("got %v, want ErrInvariant", err)
	}
}

func TestMultiProofStackUnderflow(t *testing.T) {
	zero := "0x0000000000000000000000000000000000000000000000000000000000000000"
	mp := &gomerk.MultiProof{
		Leaves:     []string{zero},
		Proof:      []string{},
		ProofFlags: []bool{true, true},
	}
	_, err := gomerk.ProcessMultiProof(mp)
	if err == nil {
		t.Error("expected error for stack underflow")
	}
}

func TestIsValidTree(t *testing.T) {
	zero := "0x0000000000000000000000000000000000000000000000000000000000000000"

	tests := []struct {
		name  string
		tree  []string
		valid bool
	}{
		{"empty", nil, false},
		{"invalid node", []string{"0x00"}, false},
		{"even count", []string{zero, zero}, false},
		{"wrong hash", []string{zero, zero, zero}, false},
		{"valid single", []string{testLeaves(1)[0].Hex()}, true},
	}
	for _, tc := range tests {
		if gomerk.IsValidTree(tc.tree) != tc.valid {
			t.Errorf("%s: got %v, want %v", tc.name, !tc.valid, tc.valid)
		}
	}

	// Valid tree
	validTree, _ := gomerk.MakeTree(testLeaves(4))
	if !gomerk.IsValidTree(validTree) {
		t.Error("valid tree should be valid")
	}
}

func TestRenderTree(t *testing.T) {
	tree, _ := gomerk.MakeTree(testLeaves(4))
	s, err := gomerk.RenderTree(tree)
	if err != nil {
		t.Fatal(err)
	}
	if s == "" {
		t.Error("render should not be empty")
	}
	// Check contains indices
	if !strings.Contains(s, "0)") || !strings.Contains(s, "1)") {
		t.Error("render should contain indices")
	}
	// Check contains hex
	if !strings.Contains(s, "0x") {
		t.Error("render should contain hex values")
	}
}

func TestRenderTreeEmpty(t *testing.T) {
	_, err := gomerk.RenderTree(nil)
	if err != gomerk.ErrEmptyTree {
		t.Errorf("got %v, want ErrEmptyTree", err)
	}
}

func TestTreeIterators(t *testing.T) {
	tree, _ := gomerk.MakeTree(testLeaves(4))

	// TreeNodes
	nodeCount := 0
	for range gomerk.TreeNodes(tree) {
		nodeCount++
	}
	if nodeCount != len(tree) {
		t.Errorf("TreeNodes: got %d, want %d", nodeCount, len(tree))
	}

	// TreeLeaves
	leafCount := 0
	for range gomerk.TreeLeaves(tree) {
		leafCount++
	}
	if leafCount != 4 {
		t.Errorf("TreeLeaves: got %d, want 4", leafCount)
	}
}
