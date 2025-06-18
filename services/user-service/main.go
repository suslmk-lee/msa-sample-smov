package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

var (
	rdb *redis.Client
	ctx = context.Background()
)

// User represents a user model
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func init() {
	// Initialize Redis client
	rdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379", // Default Redis address
		Password: "",               // No password set
		DB:       0,                // Default DB
	})
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createUser(w, r)
	case http.MethodGet:
		getUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user.ID = uuid.New().String()

	userJSON, err := json.Marshal(user)
	if err != nil {
		http.Error(w, "Failed to marshal user", http.StatusInternalServerError)
		return
	}

	if err := rdb.Set(ctx, "user:"+user.ID, userJSON, 0).Err(); err != nil {
		log.Printf("Failed to save user to Redis: %v", err)
		http.Error(w, "Failed to save user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/")
	userJSON, err := rdb.Get(ctx, "user:"+id).Result()
	if err == redis.Nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	var user User
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		http.Error(w, "Failed to unmarshal user", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func main() {
	http.HandleFunc("/", usersHandler)

	log.Println("User Service started on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
