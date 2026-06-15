---
name: lims-blockchain-architecture
description: Use this skill when designing or explaining the LIMS + blockchain doctoral prototype architecture where LIMS stores laboratory data, adapter connects LIMS with blockchain, and blockchain stores hashes for trust verification.
---

# LIMS Blockchain Architecture Skill

## Core idea

The system must not replace LIMS with blockchain.

The correct architecture is:

```text
LIMS stores laboratory data.
Adapter connects LIMS with blockchain.
Blockchain stores hashes and verifies trust.
```

The blockchain layer is a trust and integrity verification layer, not the main database.

## Main architecture

```text
React UI Dashboard
        ↓
Go Gin Adapter / Intelligent Control Layer
        ↓
Smart Contract
        ↓
Blockchain Network

SENAITE LIMS ←→ Go Gin Adapter
```

## Responsibilities

### 1. LIMS

Use SENAITE LIMS as the base LIMS.

LIMS stores:

- samples
- laboratory results
- users
- instruments
- reports
- workflow statuses
- audit records

Do not store full laboratory data in blockchain.

### 2. Adapter

The adapter is the intelligent integration layer.

It must:

- receive laboratory events from LIMS or UI
- normalize event data
- choose whether the event is important enough for blockchain registration
- calculate SHA-256 hash
- send the hash to the smart contract
- verify current LIMS data against blockchain hash
- return verification result to UI

### 3. Blockchain

Blockchain stores only:

- lims_record_id
- event_type
- data_hash
- timestamp
- status
- previous_hash if needed

Blockchain must not store confidential full laboratory results.

## Blockchain integration points

Use these event types:

```text
SAMPLE_CREATED
RESULT_SUBMITTED
RESULT_VERIFIED
REPORT_PUBLISHED
DATA_MODIFIED
INSTRUMENT_IMPORT
```

## Scientific novelty focus

When generating code or explanations, emphasize:

- integration architecture
- intelligent control layer
- hash-based integrity verification
- blockchain insertion points in LIMS workflow
- separation between data storage and trust verification

## Recommended project structure

```text
lims-blockchain-project/
│
├── docker-compose.yml
│
├── adapter/
│   ├── Dockerfile
│   ├── go.mod
│   ├── main.go
│   └── .env
│
├── blockchain/
│   ├── package.json
│   ├── hardhat.config.js
│   ├── contracts/
│   │   └── LimsHashRegistry.sol
│   └── scripts/
│       └── deploy.js
│
└── ui/
    ├── Dockerfile
    ├── package.json
    ├── index.html
    ├── vite.config.js
    └── src/
        ├── main.jsx
        ├── App.jsx
        └── style.css
```

## Important rule

Always preserve this principle:

```text
Data remains in LIMS.
Proof of integrity is stored in blockchain.
Adapter connects both layers.
```
