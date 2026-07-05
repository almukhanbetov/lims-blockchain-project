package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// registryABI covers only the LimsHashRegistry functions the adapter calls.
// Hand-written to match contracts/LimsHashRegistry.sol — kept minimal on purpose,
// this avoids depending on solc/abigen at adapter build time.
const registryABI = `[
  {
    "type": "function",
    "name": "registerHash",
    "stateMutability": "nonpayable",
    "inputs": [
      {"name": "limsRecordId", "type": "string"},
      {"name": "eventType", "type": "string"},
      {"name": "dataHash", "type": "bytes32"},
      {"name": "status", "type": "string"}
    ],
    "outputs": []
  },
  {
    "type": "function",
    "name": "verifyHash",
    "stateMutability": "view",
    "inputs": [
      {"name": "limsRecordId", "type": "string"},
      {"name": "dataHash", "type": "bytes32"}
    ],
    "outputs": [{"name": "verified", "type": "bool"}]
  }
]`

// BlockchainClient talks to the LimsHashRegistry smart contract over JSON-RPC.
// When it can't be constructed (bad RPC URL, bad key), the adapter falls back
// to the in-memory registry so the demo keeps working without a live chain.
type BlockchainClient struct {
	client   *ethclient.Client
	contract *bind.BoundContract
	auth     *bind.TransactOpts
	chainID  *big.Int

	// txMu serializes writes: auth's nonce is assigned per-call from pending state,
	// so concurrent RegisterHash calls without this would race on the same nonce.
	txMu sync.Mutex
}

func NewBlockchainClient(rpcURL, contractAddress, privateKeyHex string) (*BlockchainClient, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", rpcURL, err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("fetch chain id: %w", err)
	}

	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("create transactor: %w", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(registryABI))
	if err != nil {
		return nil, fmt.Errorf("parse abi: %w", err)
	}

	contract := bind.NewBoundContract(common.HexToAddress(contractAddress), parsedABI, client, client, client)

	return &BlockchainClient{
		client:   client,
		contract: contract,
		auth:     auth,
		chainID:  chainID,
	}, nil
}

func hashToBytes32(hexHash string) ([32]byte, error) {
	var out [32]byte
	decoded, err := hex.DecodeString(hexHash)
	if err != nil {
		return out, fmt.Errorf("decode hash: %w", err)
	}
	if len(decoded) != 32 {
		return out, fmt.Errorf("hash must be 32 bytes, got %d", len(decoded))
	}
	copy(out[:], decoded)
	return out, nil
}

// RegisterHash submits the event hash to the LimsHashRegistry contract and
// returns the transaction hash once it has been broadcast.
func (b *BlockchainClient) RegisterHash(limsRecordID, eventType, dataHash, status string) (string, error) {
	hashBytes, err := hashToBytes32(dataHash)
	if err != nil {
		return "", err
	}

	b.txMu.Lock()
	defer b.txMu.Unlock()

	tx, err := b.contract.Transact(b.auth, "registerHash", limsRecordID, eventType, hashBytes, status)
	if err != nil {
		return "", fmt.Errorf("registerHash tx: %w", err)
	}
	return tx.Hash().Hex(), nil
}

// VerifyHash asks the contract whether dataHash matches what is on chain for limsRecordID.
func (b *BlockchainClient) VerifyHash(limsRecordID, dataHash string) (bool, error) {
	hashBytes, err := hashToBytes32(dataHash)
	if err != nil {
		return false, err
	}

	var results []interface{}
	callOpts := &bind.CallOpts{Context: context.Background()}
	if err := b.contract.Call(callOpts, &results, "verifyHash", limsRecordID, hashBytes); err != nil {
		return false, fmt.Errorf("verifyHash call: %w", err)
	}
	if len(results) != 1 {
		return false, fmt.Errorf("unexpected verifyHash result shape: %d values", len(results))
	}
	verified, ok := results[0].(bool)
	if !ok {
		return false, fmt.Errorf("unexpected verifyHash result type %T", results[0])
	}
	return verified, nil
}
