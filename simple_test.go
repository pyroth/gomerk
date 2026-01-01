package gomerk_test

import (
	"encoding/json"
	"testing"

	"github.com/pyroth/gomerk"
)

func simpleLeaves(n int) []gomerk.Bytes32 {
	out := make([]gomerk.Bytes32, n)
	for i := range out {
		out[i] = gomerk.Keccak256([]byte{byte(i)})
	}
	return out
}

func TestSimpleMerkleTreeNew(t *testing.T) {
	vals := simpleLeaves(4)
	tree, err := gomerk.NewSimpleMerkleTree(vals, true)
	if err != nil {
		t.Fatal(err)
	}
	if tree.Len() != 4 {
		t.Errorf("got len %d, want 4", tree.Len())
	}
	if tree.Root() == "" {
		t.Error("root should not be empty")
	}
}

func TestSimpleMerkleTreeSingle(t *testing.T) {
	tree, err := gomerk.NewSimpleMerkleTree(simpleLeaves(1), true)
	if err != nil {
		t.Fatal(err)
	}
	if tree.Len() != 1 {
		t.Error("single leaf tree should have len 1")
	}
	if err := tree.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestSimpleMerkleTreeValidate(t *testing.T) {
	tree, _ := gomerk.NewSimpleMerkleTree(simpleLeaves(8), true)
	if err := tree.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestSimpleMerkleTreeAt(t *testing.T) {
	vals := simpleLeaves(4)
	tree, _ := gomerk.NewSimpleMerkleTree(vals, true)

	for i := 0; i < tree.Len(); i++ {
		v, ok := tree.At(i)
		if !ok {
			t.Errorf("At(%d) should exist", i)
		}
		if v == "" {
			t.Errorf("At(%d) should not be empty", i)
		}
	}

	_, ok := tree.At(-1)
	if ok {
		t.Error("At(-1) should not exist")
	}
	_, ok = tree.At(tree.Len())
	if ok {
		t.Error("At(Len()) should not exist")
	}
}

func TestSimpleMerkleTreeAll(t *testing.T) {
	vals := simpleLeaves(4)
	tree, _ := gomerk.NewSimpleMerkleTree(vals, true)

	count := 0
	for i, v := range tree.All() {
		got, ok := tree.At(i)
		if !ok || got != v {
			t.Error("iterator mismatch")
		}
		count++
	}
	if count != 4 {
		t.Errorf("got %d, want 4", count)
	}
}

func TestSimpleMerkleTreeGetProof(t *testing.T) {
	vals := simpleLeaves(8)
	tree, _ := gomerk.NewSimpleMerkleTree(vals, true)

	for _, v := range vals {
		proof, err := tree.GetProof(v)
		if err != nil {
			t.Fatal(err)
		}
		ok, _ := tree.Verify(v, proof)
		if !ok {
			t.Error("verify failed")
		}
	}
}

func TestSimpleMerkleTreeGetProofByIndex(t *testing.T) {
	vals := simpleLeaves(8)
	tree, _ := gomerk.NewSimpleMerkleTree(vals, true)

	for i := 0; i < tree.Len(); i++ {
		proof, err := tree.GetProofByIndex(i)
		if err != nil {
			t.Fatal(err)
		}
		v, _ := tree.At(i)
		vb, _ := gomerk.HexToBytes32(v)
		ok, _ := tree.Verify(vb, proof)
		if !ok {
			t.Error("verify by index failed")
		}
	}
}

func TestSimpleMerkleTreeGetProofOutOfBounds(t *testing.T) {
	tree, _ := gomerk.NewSimpleMerkleTree(simpleLeaves(4), true)
	_, err := tree.GetProofByIndex(-1)
	if err != gomerk.ErrIndexOutOfBounds {
		t.Errorf("got %v, want ErrIndexOutOfBounds", err)
	}
	_, err = tree.GetProofByIndex(100)
	if err != gomerk.ErrIndexOutOfBounds {
		t.Errorf("got %v, want ErrIndexOutOfBounds", err)
	}
}

func TestSimpleMerkleTreeLeafNotInTree(t *testing.T) {
	tree, _ := gomerk.NewSimpleMerkleTree(simpleLeaves(4), true)
	_, err := tree.GetProof(gomerk.Bytes32{0xff})
	if err != gomerk.ErrLeafNotInTree {
		t.Errorf("got %v, want ErrLeafNotInTree", err)
	}
}

func TestSimpleMerkleTreeStaticVerify(t *testing.T) {
	vals := simpleLeaves(4)
	tree, _ := gomerk.NewSimpleMerkleTree(vals, true)

	for _, v := range vals {
		proof, _ := tree.GetProof(v)
		ok, err := gomerk.VerifySimple(tree.Root(), v, proof)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Error("static verify failed")
		}
	}
}

func TestSimpleMerkleTreeRejectInvalidProof(t *testing.T) {
	vals1 := simpleLeaves(4)
	tree1, _ := gomerk.NewSimpleMerkleTree(vals1, true)

	vals2 := make([]gomerk.Bytes32, 4)
	for i := range vals2 {
		vals2[i] = gomerk.Keccak256([]byte{byte(i + 100)})
	}
	tree2, _ := gomerk.NewSimpleMerkleTree(vals2, true)

	proof, _ := tree1.GetProof(vals1[0])
	ok, _ := tree2.Verify(vals1[0], proof)
	if ok {
		t.Error("should reject invalid proof")
	}
}

func TestSimpleMerkleTreeMultiProof(t *testing.T) {
	vals := simpleLeaves(8)
	tree, _ := gomerk.NewSimpleMerkleTree(vals, true)

	mp, err := tree.GetMultiProofByIndices([]int{0, 2, 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(mp.Leaves) != 3 {
		t.Errorf("got %d leaves, want 3", len(mp.Leaves))
	}

	ok, _ := tree.VerifyMultiProof(mp)
	if !ok {
		t.Error("multiproof verify failed")
	}
}

func TestSimpleMerkleTreeMultiProofByValues(t *testing.T) {
	vals := simpleLeaves(8)
	tree, _ := gomerk.NewSimpleMerkleTree(vals, true)

	mp, err := tree.GetMultiProof([]gomerk.Bytes32{vals[0], vals[2], vals[5]})
	if err != nil {
		t.Fatal(err)
	}
	ok, _ := tree.VerifyMultiProof(mp)
	if !ok {
		t.Error("multiproof by values failed")
	}
}

func TestSimpleMerkleTreeDump(t *testing.T) {
	vals := simpleLeaves(4)
	tree, _ := gomerk.NewSimpleMerkleTree(vals, true)

	data := tree.Dump()
	if data.Format != "simple-v1" {
		t.Errorf("got %s, want simple-v1", data.Format)
	}
	if len(data.Tree) != 7 {
		t.Errorf("got %d tree nodes, want 7", len(data.Tree))
	}
	if len(data.Values) != 4 {
		t.Errorf("got %d values, want 4", len(data.Values))
	}
}

func TestSimpleMerkleTreeDumpLoad(t *testing.T) {
	vals := simpleLeaves(4)
	tree, _ := gomerk.NewSimpleMerkleTree(vals, true)

	data := tree.Dump()
	js, _ := json.Marshal(data)

	var loaded gomerk.SimpleTreeData
	json.Unmarshal(js, &loaded)

	tree2, err := gomerk.LoadSimpleMerkleTree(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if tree.Root() != tree2.Root() {
		t.Error("roots differ")
	}
	if tree.Len() != tree2.Len() {
		t.Error("lengths differ")
	}
}

func TestSimpleMerkleTreeLoadBadFormat(t *testing.T) {
	tests := []string{"nonstandard", "standard-v1", "bad"}
	for _, format := range tests {
		_, err := gomerk.LoadSimpleMerkleTree(gomerk.SimpleTreeData{Format: format})
		if err != gomerk.ErrInvalidFormat {
			t.Errorf("format %q: got %v, want ErrInvalidFormat", format, err)
		}
	}
}

func TestSimpleMerkleTreeLoadMalformedValue(t *testing.T) {
	zero := "0x0000000000000000000000000000000000000000000000000000000000000000"
	data := gomerk.SimpleTreeData{
		Format: "simple-v1",
		Tree:   []string{zero},
		Values: []gomerk.SimpleValue{{
			Value:     "0x0000000000000000000000000000000000000000000000000000000000000001",
			TreeIndex: 0,
		}},
	}
	_, err := gomerk.LoadSimpleMerkleTree(data)
	if err != gomerk.ErrInvariant {
		t.Errorf("got %v, want ErrInvariant", err)
	}
}

func TestSimpleMerkleTreeLoadInvalidTree(t *testing.T) {
	zero := "0x0000000000000000000000000000000000000000000000000000000000000000"
	data := gomerk.SimpleTreeData{
		Format: "simple-v1",
		Tree:   []string{zero, zero, zero},
		Values: []gomerk.SimpleValue{{Value: zero, TreeIndex: 2}},
	}
	_, err := gomerk.LoadSimpleMerkleTree(data)
	if err != gomerk.ErrInvariant {
		t.Errorf("got %v, want ErrInvariant", err)
	}
}

func TestSimpleMerkleTreeRender(t *testing.T) {
	tree, _ := gomerk.NewSimpleMerkleTree(simpleLeaves(4), true)
	s, err := tree.Render()
	if err != nil {
		t.Fatal(err)
	}
	if s == "" {
		t.Error("render should not be empty")
	}
}

func TestSimpleMerkleTreeUnsorted(t *testing.T) {
	vals := simpleLeaves(4)
	tree, _ := gomerk.NewSimpleMerkleTree(vals, false)
	if err := tree.Validate(); err != nil {
		t.Fatal(err)
	}

	for _, v := range vals {
		proof, _ := tree.GetProof(v)
		ok, _ := tree.Verify(v, proof)
		if !ok {
			t.Error("unsorted tree verify failed")
		}
	}
}
