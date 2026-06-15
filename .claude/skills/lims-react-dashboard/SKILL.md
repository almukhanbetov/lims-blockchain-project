---
name: lims-react-dashboard
description: Use this skill when creating or modifying the React UI dashboard for the LIMS blockchain project. The UI shows hash registration, blockchain trust verification, and the architecture model.
---

# LIMS React Dashboard Skill

## Purpose

Create a React UI dashboard for the LIMS blockchain control layer.

The UI must show the idea clearly:

```text
LIMS stores laboratory data.
Adapter connects LIMS with blockchain.
Blockchain stores hash and verifies trust.
```

## UI pages

Create these pages or sections:

1. Dashboard
2. Register Event
3. Hash Registry
4. Verify Data
5. Architecture

## Dashboard must show

- total hash records
- verified events
- integrity model: SHA-256
- system diagram

Diagram:

```text
Lab PC / Sensors → SENAITE LIMS → Go Adapter → Smart Contract → Blockchain
```

## Register Event page

Create form fields:

```text
lims_record_id
event_type
sample_id
result
user_id
status
timestamp
```

Event types:

```text
SAMPLE_CREATED
RESULT_SUBMITTED
RESULT_VERIFIED
REPORT_PUBLISHED
DATA_MODIFIED
INSTRUMENT_IMPORT
```

The form sends POST request to:

```text
POST http://localhost:8090/events/hash
```

## Hash Registry page

Show table columns:

```text
LIMS Record ID
Event Type
Data Hash
Timestamp
Status
```

Load data from:

```text
GET http://localhost:8090/events
```

## Verify Data page

Use the same event form.

Send request to:

```text
POST http://localhost:8090/events/verify
```

Show result:

- verified true: green success block
- verified false: red warning block

## Architecture page

Explain four layers:

1. SENAITE LIMS — stores laboratory data.
2. Intelligent Control Layer — adapter calculates hash and decides what to register.
3. Smart Contract — stores hash and metadata.
4. Blockchain — trusted immutable verification layer.

## Technology

Use:

```text
React
Vite
Axios
CSS
Docker with Nginx
```

## Docker rule

UI must run as a separate service:

```text
ui
```

Expose:

```text
3000:80
```

## Design rule

Use a serious scientific dashboard style.

Avoid making it look like a toy project.

The dashboard should be suitable for dissertation demonstration.
