---
name: lims-adapter-go
description: Use this skill when creating or modifying the Go Gin adapter service for the LIMS blockchain project. The adapter connects LIMS with blockchain, calculates hashes, registers events, and verifies trust.
---

# LIMS Adapter Go Skill

## Purpose

Create a Go Gin microservice named `adapter`.

The adapter is the intelligent control layer between LIMS and blockchain.

```text
LIMS → Adapter → Blockchain
UI   → Adapter → Blockchain
```

## Main responsibilities

The adapter must:

1. Receive laboratory events.
2. Calculate deterministic SHA-256 hash.
3. Register hash in blockchain smart contract.
4. Return registry status.
5. Verify current laboratory data against stored blockchain hash.
6. Detect data changes.

## Required API endpoints

Create these endpoints:

```text
GET  /
POST /events/hash
GET  /events
POST /events/verify
```

Optional future endpoints:

```text
GET  /health
GET  /events/:lims_record_id
POST /events/from-lims
```

## Event model

Use this JSON structure:

```json
{
  "lims_record_id": "SAMPLE-2026-001",
  "event_type": "RESULT_VERIFIED",
  "sample_id": "S-001",
  "result": "pH=7.2",
  "user_id": "lab_user_01",
  "status": "VERIFIED",
  "timestamp": "2026-06-11T15:00:00Z"
}
```

## Hash rule

Calculate hash from stable ordered fields:

```text
lims_record_id + event_type + sample_id + result + user_id + status + timestamp
```

Use SHA-256.

The same event data must always produce the same hash.

## Go dependencies

Use:

```text
github.com/gin-gonic/gin
github.com/joho/godotenv
```

When real blockchain integration is added, use Ethereum client libraries if needed.

## Environment variables

Use `.env`:

```env
PORT=8090
SENAITE_URL=http://senaite:8080
BLOCKCHAIN_RPC=http://blockchain:8545
CONTRACT_ADDRESS=
```

## Implementation rules

- Keep LIMS data in LIMS.
- Do not store full confidential laboratory data in blockchain.
- Store only hash and metadata.
- Keep adapter stateless where possible.
- Return clear JSON responses.
- Add CORS if React UI cannot connect.
- Use clean structs for request and response data.

## Verification logic

When verifying:

1. Receive event data.
2. Calculate hash again.
3. Compare with stored blockchain hash.
4. Return:

```json
{
  "verified": true,
  "message": "Data integrity verified",
  "hash": "..."
}
```

or:

```json
{
  "verified": false,
  "message": "Hash mismatch. Data may have been changed",
  "hash": "..."
}
```

## Docker rule

Adapter must run in Docker with this service name:

```text
adapter
```

Expose port:

```text
8090
```
