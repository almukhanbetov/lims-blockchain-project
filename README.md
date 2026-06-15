# LIMS Blockchain — Hash Registry Prototype

Doctoral research prototype demonstrating blockchain-based integrity verification for laboratory data.

**Architecture principle:** LIMS stores full laboratory data. Blockchain stores only SHA-256 hashes as proof of integrity. The adapter is the intelligent control layer connecting both.

```
Lab Instruments → SENAITE LIMS ↔ Go Adapter → LimsHashRegistry (Solidity) → Blockchain
                                      ↑
                                 React UI Dashboard
```

---

## Services

| Service         | Port  | Role                                      |
|----------------|-------|-------------------------------------------|
| SENAITE LIMS   | 8080  | Laboratory data storage                   |
| Hardhat Node   | 8545  | Local Ethereum blockchain (development)   |
| Go Adapter     | 8090  | Intelligent control layer (hash, verify)  |
| React UI       | 3000  | Dashboard (hash registry, verification)   |

---

## Quick Start — Full Docker Stack

```bash
# 1. Start all services
docker-compose up --build

# 2. Open dashboard
open http://localhost:3000

# 3. Adapter info
curl http://localhost:8090/
```

> SENAITE LIMS takes several minutes to initialise on first start.
> The adapter and UI work independently of SENAITE for demonstration.

---

## Blockchain Setup (Deploy Smart Contract)

```bash
cd blockchain

# Install Hardhat dependencies
npm install

# Start local Hardhat node (separate terminal)
npx hardhat node

# Deploy LimsHashRegistry contract
npx hardhat run scripts/deploy.js --network localhost
```

Copy the deployed contract address and set it in `adapter/.env`:

```env
CONTRACT_ADDRESS=0xYourContractAddressHere
```

Then restart the adapter:

```bash
docker-compose restart adapter
```

---

## Adapter API

### GET /
Returns adapter status and configuration.

```bash
curl http://localhost:8090/
```

### POST /events/hash
Calculate SHA-256 hash of a laboratory event and register it.

```bash
curl -X POST http://localhost:8090/events/hash \
  -H "Content-Type: application/json" \
  -d '{
    "lims_record_id": "SAMPLE-2026-001",
    "event_type": "RESULT_VERIFIED",
    "sample_id": "S-001",
    "result": "pH=7.2",
    "user_id": "lab_user_01",
    "status": "VERIFIED",
    "timestamp": "2026-06-15T12:00:00Z"
  }'
```

**Response:**
```json
{
  "success": true,
  "lims_record_id": "SAMPLE-2026-001",
  "event_type": "RESULT_VERIFIED",
  "data_hash": "a1b2c3d4...",
  "timestamp": "2026-06-15T12:00:00Z",
  "message": "Hash registered in blockchain"
}
```

### GET /events
Returns all registered hash records.

```bash
curl http://localhost:8090/events
```

### POST /events/verify
Recalculate hash and compare with stored blockchain hash.
Send the **same data** that was registered. A mismatch detects tampering.

```bash
curl -X POST http://localhost:8090/events/verify \
  -H "Content-Type: application/json" \
  -d '{
    "lims_record_id": "SAMPLE-2026-001",
    "event_type": "RESULT_VERIFIED",
    "sample_id": "S-001",
    "result": "pH=7.2",
    "user_id": "lab_user_01",
    "status": "VERIFIED",
    "timestamp": "2026-06-15T12:00:00Z"
  }'
```

**Verified response:**
```json
{ "verified": true, "message": "Data integrity verified", "hash": "a1b2c3d4..." }
```

**Tampered data response:**
```json
{ "verified": false, "message": "Hash mismatch. Data may have been changed", ... }
```

---

## Blockchain Event Types

| Event Type        | Trigger                          |
|-------------------|----------------------------------|
| SAMPLE_CREATED    | New sample received in LIMS      |
| RESULT_SUBMITTED  | Analyst submits test result      |
| RESULT_VERIFIED   | Result verified by supervisor    |
| REPORT_PUBLISHED  | Laboratory report published      |
| DATA_MODIFIED     | Any modification detected        |
| INSTRUMENT_IMPORT | Data imported from instrument    |

---

## Development — Adapter Only

```bash
cd adapter
go mod tidy
go run main.go
```

Requires Go 1.21+. The adapter runs on port 8090.

## Development — UI Only

```bash
cd ui
npm install
npm run dev
```

UI runs on http://localhost:5173 with proxy to adapter at http://localhost:8090.

---

## Hash Calculation

The adapter calculates SHA-256 from these fields in fixed order:

```
lims_record_id | event_type | sample_id | result | user_id | status | timestamp
```

The same input always produces the same hash. If any field changes, the hash changes — detected on next verification.

---

## Smart Contract Functions

```solidity
// Register hash in blockchain
function registerHash(string limsRecordId, string eventType, bytes32 dataHash, string status)

// Retrieve stored hash record
function getHash(string limsRecordId) returns (...)

// Verify integrity: returns true if hashes match
function verifyHash(string limsRecordId, bytes32 dataHash) returns (bool)
```

---

## Project Structure

```
lims-blockchain-project/
├── docker-compose.yml
├── adapter/
│   ├── Dockerfile
│   ├── go.mod
│   ├── main.go              ← intelligent control layer
│   └── .env
├── blockchain/
│   ├── package.json
│   ├── hardhat.config.js
│   ├── contracts/
│   │   └── LimsHashRegistry.sol   ← smart contract
│   └── scripts/
│       └── deploy.js
└── ui/
    ├── Dockerfile
    ├── nginx.conf
    ├── package.json
    ├── index.html
    ├── vite.config.js
    └── src/
        ├── main.jsx
        ├── App.jsx          ← dashboard, register, registry, verify, architecture
        └── style.css
```
