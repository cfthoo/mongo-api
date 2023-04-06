package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
)

type MongoClient struct {
	Client *mongo.Client
}

type User struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username string             `json:"username,omitempty" bson:"username,omitempty"`
	Email    string             `json:"email,omitempty" bson:"email,omitempty"`
}

// Define a data model for the image
type Image struct {
	Name     string    `json:"name"`
	Data     []byte    `json:"data"`
	MimeType string    `json:"mime_type"`
	Created  time.Time `json:"created"`
}

func (mc *MongoClient) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert user into MongoDB
	result, err := mc.Client.Database("mydb").Collection("users").InsertOne(context.Background(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id := result.InsertedID.(primitive.ObjectID).Hex()
	w.Write([]byte(id))
}

func (mc *MongoClient) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL
	id, err := primitive.ObjectIDFromHex(r.URL.Path[len("/users/"):])
	if err != nil {
		http.Error(w, "id parameter is required", http.StatusBadRequest)
		return
	}
	fmt.Println(id)
	// id, err := strconv.Atoi(idStr)
	// if err != nil {
	// 	http.Error(w, "invalid id parameter", http.StatusBadRequest)
	// 	return
	// }

	// delete the user by id
	result, err := mc.Client.Database("mydb").Collection("users").DeleteOne(context.Background(), bson.M{"id": id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "User deleted successfully")
}

func (mc *MongoClient) GetUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL
	id, err := primitive.ObjectIDFromHex(r.URL.Path[len("/users/"):])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Query user by id
	var user User
	err = mc.Client.Database("mydb").Collection("users").FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return user in JSON format
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (mc *MongoClient) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	// List all users in MongoDB
	cursor, err := mc.Client.Database("mydb").Collection("users").Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	// Write users to response as JSON array
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("["))
	for cursor.Next(context.Background()) {
		var user User
		err = cursor.Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonData, err := json.Marshal(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(jsonData)
		if cursor.RemainingBatchLength() > 0 {
			w.Write([]byte(","))
		}
	}
	w.Write([]byte("]"))

}
func (mc *MongoClient) UploadImage(w http.ResponseWriter, r *http.Request) {
	// in order to use this function , the frontend/client has to encode the image
	// to binary format.

	// Read image data from the request body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Decode image data from base64
	decodedData, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		http.Error(w, "Failed to decode image data", http.StatusBadRequest)
		return
	}

	// Create new instance for Image data model with the decoded image binary
	image := Image{
		Name:     r.FormValue("name"),
		Data:     decodedData,
		MimeType: r.FormValue("mime_type"),
		Created:  time.Now(),
	}

	// Save image data to MongoDB by using GridFS
	gridFS, _ := gridfs.NewBucket(mc.Client.Database("mydb"))
	// if err != nil {
	// 	http.Error(w, "Failed to save image data", http.StatusInternalServerError)
	// 	return
	// }
	_, err = gridFS.UploadFromStream(image.Name, ioutil.NopCloser(bytes.NewReader(image.Data)))
	if err != nil {
		http.Error(w, "Failed to save image data", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Image uploaded successfully")
}
