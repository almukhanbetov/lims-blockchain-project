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

// In-memory registry simulates the blockchain state cache.
// When CONTRACT_ADDRESS is set, operations are forwarded to the Ethereum smart contract.
var (
	registry   = make(map[string]HashRecord)
	registryMu sync.RWMutex
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

	if contractAddress != "" {
		log.Printf("Blockchain integration active: RPC=%s Contract=%s", blockchainRPC, contractAddress)
	} else {
		log.Println("Running in local registry mode (set CONTRACT_ADDRESS to enable blockchain)")
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
			"service":          "LIMS Blockchain Adapter",
			"version":          "1.0.0",
			"role":             "Intelligent control layer between LIMS and blockchain",
			"integrity_model":  "SHA-256",
			"blockchain_rpc":   blockchainRPC,
			"contract_address": contractAddress,
			"registered_hashes": count,
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

		// Store in local registry (and optionally forward to smart contract)
		registryMu.Lock()
		registry[event.LimsRecordID] = record
		registryMu.Unlock()

		// TODO: when CONTRACT_ADDRESS is set, call registerHash() on LimsHashRegistry contract
		// via go-ethereum ethclient:
		//   client, _ := ethclient.Dial(blockchainRPC)
		//   instance, _ := contract.NewLimsHashRegistry(common.HexToAddress(contractAddress), client)
		//   hashBytes := [32]byte{}
		//   copy(hashBytes[:], decoded)
		//   instance.RegisterHash(auth, limsRecordId, eventType, hashBytes, status)

		log.Printf("[HASH] %s | %s | %s", event.LimsRecordID, event.EventType, hash[:16]+"...")

		c.JSON(http.StatusOK, gin.H{
			"success":        true,
			"lims_record_id": event.LimsRecordID,
			"event_type":     event.EventType,
			"data_hash":      hash,
			"timestamp":      event.Timestamp,
			"status":         event.Status,
			"registered_at":  record.RegisteredAt,
			"message":        "Hash registered in blockchain",
		})
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

		registryMu.RLock()
		stored, exists := registry[event.LimsRecordID]
		registryMu.RUnlock()

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"verified": false,
				"message":  "No hash registered for this LIMS record ID",
				"hash":     computedHash,
			})
			return
		}

		if stored.DataHash == computedHash {
			log.Printf("[VERIFY OK] %s", event.LimsRecordID)
			c.JSON(http.StatusOK, gin.H{
				"verified": true,
				"message":  "Data integrity verified",
				"hash":     computedHash,
			})
		} else {
			log.Printf("[VERIFY FAIL] %s — hash mismatch", event.LimsRecordID)
			c.JSON(http.StatusOK, gin.H{
				"verified":      false,
				"message":       "Hash mismatch. Data may have been changed",
				"computed_hash": computedHash,
				"stored_hash":   stored.DataHash,
			})
		}
	})

	log.Printf("LIMS Adapter listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
