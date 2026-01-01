// Standard Merkle Tree Example
//
// Demonstrates StandardMerkleTree with ABI-encoded structured data.
package main

import (
	"encoding/json"
	"fmt"

	"github.com/pyroth/gomerk"
)

func main() {
	// Define values with multiple fields
	values := [][]any{
		{"0x1111111111111111111111111111111111111111", uint64(100)},
		{"0x2222222222222222222222222222222222222222", uint64(200)},
		{"0x3333333333333333333333333333333333333333", uint64(300)},
		{"0x4444444444444444444444444444444444444444", uint64(400)},
	}

	// Define ABI types for encoding
	encoding := []string{"address", "uint256"}

	// Build tree
	tree, err := gomerk.NewStandardMerkleTree(values, encoding, true)
	if err != nil {
		panic(err)
	}

	fmt.Println("=== Standard Merkle Tree ===")
	fmt.Println("Root:", tree.Root())
	fmt.Println("Encoding:", tree.LeafEncoding())
	fmt.Println("Leaves:", tree.Len())

	// Generate proof for first value
	proof, _ := tree.GetProof(values[0])
	fmt.Println("\nProof for values[0]:", proof)

	// Verify
	valid, _ := tree.Verify(values[0], proof)
	fmt.Println("Verified:", valid)

	// Static verification
	valid, _ = gomerk.VerifyStandard(tree.Root(), encoding, values[0], proof)
	fmt.Println("Static verify:", valid)

	// Serialize tree
	data := tree.Dump()
	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println("\nSerialized tree:")
	fmt.Println(string(jsonBytes))
}
