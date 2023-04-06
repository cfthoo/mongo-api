package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mongo-api/api"
)

var client *mongo.Client

func main() {
	// Initiate a MongoDB client
	var err error

	// please make sure the mongodb server is started.
	mongoUrl := "mongodb://localhost:27017"

	client, err = mongo.NewClient(options.Client().ApplyURI(mongoUrl))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	mongoApi := api.MongoClient{Client: client}

	router := mux.NewRouter()

	// This is to enable CORS for all routes (including frontend)
	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
	)

	router.HandleFunc("/users", mongoApi.CreateUserHandler).Methods("POST")
	router.HandleFunc("/users/{id}", mongoApi.DeleteUserHandler).Methods("DELETE")
	router.HandleFunc("/users/all", mongoApi.ListUsersHandler).Methods("GET")
	router.HandleFunc("/users/{id}", mongoApi.GetUserByIdHandler).Methods("GET")
	router.HandleFunc("/image", mongoApi.UploadImage).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", cors(router)))
}
