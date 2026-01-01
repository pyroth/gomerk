package gomerk

import "golang.org/x/crypto/sha3"

func Keccak256(data []byte) (h Bytes32) {
	d := sha3.NewLegacyKeccak256()
	d.Write(data)
	d.Sum(h[:0])
	return
}

func HashLeaf(data []byte) Bytes32  { h := Keccak256(data); return Keccak256(h[:]) }
func HashNode(a, b Bytes32) Bytes32 { return Keccak256(ConcatSorted(a, b)) }
