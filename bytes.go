package gomerk

import (
	"bytes"
	"encoding/hex"
	"strings"
)

type Bytes32 [32]byte

func (b Bytes32) Hex() string           { return "0x" + hex.EncodeToString(b[:]) }
func (b Bytes32) String() string        { return b.Hex() }
func (b Bytes32) IsZero() bool          { return b == Bytes32{} }
func (a Bytes32) Compare(b Bytes32) int { return bytes.Compare(a[:], b[:]) }
func (a Bytes32) Less(b Bytes32) bool   { return a.Compare(b) < 0 }

func HexToBytes32(s string) (b Bytes32, err error) {
	data, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil {
		return b, ErrInvalidHex
	}
	if len(data) != 32 {
		return b, ErrInvalidNodeLength
	}
	return Bytes32(data), nil
}

func MustHexToBytes32(s string) Bytes32 {
	b, err := HexToBytes32(s)
	if err != nil {
		panic(err)
	}
	return b
}

func ConcatSorted(a, b Bytes32) []byte {
	if a.Less(b) {
		return append(a[:], b[:]...)
	}
	return append(b[:], a[:]...)
}
