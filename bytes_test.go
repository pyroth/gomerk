package gomerk_test

import (
	"testing"

	"github.com/pyroth/gomerk"
)

func TestBytes32Hex(t *testing.T) {
	var b gomerk.Bytes32
	for i := range b {
		b[i] = byte(i)
	}
	got := b.Hex()
	want := "0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestBytes32String(t *testing.T) {
	b := gomerk.Bytes32{0xff}
	if b.String() != b.Hex() {
		t.Error("String() should equal Hex()")
	}
}

func TestBytes32IsZero(t *testing.T) {
	var zero gomerk.Bytes32
	if !zero.IsZero() {
		t.Error("zero value should be zero")
	}
	nonZero := gomerk.Bytes32{1}
	if nonZero.IsZero() {
		t.Error("non-zero value should not be zero")
	}
}

func TestBytes32Compare(t *testing.T) {
	tests := []struct {
		a, b gomerk.Bytes32
		want int
	}{
		{gomerk.Bytes32{}, gomerk.Bytes32{}, 0},
		{gomerk.Bytes32{}, gomerk.Bytes32{0: 1}, -1},
		{gomerk.Bytes32{0: 1}, gomerk.Bytes32{}, 1},
		{gomerk.Bytes32{31: 1}, gomerk.Bytes32{31: 2}, -1},
	}
	for _, tc := range tests {
		if got := tc.a.Compare(tc.b); got != tc.want {
			t.Errorf("Compare(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestBytes32Less(t *testing.T) {
	a, b := gomerk.Bytes32{}, gomerk.Bytes32{31: 1}
	if !a.Less(b) {
		t.Error("a should be less than b")
	}
	if b.Less(a) {
		t.Error("b should not be less than a")
	}
	if a.Less(a) {
		t.Error("a should not be less than itself")
	}
}

func TestHexToBytes32(t *testing.T) {
	tests := []struct {
		input string
		want  gomerk.Bytes32
		err   bool
	}{
		{"0x0000000000000000000000000000000000000000000000000000000000000001", gomerk.Bytes32{31: 1}, false},
		{"0000000000000000000000000000000000000000000000000000000000000001", gomerk.Bytes32{31: 1}, false},
		{"0x00", gomerk.Bytes32{}, true},
		{"invalid", gomerk.Bytes32{}, true},
	}
	for _, tc := range tests {
		got, err := gomerk.HexToBytes32(tc.input)
		if tc.err {
			if err == nil {
				t.Errorf("HexToBytes32(%q) expected error", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("HexToBytes32(%q) unexpected error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("HexToBytes32(%q) = %v, want %v", tc.input, got, tc.want)
			}
		}
	}
}

func TestHexRoundtrip(t *testing.T) {
	var orig gomerk.Bytes32
	for i := range orig {
		orig[i] = byte(i * 7)
	}
	got, err := gomerk.HexToBytes32(orig.Hex())
	if err != nil {
		t.Fatal(err)
	}
	if got != orig {
		t.Error("roundtrip failed")
	}
}

func TestMustHexToBytes32(t *testing.T) {
	b := gomerk.MustHexToBytes32("0x0000000000000000000000000000000000000000000000000000000000000001")
	if b[31] != 1 {
		t.Error("MustHexToBytes32 failed")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustHexToBytes32 should panic on invalid input")
		}
	}()
	gomerk.MustHexToBytes32("invalid")
}

func TestConcatSorted(t *testing.T) {
	a := gomerk.Bytes32{0: 1}
	b := gomerk.Bytes32{0: 2}

	// a < b, so result should be a || b
	result := gomerk.ConcatSorted(a, b)
	if len(result) != 64 {
		t.Errorf("got len %d, want 64", len(result))
	}
	if result[0] != 1 || result[32] != 2 {
		t.Error("ConcatSorted(a, b) order wrong")
	}

	// b > a, so result should still be a || b (sorted)
	result2 := gomerk.ConcatSorted(b, a)
	if result2[0] != 1 || result2[32] != 2 {
		t.Error("ConcatSorted(b, a) order wrong")
	}
}
