// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/utils/cryptography/MerkleProof.sol";

/// @title MerkleAirdrop
/// @notice Token airdrop contract using Merkle proofs for verification.
contract MerkleAirdrop {
    using SafeERC20 for IERC20;

    IERC20 public immutable token;
    bytes32 public immutable merkleRoot;
    mapping(address => bool) public claimed;

    event Claimed(address indexed account, uint256 amount);

    error AlreadyClaimed();
    error InvalidProof();

    constructor(address token_, bytes32 merkleRoot_) {
        token = IERC20(token_);
        merkleRoot = merkleRoot_;
    }

    /// @notice Claim airdrop tokens with merkle proof.
    /// @param amount Token amount to claim
    /// @param proof Merkle proof for verification
    function claim(uint256 amount, bytes32[] calldata proof) external {
        if (claimed[msg.sender]) revert AlreadyClaimed();

        bytes32 leaf = keccak256(bytes.concat(keccak256(abi.encode(msg.sender, amount))));
        if (!MerkleProof.verify(proof, merkleRoot, leaf)) revert InvalidProof();

        claimed[msg.sender] = true;
        token.safeTransfer(msg.sender, amount);

        emit Claimed(msg.sender, amount);
    }

    /// @notice Check if an address can claim.
    /// @param account Address to check
    /// @param amount Expected claim amount
    /// @param proof Merkle proof
    /// @return canClaim True if claim would succeed
    function canClaim(
        address account,
        uint256 amount,
        bytes32[] calldata proof
    ) external view returns (bool canClaim) {
        if (claimed[account]) return false;
        bytes32 leaf = keccak256(bytes.concat(keccak256(abi.encode(account, amount))));
        return MerkleProof.verify(proof, merkleRoot, leaf);
    }
}
