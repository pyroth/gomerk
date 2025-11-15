package main

import (
	"encoding/hex"
	"fmt"

	"github.com/pyroth/gomerk"
)

func main() {
	data := [][]byte{[]byte("data1"), []byte("data2"), []byte("data3")}
	tree, err := gomerk.NewMerkleTree(data)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	root := tree.RootHash()
	fmt.Println("Root Hash:", hex.EncodeToString(root))

	proof, err := tree.GenerateProof(1) // For "data2"
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	leaf := gomerk.HashLeaf([]byte("data2"))
	valid := gomerk.VerifyProof(proof, root, leaf, 1)
	fmt.Println("Proof Valid:", valid) // Should be true
}
