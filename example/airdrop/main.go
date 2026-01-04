// Merkle Airdrop Example
//
// Complete token airdrop workflow: generate tree, export proofs, serve API.
package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/pyroth/gomerk"
)

var encoding = []string{"address", "uint256"}

func main() {
	cmd := flag.String("cmd", "generate", "Command: generate|serve")
	csvFile := flag.String("csv", "airdrop.csv", "Input CSV file")
	treeFile := flag.String("tree", "airdrop-tree.json", "Tree output file")
	proofsFile := flag.String("proofs", "airdrop-proofs.json", "Proofs output file")
	addr := flag.String("addr", ":8080", "Server address")
	flag.Parse()

	switch *cmd {
	case "generate":
		generate(*csvFile, *treeFile, *proofsFile)
	case "serve":
		serve(*treeFile, *addr)
	default:
		log.Fatalf("Unknown command: %s", *cmd)
	}
}

// generate builds merkle tree from CSV and exports proofs.
func generate(csvPath, treePath, proofsPath string) {
	// Load recipients
	recipients := must(loadCSV(csvPath))
	fmt.Printf("Loaded %d recipients\n", len(recipients))

	// Build tree
	tree := must(gomerk.NewStandardMerkleTree(recipients, encoding, true))
	fmt.Printf("Merkle Root: %s\n", tree.Root())

	// Save tree
	os.WriteFile(treePath, must(json.MarshalIndent(tree.Dump(), "", "  ")), 0644)
	fmt.Printf("Tree saved to %s\n", treePath)

	// Generate all proofs
	proofs := make(map[string]ProofData)
	for i, v := range tree.All() {
		addr := v[0].(string)
		proofs[strings.ToLower(addr)] = ProofData{
			Address: addr,
			Amount:  v[1].(string),
			Proof:   must(tree.GetProofByIndex(i)),
		}
	}

	os.WriteFile(proofsPath, must(json.MarshalIndent(proofs, "", "  ")), 0644)
	fmt.Printf("Proofs saved to %s\n", proofsPath)
}

// serve starts HTTP API for proof queries.
func serve(treePath, addr string) {
	data := must(os.ReadFile(treePath))
	var treeData gomerk.StandardTreeData
	must0(json.Unmarshal(data, &treeData))
	tree := must(gomerk.LoadStandardMerkleTree(treeData))

	fmt.Printf("Loaded tree with %d leaves\n", tree.Len())
	fmt.Printf("Root: %s\n", tree.Root())

	// Build index for O(1) lookup
	index := make(map[string]int)
	for i, v := range tree.All() {
		index[strings.ToLower(v[0].(string))] = i
	}

	http.HandleFunc("/proof/", func(w http.ResponseWriter, r *http.Request) {
		addr := strings.ToLower(strings.TrimPrefix(r.URL.Path, "/proof/"))
		i, ok := index[addr]
		if !ok {
			http.Error(w, `{"error":"address not found"}`, http.StatusNotFound)
			return
		}

		v, _ := tree.At(i)
		proof, _ := tree.GetProofByIndex(i)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ProofData{
			Address: v[0].(string),
			Amount:  v[1].(string),
			Proof:   proof,
		})
	})

	http.HandleFunc("/root", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"root": tree.Root()})
	})

	fmt.Printf("Server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

type ProofData struct {
	Address string   `json:"address"`
	Amount  string   `json:"amount"`
	Proof   []string `json:"proof"`
}

func loadCSV(path string) ([][]any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}

	var result [][]any
	for _, row := range records[1:] { // skip header
		result = append(result, []any{row[0], row[1]})
	}
	return result, nil
}

func must[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return v
}

func must0(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
