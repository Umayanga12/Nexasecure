package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type requestToken struct {
	tokenid string `json:"uuid"`
	Token   string `json:"token"`
}

type tokenDatabase struct {
	mu     sync.RWMutex
	tokens map[string]requestToken
}

var db = tokenDatabase{
	tokens: make(map[string]requestToken),
}

func main() {
	http.HandleFunc("/setToken", setTokenHandler)
	http.HandleFunc("/getToken", getTokenHandler)

	port := ":3030"
	fmt.Printf("Starting server on port %s...\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func setTokenHandler(w http.ResponseWriter, r *http.Request) {
	var token requestToken
	err := json.NewDecoder(r.Body).Decode(&token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	db.tokens[token.tokenid] = token
	fmt.Fprintf(w, "Token set successfully for UUID: %s\n", token.tokenid)
}

func getTokenHandler(w http.ResponseWriter, r *http.Request) {
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		http.Error(w, "UUID parameter is required", http.StatusBadRequest)
		return
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	token, ok := db.tokens[uuid]
	if !ok {
		http.Error(w, fmt.Sprintf("Token not found for UUID: %s", uuid), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(token)
}
