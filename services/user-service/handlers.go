package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

func usersHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	// Remove "users/" prefix if present (from API Gateway routing)
	path = strings.TrimPrefix(path, "users/")
	w.Header().Set("Content-Type", "application/json")
	
	switch r.Method {
	case http.MethodPost:
		if path == "" {
			createUserHandler(w, r)
		} else {
			http.Error(w, "Invalid path for POST", http.StatusBadRequest)
		}
	case http.MethodGet:
		if path == "" {
			getAllUsersHandler(w, r)
		} else {
			getUserHandler(w, r, path)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user.ID = uuid.New().String()

	if err := saveUser(user); err != nil {
		http.Error(w, "Failed to save user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := getAllUsers()
	if err != nil {
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(users)
}

func getUserHandler(w http.ResponseWriter, r *http.Request, userID string) {
	user, err := findUserByID(userID)
	if err == redis.Nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}
