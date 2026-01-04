# gomerk

**A Go library to generate merkle trees and merkle proofs.**

Well suited for airdrops and similar mechanisms in combination with Solidity [`OpenZeppelin MerkleProof`] utilities.

[`OpenZeppelin MerkleProof`]: https://docs.openzeppelin.com/contracts/4.x/api/utils#MerkleProof

[![Go Reference](https://pkg.go.dev/badge/github.com/pyroth/gomerk.svg)](https://pkg.go.dev/github.com/pyroth/gomerk)

## Quick Start

```bash
go get github.com/pyroth/gomerk
```

### Building a Tree

```go
values := [][]any{
    {"0x1111111111111111111111111111111111111111", "5000000000000000000"},
    {"0x2222222222222222222222222222222222222222", "2500000000000000000"},
}

tree, _ := gomerk.NewStandardMerkleTree(values, []string{"address", "uint256"}, true)
fmt.Println("Root:", tree.Root())

// Serialize for distribution
jsonBytes, _ := json.MarshalIndent(tree.Dump(), "", "  ")
os.WriteFile("tree.json", jsonBytes, 0644)
```

### Obtaining a Proof

```go
// Load tree
jsonData, _ := os.ReadFile("tree.json")
var data gomerk.StandardTreeData
json.Unmarshal(jsonData, &data)
tree, _ := gomerk.LoadStandardMerkleTree(data)

// Get proof by value
proof, _ := tree.GetProof([]any{"0x1111111111111111111111111111111111111111", "5000000000000000000"})
fmt.Println("Proof:", proof)
```

### Validating a Proof in Solidity

Once the proof has been generated, it can be validated in Solidity using [`OpenZeppelin MerkleProof`] as in the following example:

```solidity
pragma solidity ^0.8.4;

import "@openzeppelin/contracts/utils/cryptography/MerkleProof.sol";

contract Verifier {
    bytes32 private root;

    constructor(bytes32 _root) {
        // (1)
        root = _root;
    }

    function verify(
        bytes32[] memory proof,
        address addr,
        uint256 amount
    ) public {
        // (2)
        bytes32 leaf = keccak256(bytes.concat(keccak256(abi.encode(addr, amount))));
        // (3)
        require(MerkleProof.verify(proof, root, leaf), "Invalid proof");
        // (4)
        // ...
    }
}
```

1. Store the tree root in your contract.
2. Compute the [leaf hash](#leaf-hash) for the provided `addr` and `amount` ABI encoded values.
3. Verify it using [`OpenZeppelin MerkleProof`]'s `verify` function.
4. Use the verification to make further operations on the contract. (Consider you may want to add a mechanism to prevent reuse of a leaf).

## Standard Merkle Trees

This library works on "standard" merkle trees designed for Ethereum smart contracts. We have defined them with a few characteristics that make them secure and good for on-chain verification.

- The tree is shaped as a [complete binary tree](https://xlinux.nist.gov/dads/HTML/completeBinaryTree.html).
- The leaves are sorted.
- The leaves are the result of ABI encoding a series of values.
- The hash used is Keccak256.
- The leaves are double-hashed[^1] to prevent [second preimage attacks].

[second preimage attacks]: https://flawed.net.nz/2018/02/21/attacking-merkle-trees-with-a-second-preimage-attack/

## Simple Merkle Trees

The library also supports "simple" merkle trees, which are a simplified version of the standard ones. They are designed to be more flexible and accept arbitrary `bytes32` data as leaves. It keeps the same tree shape and internal pair hashing algorithm.

As opposed to standard trees, leaves are not double-hashed. Instead they are hashed once and then hashed in pairs inside the tree. This is useful to override the leaf hashing algorithm and use a different one prior to building the tree.

Users of tooling that produced trees without double leaf hashing can use this feature to build a representation of the tree in Go. We recommend this approach exclusively for trees that are already built on-chain. Otherwise the standard tree may be a better fit.

```go
values := []gomerk.Bytes32{
    gomerk.Keccak256([]byte("Value 1")),
    gomerk.Keccak256([]byte("Value 2")),
}
tree, _ := gomerk.NewSimpleMerkleTree(values, true)
// SimpleMerkleTree shares the same API as StandardMerkleTree
```

## Advanced Usage

### Leaf Hash

The Standard Merkle Tree uses an opinionated double leaf hashing algorithm. For example, a leaf in the tree with value `[addr, amount]` can be computed in Solidity as follows:

```solidity
bytes32 leaf = keccak256(bytes.concat(keccak256(abi.encode(addr, amount))));
```

This is an opinionated design that we believe will offer the best out of the box experience for most users. However, there are advanced use cases where a different leaf hashing algorithm may be needed. For those, the `SimpleMerkleTree` can be used to build a tree with custom leaf hashing.

### Leaf Ordering

Each leaf of a merkle tree can be proven individually. The relative ordering of leaves is mostly irrelevant when the only objective is to prove the inclusion of individual leaves in the tree. Proving multiple leaves at once is however a little bit more difficult.

This library proposes a mechanism to prove (and verify) that sets of leaves are included in the tree. These "multiproofs" can also be verified onchain using the implementation available in `@openzeppelin/contracts`. This mechanism requires the leaves to be ordered respective to their position in the tree. For example, if the tree leaves are (in hex form) `[ 0xAA...AA, 0xBB...BB, 0xCC...CC, 0xDD...DD]`, then you'd be able to prove `[0xBB...BB, 0xDD...DD]` as a subset of the leaves, but not `[0xDD...DD, 0xBB...BB]`.

Since this library knows the entire tree, you can generate a multiproof with the requested leaves in any order. The library will re-order them so that they appear inside the proof in the correct order. The `MultiProof` object returned by `tree.GetMultiProofByIndices(...)` will have the leaves ordered according to their position in the tree, and not in the order in which you provided them.

By default, the library orders the leaves according to their hash when building the tree. This is so that a smart contract can build the hashes of a set of leaves and order them correctly without any knowledge of the tree itself. Said differently, it is simpler for a smart contract to process a multiproof for leaves that it rebuilt itself if the corresponding tree is ordered.

However, some trees are constructed iteratively from unsorted data, causing the leaves to be unsorted as well. For this library to be able to represent such trees, the call to `NewStandardMerkleTree` includes an option to disable sorting. Using that option, the leaves are kept in the order in which they were provided. Note that this option has no effect on your ability to generate and verify proofs and multiproofs in Go, but that it may introduce challenges when verifying multiproofs onchain. We recommend only using it for building a representation of trees that are built (onchain) using an iterative process.

## Examples

See the [`example/`](./example) directory for complete working examples:

- [`example/standard/`](./example/standard) - Standard Merkle Tree with ABI-encoded values
- [`example/simple/`](./example/simple) - Simple Merkle Tree with raw bytes32 values
- [`example/airdrop/`](./example/airdrop) - Complete token airdrop workflow with CLI and HTTP API

## License

MIT

[^1]: The underlying reason for hashing the leaves twice is to prevent the leaf values from being 64 bytes long _prior_ to hashing. Otherwise, the concatenation of a sorted pair of internal nodes in the Merkle tree could be reinterpreted as a leaf value. See [OpenZeppelin issue #3091](https://github.com/OpenZeppelin/openzeppelin-contracts/issues/3091) for more details.
