package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// LimsEvent is the input model received from LIMS or UI.
// Full laboratory data stays in LIMS; only metadata is processed here.
type LimsEvent struct {
	LimsRecordID string `json:"lims_record_id" binding:"required"`
	EventType    string `json:"event_type"     binding:"required"`
	SampleID     string `json:"sample_id"`
	Result       string `json:"result"`
	UserID       string `json:"user_id"`
	Status       string `json:"status"`
	Timestamp    string `json:"timestamp"`
}

// HashRecord is what the adapter stores — mirrors what is sent to the smart contract.
// Blockchain stores only hash + metadata, never full laboratory data.
type HashRecord struct {
	LimsRecordID string    `json:"lims_record_id"`
	EventType    string    `json:"event_type"`
	DataHash     string    `json:"data_hash"`
	Status       string    `json:"status"`
	RegisteredAt time.Time `json:"registered_at"`
}

// In-memory registry acts as a fast-read cache of whatever is on chain.
// When CONTRACT_ADDRESS/BLOCKCHAIN_RPC are set and reachable, every write is
// also forwarded to the LimsHashRegistry smart contract via bc.
var (
	registry   = make(map[string]HashRecord)
	registryMu sync.RWMutex

	bc *BlockchainClient
)

// calculateHash produces a deterministic SHA-256 hash from stable ordered fields.
// The same event data will always produce the same hash — this is the core integrity guarantee.
func calculateHash(e LimsEvent) string {
	raw := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
		e.LimsRecordID,
		e.EventType,
		e.SampleID,
		e.Result,
		e.UserID,
		e.Status,
		e.Timestamp,
	)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// corsMiddleware enables cross-origin requests from the React UI.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file, reading environment variables directly")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	contractAddress := os.Getenv("CONTRACT_ADDRESS")
	blockchainRPC := os.Getenv("BLOCKCHAIN_RPC")
	privateKey := os.Getenv("ADAPTER_PRIVATE_KEY")

	if contractAddress != "" && blockchainRPC != "" && privateKey != "" {
		client, err := NewBlockchainClient(blockchainRPC, contractAddress, privateKey)
		if err != nil {
			log.Printf("Blockchain integration disabled, falling back to local registry: %v", err)
		} else {
			bc = client
			log.Printf("Blockchain integration active: RPC=%s Contract=%s", blockchainRPC, contractAddress)
		}
	} else {
		log.Println("Running in local registry mode (set CONTRACT_ADDRESS, BLOCKCHAIN_RPC and ADAPTER_PRIVATE_KEY to enable blockchain)")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(corsMiddleware())

	// GET / — adapter info and status
	r.GET("/", func(c *gin.Context) {
		registryMu.RLock()
		count := len(registry)
		registryMu.RUnlock()

		c.JSON(http.StatusOK, gin.H{
			"service":              "LIMS Blockchain Adapter",
			"version":              "1.0.0",
			"role":                 "Intelligent control layer between LIMS and blockchain",
			"integrity_model":      "SHA-256",
			"blockchain_rpc":       blockchainRPC,
			"contract_address":     contractAddress,
			"blockchain_connected": bc != nil,
			"registered_hashes":    count,
			"endpoints": []string{
				"GET  /",
				"POST /events/hash",
				"GET  /events",
				"POST /events/verify",
			},
		})
	})

	// POST /events/hash
	// Receives a laboratory event, calculates SHA-256 hash, registers it in blockchain.
	// Full laboratory data stays in LIMS — only the hash and metadata are stored here.
	r.POST("/events/hash", func(c *gin.Context) {
		var event LimsEvent
		if err := c.ShouldBindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Normalise timestamp — must be stable for deterministic hash
		if event.Timestamp == "" {
			event.Timestamp = time.Now().UTC().Format(time.RFC3339)
		}
		if event.Status == "" {
			event.Status = "REGISTERED"
		}

		hash := calculateHash(event)

		record := HashRecord{
			LimsRecordID: event.LimsRecordID,
			EventType:    event.EventType,
			DataHash:     hash,
			Status:       event.Status,
			RegisteredAt: time.Now().UTC(),
		}

		// Store in local registry as a read cache
		registryMu.Lock()
		registry[event.LimsRecordID] = record
		registryMu.Unlock()

		log.Printf("[HASH] %s | %s | %s", event.LimsRecordID, event.EventType, hash[:16]+"...")

		response := gin.H{
			"success":        true,
			"lims_record_id": event.LimsRecordID,
			"event_type":     event.EventType,
			"data_hash":      hash,
			"timestamp":      event.Timestamp,
			"status":         event.Status,
			"registered_at":  record.RegisteredAt,
		}

		if bc != nil {
			txHash, err := bc.RegisterHash(event.LimsRecordID, event.EventType, hash, event.Status)
			if err != nil {
				log.Printf("[HASH] blockchain registerHash failed for %s: %v", event.LimsRecordID, err)
				response["message"] = "Hash cached locally, blockchain registration failed"
				response["blockchain_error"] = err.Error()
			} else {
				response["message"] = "Hash registered in blockchain"
				response["tx_hash"] = txHash
			}
		} else {
			response["message"] = "Hash registered in local registry (blockchain not connected)"
		}

		c.JSON(http.StatusOK, response)
	})

	// GET /events — returns all registered hash records
	r.GET("/events", func(c *gin.Context) {
		registryMu.RLock()
		records := make([]HashRecord, 0, len(registry))
		for _, rec := range registry {
			records = append(records, rec)
		}
		registryMu.RUnlock()

		// Sort by registered_at descending for consistent display
		sort.Slice(records, func(i, j int) bool {
			return records[i].RegisteredAt.After(records[j].RegisteredAt)
		})

		c.JSON(http.StatusOK, gin.H{
			"count":   len(records),
			"records": records,
		})
	})

	// POST /events/verify
	// Recalculates hash from submitted event data and compares with stored blockchain hash.
	// This is the core integrity verification — detects any data modification in LIMS.
	r.POST("/events/verify", func(c *gin.Context) {
		var event LimsEvent
		if err := c.ShouldBindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		computedHash := calculateHash(event)

		if bc != nil {
			verified, err := bc.VerifyHash(event.LimsRecordID, computedHash)
			if err != nil {
				log.Printf("[VERIFY ERROR] %s: %v", event.LimsRecordID, err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"verified": false,
					"message":  "Blockchain verification failed: " + err.Error(),
					"hash":     computedHash,
				})
				return
			}

			message := "Data integrity verified on blockchain"
			if !verified {
				message = "Hash mismatch on blockchain. Data may have been changed"
			}
			log.Printf("[VERIFY %v] %s (source: blockchain)", verified, event.LimsRecordID)
			c.JSON(http.StatusOK, gin.H{
				"verified": verified,
				"message":  message,
				"hash":     computedHash,
				"source":   "blockchain",
			})
			return
		}

		registryMu.RLock()
		stored, exists := registry[event.LimsRecordID]
		registryMu.RUnlock()

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"verified": false,
				"message":  "No hash registered for this LIMS record ID",
				"hash":     computedHash,
				"source":   "local_cache",
			})
			return
		}

		if stored.DataHash == computedHash {
			log.Printf("[VERIFY OK] %s (source: local_cache)", event.LimsRecordID)
			c.JSON(http.StatusOK, gin.H{
				"verified": true,
				"message":  "Data integrity verified",
				"hash":     computedHash,
				"source":   "local_cache",
			})
		} else {
			log.Printf("[VERIFY FAIL] %s — hash mismatch (source: local_cache)", event.LimsRecordID)
			c.JSON(http.StatusOK, gin.H{
				"verified":      false,
				"message":       "Hash mismatch. Data may have been changed",
				"computed_hash": computedHash,
				"stored_hash":   stored.DataHash,
				"source":        "local_cache",
			})
		}
	})

	log.Printf("LIMS Adapter listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
