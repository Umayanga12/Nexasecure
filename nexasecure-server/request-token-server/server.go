package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func connectToMongoDB() {
	var err error
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27019")
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to DB!")
}

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

	json.NewEncoder(w).Encode(results)
}

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

	json.NewEncoder(w).Encode(updateResult)
}

func main() {
	// Connect to MongoDB
	connectToMongoDB()

	// Set up handlers for read and update
	http.HandleFunc("/read", readHandler)
	http.HandleFunc("/update", updateHandler)

	// Start the server
	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
