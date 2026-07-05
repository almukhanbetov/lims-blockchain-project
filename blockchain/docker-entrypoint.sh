#!/bin/sh
set -e

npx hardhat node --hostname 0.0.0.0 &
NODE_PID=$!

echo "Waiting for Hardhat node to accept connections..."
node wait-for-node.js
echo "Hardhat node is ready."

# Deploying here is the very first transaction on this fresh chain, sent from the
# default account (index 0) at nonce 0 — that makes the resulting contract address
# deterministic: 0x5FbDB2315678afecb367f032d93F642f64180aa3.
echo "Deploying LimsHashRegistry..."
npx hardhat run scripts/deploy.js --network localhost
touch /tmp/deployed

wait "$NODE_PID"
