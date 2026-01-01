package gomerk_test

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/pyroth/gomerk"
)

func airdropData(n int) [][]any {
	vals := make([][]any, n)
	for i := range vals {
		vals[i] = []any{
			"0x" + padAddr(i+1),
			(i + 1) * 100,
		}
	}
	return vals
}

func padAddr(n int) string {
	s := make([]byte, 40)
	for i := range s {
		s[i] = "0123456789abcdef"[(n+i)%16]
	}
	return string(s)
}

func TestStandardMerkleTreeNew(t *testing.T) {
	vals := airdropData(4)
	tree, err := gomerk.NewStandardMerkleTree(vals, []string{"address", "uint256"}, true)
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

func TestStandardMerkleTreeSingle(t *testing.T) {
	tree, err := gomerk.NewStandardMerkleTree(airdropData(1), []string{"address", "uint256"}, true)
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

func TestStandardMerkleTreeLeafEncoding(t *testing.T) {
	enc := []string{"address", "uint256"}
	tree, _ := gomerk.NewStandardMerkleTree(airdropData(4), enc, true)
	if !slices.Equal(tree.LeafEncoding(), enc) {
		t.Error("LeafEncoding mismatch")
	}
}

func TestStandardMerkleTreeAt(t *testing.T) {
	vals := airdropData(4)
	tree, _ := gomerk.NewStandardMerkleTree(vals, []string{"address", "uint256"}, true)

	for i := 0; i < tree.Len(); i++ {
		v, ok := tree.At(i)
		if !ok {
			t.Errorf("At(%d) should exist", i)
		}
		if len(v) != 2 {
			t.Errorf("At(%d) should have 2 elements", i)
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

func TestStandardMerkleTreeAll(t *testing.T) {
	vals := airdropData(4)
	tree, _ := gomerk.NewStandardMerkleTree(vals, []string{"address", "uint256"}, true)

	count := 0
	for range tree.All() {
		count++
	}
	if count != 4 {
		t.Errorf("got %d, want 4", count)
	}
}

func TestStandardMerkleTreeGetProof(t *testing.T) {
	vals := airdropData(8)
	tree, _ := gomerk.NewStandardMerkleTree(vals, []string{"address", "uint256"}, true)

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

func TestStandardMerkleTreeGetProofByIndex(t *testing.T) {
	vals := airdropData(8)
	tree, _ := gomerk.NewStandardMerkleTree(vals, []string{"address", "uint256"}, true)

	for i := 0; i < tree.Len(); i++ {
		proof, err := tree.GetProofByIndex(i)
		if err != nil {
			t.Fatal(err)
		}
		v, _ := tree.At(i)
		ok, _ := tree.Verify(v, proof)
		if !ok {
			t.Error("verify by index failed")
		}
	}
}

func TestStandardMerkleTreeGetProofOutOfBounds(t *testing.T) {
	tree, _ := gomerk.NewStandardMerkleTree(airdropData(4), []string{"address", "uint256"}, true)
	_, err := tree.GetProofByIndex(-1)
	if err != gomerk.ErrIndexOutOfBounds {
		t.Errorf("got %v, want ErrIndexOutOfBounds", err)
	}
}

func TestStandardMerkleTreeLeafNotInTree(t *testing.T) {
	tree, _ := gomerk.NewStandardMerkleTree(airdropData(4), []string{"address", "uint256"}, true)
	_, err := tree.GetProof([]any{"0x9999999999999999999999999999999999999999", 9999})
	if err != gomerk.ErrLeafNotInTree {
		t.Errorf("got %v, want ErrLeafNotInTree", err)
	}
}

func TestStandardMerkleTreeStaticVerify(t *testing.T) {
	vals := airdropData(4)
	enc := []string{"address", "uint256"}
	tree, _ := gomerk.NewStandardMerkleTree(vals, enc, true)

	for _, v := range vals {
		proof, _ := tree.GetProof(v)
		ok, err := gomerk.VerifyStandard(tree.Root(), enc, v, proof)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Error("static verify failed")
		}
	}
}

func TestStandardMerkleTreeRejectInvalidProof(t *testing.T) {
	vals1 := airdropData(4)
	tree1, _ := gomerk.NewStandardMerkleTree(vals1, []string{"address", "uint256"}, true)

	vals2 := make([][]any, 4)
	for i := range vals2 {
		vals2[i] = []any{"0x" + padAddr(i+100), (i + 100) * 100}
	}
	tree2, _ := gomerk.NewStandardMerkleTree(vals2, []string{"address", "uint256"}, true)

	proof, _ := tree1.GetProof(vals1[0])
	ok, _ := tree2.Verify(vals1[0], proof)
	if ok {
		t.Error("should reject invalid proof")
	}
}

func TestStandardMerkleTreeMultiProof(t *testing.T) {
	vals := airdropData(8)
	tree, _ := gomerk.NewStandardMerkleTree(vals, []string{"address", "uint256"}, true)

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

func TestStandardMerkleTreeDump(t *testing.T) {
	vals := airdropData(4)
	tree, _ := gomerk.NewStandardMerkleTree(vals, []string{"address", "uint256"}, true)

	data := tree.Dump()
	if data.Format != "standard-v1" {
		t.Errorf("got %s, want standard-v1", data.Format)
	}
	if !slices.Equal(data.LeafEncoding, []string{"address", "uint256"}) {
		t.Error("LeafEncoding mismatch")
	}
}

func TestStandardMerkleTreeDumpLoad(t *testing.T) {
	vals := airdropData(4)
	tree, _ := gomerk.NewStandardMerkleTree(vals, []string{"address", "uint256"}, true)

	data := tree.Dump()
	js, _ := json.Marshal(data)

	var loaded gomerk.StandardTreeData
	json.Unmarshal(js, &loaded)

	tree2, err := gomerk.LoadStandardMerkleTree(loaded)
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

func TestStandardMerkleTreeLoadBadFormat(t *testing.T) {
	tests := []string{"nonstandard", "simple-v1", "bad"}
	for _, format := range tests {
		_, err := gomerk.LoadStandardMerkleTree(gomerk.StandardTreeData{Format: format})
		if err != gomerk.ErrInvalidFormat {
			t.Errorf("format %q: got %v, want ErrInvalidFormat", format, err)
		}
	}
}

func TestStandardMerkleTreeRender(t *testing.T) {
	tree, _ := gomerk.NewStandardMerkleTree(airdropData(4), []string{"address", "uint256"}, true)
	s, err := tree.Render()
	if err != nil {
		t.Fatal(err)
	}
	if s == "" {
		t.Error("render should not be empty")
	}
}

func TestStandardMerkleTreeUnsorted(t *testing.T) {
	vals := airdropData(4)
	tree, _ := gomerk.NewStandardMerkleTree(vals, []string{"address", "uint256"}, false)
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

// ABI Type Tests

func TestStandardMerkleTreeBytes32(t *testing.T) {
	vals := [][]any{
		{"0x1111111111111111111111111111111111111111111111111111111111111111", 100},
		{"0x2222222222222222222222222222222222222222222222222222222222222222", 200},
	}
	tree, err := gomerk.NewStandardMerkleTree(vals, []string{"bytes32", "uint256"}, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range vals {
		proof, _ := tree.GetProof(v)
		ok, _ := tree.Verify(v, proof)
		if !ok {
			t.Error("bytes32 verify failed")
		}
	}
}

func TestStandardMerkleTreeUintTypes(t *testing.T) {
	vals := [][]any{
		{100, 200, 50},
		{300, 400, 60},
	}
	tree, err := gomerk.NewStandardMerkleTree(vals, []string{"uint256", "uint128", "uint64"}, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range vals {
		proof, _ := tree.GetProof(v)
		ok, _ := tree.Verify(v, proof)
		if !ok {
			t.Error("uint types verify failed")
		}
	}
}

func TestStandardMerkleTreeBool(t *testing.T) {
	vals := [][]any{
		{true, 100},
		{false, 200},
	}
	tree, err := gomerk.NewStandardMerkleTree(vals, []string{"bool", "uint256"}, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range vals {
		proof, _ := tree.GetProof(v)
		ok, _ := tree.Verify(v, proof)
		if !ok {
			t.Error("bool verify failed")
		}
	}
}

func TestStandardMerkleTreeString(t *testing.T) {
	vals := [][]any{
		{"hello", 100},
		{"world", 200},
	}
	tree, err := gomerk.NewStandardMerkleTree(vals, []string{"string", "uint256"}, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range vals {
		proof, _ := tree.GetProof(v)
		ok, _ := tree.Verify(v, proof)
		if !ok {
			t.Error("string verify failed")
		}
	}
}

func TestStandardMerkleTreeBytes(t *testing.T) {
	vals := [][]any{
		{"0x1234", 100},
		{"0xabcd", 200},
	}
	tree, err := gomerk.NewStandardMerkleTree(vals, []string{"bytes", "uint256"}, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range vals {
		proof, _ := tree.GetProof(v)
		ok, _ := tree.Verify(v, proof)
		if !ok {
			t.Error("bytes verify failed")
		}
	}
}

func TestStandardMerkleTreeIntSigned(t *testing.T) {
	vals := [][]any{
		{-100, 100},
		{200, 200},
	}
	tree, err := gomerk.NewStandardMerkleTree(vals, []string{"int256", "uint256"}, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range vals {
		proof, _ := tree.GetProof(v)
		ok, _ := tree.Verify(v, proof)
		if !ok {
			t.Error("int256 verify failed")
		}
	}
}

func TestStandardMerkleTreeLarge(t *testing.T) {
	vals := airdropData(100)
	tree, err := gomerk.NewStandardMerkleTree(vals, []string{"address", "uint256"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if tree.Len() != 100 {
		t.Errorf("got %d, want 100", tree.Len())
	}
	if err := tree.Validate(); err != nil {
		t.Fatal(err)
	}

	// Sample verification
	for _, i := range []int{0, 25, 50, 75, 99} {
		proof, _ := tree.GetProofByIndex(i)
		v, _ := tree.At(i)
		ok, _ := tree.Verify(v, proof)
		if !ok {
			t.Errorf("large tree verify at %d failed", i)
		}
	}
}

func TestMultiProofJSON(t *testing.T) {
	mp := &gomerk.MultiProof{
		Leaves:     []string{"0x0000000000000000000000000000000000000000000000000000000000000001"},
		Proof:      []string{"0x0000000000000000000000000000000000000000000000000000000000000002"},
		ProofFlags: []bool{true, false},
	}

	js, err := json.Marshal(mp)
	if err != nil {
		t.Fatal(err)
	}

	var loaded gomerk.MultiProof
	if err := json.Unmarshal(js, &loaded); err != nil {
		t.Fatal(err)
	}

	if len(loaded.Leaves) != 1 || len(loaded.Proof) != 1 || len(loaded.ProofFlags) != 2 {
		t.Error("JSON roundtrip failed")
	}
}
