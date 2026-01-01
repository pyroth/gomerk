// Simple Merkle Tree Example
//
// Demonstrates basic usage of SimpleMerkleTree for Bytes32 values.
package main

import (
	"fmt"

	"github.com/pyroth/gomerk"
)

func main() {
	// Create leaf values (32-byte hashes)
	values := []gomerk.Bytes32{
		gomerk.MustHexToBytes32("0x0000000000000000000000000000000000000000000000000000000000000001"),
		gomerk.MustHexToBytes32("0x0000000000000000000000000000000000000000000000000000000000000002"),
		gomerk.MustHexToBytes32("0x0000000000000000000000000000000000000000000000000000000000000003"),
		gomerk.MustHexToBytes32("0x0000000000000000000000000000000000000000000000000000000000000004"),
	}

	// Build tree (sorted leaves for deterministic root)
	tree, err := gomerk.NewSimpleMerkleTree(values, true)
	if err != nil {
		panic(err)
	}

	fmt.Println("=== Simple Merkle Tree ===")
	fmt.Println("Root:", tree.Root())
	fmt.Println("Leaves:", tree.Len())

	// Generate and verify proof
	leaf := values[0]
	proof, _ := tree.GetProof(leaf)
	fmt.Println("\nProof for leaf 0:", proof)

	valid, _ := tree.Verify(leaf, proof)
	fmt.Println("Verified:", valid)

	// Static verification (without tree instance)
	valid, _ = gomerk.VerifySimple(tree.Root(), leaf, proof)
	fmt.Println("Static verify:", valid)

	// Iterate over all values
	fmt.Println("\nAll values:")
	for i, v := range tree.All() {
		fmt.Printf("  [%d] %s\n", i, v)
	}

	// Render tree structure
	fmt.Println("\nTree structure:")
	rendered, _ := tree.Render()
	fmt.Println(rendered)
}
