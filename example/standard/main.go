// Standard Merkle Tree Example
//
// Demonstrates StandardMerkleTree with ABI-encoded structured data.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/pyroth/gomerk"
)

var encoding = []string{"address", "uint256"}

func main() {
	values := [][]any{
		{"0x1111111111111111111111111111111111111111", "1000000000000000000"},
		{"0x2222222222222222222222222222222222222222", "2000000000000000000"},
		{"0x3333333333333333333333333333333333333333", "3000000000000000000"},
		{"0x4444444444444444444444444444444444444444", "4000000000000000000"},
	}

	tree := must(gomerk.NewStandardMerkleTree(values, encoding, true))

	fmt.Printf("Root:   %s\n", tree.Root())
	fmt.Printf("Leaves: %d\n\n", tree.Len())

	// Generate and verify proof
	proof := must(tree.GetProof(values[0]))
	fmt.Printf("Proof for %s:\n%v\n\n", values[0][0], proof)

	fmt.Printf("Tree verify:   %v\n", must(tree.Verify(values[0], proof)))
	fmt.Printf("Static verify: %v\n\n", must(gomerk.VerifyStandard(tree.Root(), encoding, values[0], proof)))

	// Serialize
	os.WriteFile("tree.json", must(json.MarshalIndent(tree.Dump(), "", "  ")), 0644)
	fmt.Println("Tree saved to tree.json")
}

func must[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return v
}
