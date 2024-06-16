package main

import (
	"crypto/sha256"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

type DOT struct {
	JWT string
	DOT string
}

type Request struct {
	sender DOT
	data   string
}
type Block struct {
	hash         string
	previousHash string
	POOA         int
	Request      Request
	Timestamp    time.Time
}

type Blockchain struct {
	GenisisBlock Block
	chain        []Block
	difficulty   int
}

func (blockInstence Block) calculateHash() string {
	data, err := json.Marshal(blockInstence.Request)
	if err != nil {
		panic(err)
	}
	blockData := blockInstence.previousHash + string(data) + blockInstence.Timestamp.String() + strconv.Itoa(blockInstence.POOA)
	blockHash := sha256.Sum256([]byte(blockData))
	return string(blockHash[:])
}

func (blockInstence *Block) mine(difficulty int) {
	for !strings.HasPrefix(blockInstence.hash, strings.Repeat("0", difficulty)) {
		blockInstence.POOA++
		blockInstence.hash = blockInstence.calculateHash()
	}
}

func createBlockchain(difficulty int) Blockchain {
	genisBlock := Block{
		hash:      "0000",
		Timestamp: time.Now(),
	}

	return Blockchain{
		GenisisBlock: genisBlock,
		chain:        []Block{genisBlock},
		difficulty:   difficulty,
	}
}

func (blockchainInstence *Blockchain) addBlock(from, to string, request Request) {
	newBlock := Block{
		hash:         "",
		previousHash: blockchainInstence.chain[len(blockchainInstence.chain)-1].hash,
		Request:      request,
		Timestamp:    time.Now(),
	}
	newBlock.mine(blockchainInstence.difficulty)
	blockchainInstence.chain = append(blockchainInstence.chain, newBlock)

}

func (chain Blockchain) isValid() bool {
	for i := range chain.chain[1:] {
		previousBlock := chain.chain[i]
		currentBlock := chain.chain[i+1]
		if currentBlock.hash != currentBlock.calculateHash() || currentBlock.previousHash != previousBlock.hash {
			return false
		}
	}
	return true
}

func main() {

}
