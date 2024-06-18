package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

// Function to connect to MongoDB
func connectToMongoDB() {
	var err error
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
}

// Handler to read documents from MongoDB
func readHandler(w http.ResponseWriter, r *http.Request) {
	collection := client.Database("testdb").Collection("testcollection")

	var results []bson.M
	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, result)
	}

	if err := cur.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// Handler to update a document in MongoDB
func updateHandler(w http.ResponseWriter, r *http.Request) {
	collection := client.Database("testdb").Collection("testcollection")

	type RequestBody struct {
		Filter bson.M `json:"filter"`
		Update bson.M `json:"update"`
	}

	var reqBody RequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updateResult, err := collection.UpdateOne(context.Background(), reqBody.Filter, bson.M{"$set": reqBody.Update})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updateResult)
}

// Handler to make a request to an external API
func externalAPIHandler(w http.ResponseWriter, r *http.Request) {
	externalAPIURL := "https://api.example.com/data" // Replace with the actual external API URL

	resp, err := http.Get(externalAPIURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Error from external API: %s", resp.Status), resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func main() {
	// Connect to MongoDB
	connectToMongoDB()

	// Set up handlers for read, update, and external API call
	http.HandleFunc("/read", readHandler)
	http.HandleFunc("/pass", updateHandler)
	http.HandleFunc("/external", externalAPIHandler)

	// Start the server
	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8085", nil))
}
