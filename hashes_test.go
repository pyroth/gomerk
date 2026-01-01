package gomerk_test

import (
	"testing"

	"github.com/pyroth/gomerk"
)

func TestKeccak256(t *testing.T) {
	tests := []struct {
		input []byte
		want  string
	}{
		{[]byte("hello"), "0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8"},
		{nil, "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"},
		{[]byte{}, "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"},
		{[]byte("abc"), "0x4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45"},
	}
	for _, tc := range tests {
		got := gomerk.Keccak256(tc.input).Hex()
		if got != tc.want {
			t.Errorf("Keccak256(%q) = %s, want %s", tc.input, got, tc.want)
		}
	}
}

func TestHashLeaf(t *testing.T) {
	// hashLeaf should double-hash: keccak256(keccak256(data))
	data := []byte{1, 2, 3}
	first := gomerk.Keccak256(data)
	want := gomerk.Keccak256(first[:])
	got := gomerk.HashLeaf(data)
	if got != want {
		t.Error("hashLeaf should be double keccak256")
	}
}

func TestHashNodeCommutative(t *testing.T) {
	tests := []struct {
		a, b gomerk.Bytes32
	}{
		{gomerk.Bytes32{1}, gomerk.Bytes32{2}},
		{gomerk.Bytes32{0xff}, gomerk.Bytes32{0x00}},
		{gomerk.Bytes32{31: 1}, gomerk.Bytes32{31: 2}},
	}
	for _, tc := range tests {
		h1 := gomerk.HashNode(tc.a, tc.b)
		h2 := gomerk.HashNode(tc.b, tc.a)
		if h1 != h2 {
			t.Errorf("hashNode not commutative for %v, %v", tc.a, tc.b)
		}
	}
}

func TestHashNodeDifferentInputs(t *testing.T) {
	a := gomerk.Bytes32{1}
	b := gomerk.Bytes32{2}
	c := gomerk.Bytes32{3}

	h1 := gomerk.HashNode(a, b)
	h2 := gomerk.HashNode(a, c)
	if h1 == h2 {
		t.Error("different inputs should produce different hashes")
	}
}
