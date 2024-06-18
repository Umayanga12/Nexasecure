package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type block struct {
	Data      map[string]interface{} `json:"data"`
	Hash      string                 `json:"hash"`
	PrevHash  string                 `json:"prev_hash"`
	Timestamp time.Time              `json:"timestamp"`
	Poa       int                    `json:"poa"`
}

type blockchain struct {
	GenesisBlock block   `json:"genesis_block"`
	Chain        []block `json:"chain"`
	Difficulty   int     `json:"difficulty"`
}

func (b block) calHash() string {
	data, _ := json.Marshal(b.Data)
	blockData := b.PrevHash + string(data) + b.Timestamp.String() + strconv.Itoa(b.Poa)
	blockHash := sha256.Sum256([]byte(blockData))
	return fmt.Sprintf("%x", blockHash)
}

func (b *block) mine(difficulty int) {
	for !strings.HasPrefix(b.Hash, strings.Repeat("0", difficulty)) {
		b.Poa++
		b.Hash = b.calHash()
	}
}

func createBlockchain(difficulty int) blockchain {
	genesisBlock := block{
		Hash:      "0",
		Timestamp: time.Now(),
	}
	return blockchain{
		GenesisBlock: genesisBlock,
		Chain:        []block{genesisBlock},
		Difficulty:   difficulty,
	}
}

func (b *blockchain) addBlock(token string, storetime string, id uuid.UUID) {
	blockData := map[string]interface{}{
		"token": token,
		"uuid":  id.String(),
		"time":  storetime,
	}
	lastBlock := b.Chain[len(b.Chain)-1]
	newBlock := block{
		Data:      blockData,
		PrevHash:  lastBlock.Hash,
		Timestamp: time.Now(),
	}
	newBlock.mine(b.Difficulty)
	b.Chain = append(b.Chain, newBlock)
}

func (b blockchain) isValid() bool {
	for i := 1; i < len(b.Chain); i++ {
		previousHash := b.Chain[i-1].Hash
		currentBlock := b.Chain[i]
		if currentBlock.PrevHash != previousHash || currentBlock.calHash() != currentBlock.Hash {
			return false
		}
	}
	return true
}

var Blockchain blockchain

func handleAddBlock(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token     string `json:"token"`
		Storetime string `json:"storetime"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if data.Token == "" || data.Storetime == "" {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	id := uuid.New()
	Blockchain.addBlock(data.Token, data.Storetime, id)

	if Blockchain.isValid() {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]bool{"is_valid": true})
	} else {
		// Remove the invalid block
		Blockchain.Chain = Blockchain.Chain[:len(Blockchain.Chain)-1]
		http.Error(w, "Blockchain is invalid after adding the block", http.StatusInternalServerError)
	}
}

func startValidationTicker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if !Blockchain.isValid() {
				fmt.Println("Blockchain is invalid!")
				// Optionally, take other actions, like alerting or halting operations
			} else {
				fmt.Println("Blockchain is valid.")
			}
		}
	}()
}

func main() {
	difficulty := 2
	Blockchain = createBlockchain(difficulty)

	http.HandleFunc("/addBlock", handleAddBlock)

	startValidationTicker(10 * time.Second)
	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8088", nil)
}
