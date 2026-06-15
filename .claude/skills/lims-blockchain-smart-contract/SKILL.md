---
name: lims-blockchain-smart-contract
description: Use this skill when creating or modifying Solidity smart contracts and Hardhat configuration for the LIMS blockchain project. Blockchain stores laboratory event hashes and verifies trust.
---

# LIMS Blockchain Smart Contract Skill

## Purpose

Create a blockchain layer that stores hashes of laboratory events.

The blockchain must not store full laboratory data.

It stores only proof of integrity.

```text
LIMS data → hash → smart contract → blockchain
```

## Smart contract name

Use:

```text
LimsHashRegistry.sol
```

## Contract responsibilities

The contract must:

1. Register event hash.
2. Store metadata.
3. Return hash by LIMS record ID.
4. Verify whether a submitted hash matches the stored hash.
5. Emit events for auditability.

## Data structure

Use a struct similar to:

```solidity
struct HashRecord {
    string limsRecordId;
    string eventType;
    bytes32 dataHash;
    uint256 timestamp;
    string status;
    address registeredBy;
}
```

## Required functions

Create functions:

```solidity
function registerHash(
    string memory limsRecordId,
    string memory eventType,
    bytes32 dataHash,
    string memory status
) public
```

```solidity
function getHash(string memory limsRecordId) public view returns (...)
```

```solidity
function verifyHash(string memory limsRecordId, bytes32 dataHash) public view returns (bool)
```

## Events

Emit:

```solidity
event HashRegistered(
    string limsRecordId,
    string eventType,
    bytes32 dataHash,
    uint256 timestamp,
    string status,
    address registeredBy
);
```

## Hardhat files

Use this structure:

```text
blockchain/
├── package.json
├── hardhat.config.js
├── contracts/
│   └── LimsHashRegistry.sol
└── scripts/
    └── deploy.js
```

## Hardhat commands

Use:

```bash
npm install
npx hardhat compile
npx hardhat node
npx hardhat run scripts/deploy.js --network localhost
```

## Scientific rule

Always explain that blockchain provides:

- immutability
- timestamped hash registration
- trusted audit trail
- integrity verification

But blockchain does not replace LIMS.

## Security rule

Do not store:

- patient data
- full laboratory results
- personal data
- confidential files

Store only hash and minimal metadata.
