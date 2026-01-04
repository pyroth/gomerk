// Simple Merkle Tree Example
//
// Demonstrates SimpleMerkleTree with raw bytes32 values and custom hashing.
package main

import (
	"fmt"
	"log"

	"github.com/pyroth/gomerk"
)

func main() {
	// Custom hashed values
	values := []gomerk.Bytes32{
		gomerk.Keccak256([]byte("alice")),
		gomerk.Keccak256([]byte("bob")),
		gomerk.Keccak256([]byte("charlie")),
		gomerk.Keccak256([]byte("dave")),
	}

	tree := must(gomerk.NewSimpleMerkleTree(values, true))

	fmt.Printf("Root:   %s\n", tree.Root())
	fmt.Printf("Leaves: %d\n\n", tree.Len())

	// Generate and verify proof
	proof := must(tree.GetProof(values[0]))
	fmt.Printf("Proof for alice: %v\n\n", proof)

	fmt.Printf("Tree verify:   %v\n", must(tree.Verify(values[0], proof)))
	fmt.Printf("Static verify: %v\n\n", must(gomerk.VerifySimple(tree.Root(), values[0], proof)))

	// Tree visualization
	fmt.Println("Tree structure:")
	fmt.Println(must(tree.Render()))
}

func must[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return v
}
