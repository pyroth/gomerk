# Merkle Airdrop Example

Complete token airdrop workflow using Merkle proofs.

## Usage

### 1. Generate Merkle Tree

```bash
go run main.go -cmd=generate
```

Output:

- `airdrop-tree.json` - Full tree for loading
- `airdrop-proofs.json` - Pre-generated proofs for all addresses
- Merkle root for smart contract deployment

### 2. Deploy Contract

Deploy `contract.sol` with:

- Token address
- Merkle root from step 1

### 3. Serve Proof API

```bash
go run main.go -cmd=serve -addr=:8080
```

Endpoints:

- `GET /root` - Returns merkle root
- `GET /proof/{address}` - Returns proof for address

### Example API Response

```json
{
  "address": "0x1111111111111111111111111111111111111111",
  "amount": "1000000000000000000",
  "proof": [
    "0x...",
    "0x..."
  ]
}
```

## Files

| File | Description |
| ---- | ----------- |
| `main.go` | CLI tool for generation and serving |
| `airdrop.csv` | Sample recipient list |
| `contract.sol` | Solidity airdrop contract |
