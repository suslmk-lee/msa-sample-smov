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

// Movie represents a movie model
type Movie struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Director string `json:"director"`
	Genre    string `json:"genre"`
}

func init() {
	// Initialize Redis client
	rdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379", // Redis address
		Password: "",                 // No password set
		DB:       0,                  // Default DB
	})
}

func moviesHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		if path == "" {
			getAllMovies(w, r)
		} else {
			getMovie(w, r, path)
		}
	case http.MethodPost:
		if path == "" {
			createMovie(w, r)
		} else {
			http.Error(w, "Method not allowed on specific resource", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func createMovie(w http.ResponseWriter, r *http.Request) {
	var movie Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	movie.ID = uuid.New().String()

	movieJSON, err := json.Marshal(movie)
	if err != nil {
		http.Error(w, "Failed to marshal movie", http.StatusInternalServerError)
		return
	}

	if err := rdb.Set(ctx, "movie:"+movie.ID, movieJSON, 0).Err(); err != nil {
		log.Printf("Failed to save movie to Redis: %v", err)
		http.Error(w, "Failed to save movie", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(movie)
}

func getMovie(w http.ResponseWriter, r *http.Request, id string) {
	movieJSON, err := rdb.Get(ctx, "movie:"+id).Result()
	if err == redis.Nil {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Failed to get movie from Redis: %v", err)
		http.Error(w, "Failed to get movie", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(movieJSON))
}

func getAllMovies(w http.ResponseWriter, r *http.Request) {
	keys, err := rdb.Keys(ctx, "movie:*").Result()
	if err != nil {
		log.Printf("Failed to get movie keys from Redis: %v", err)
		http.Error(w, "Failed to retrieve movies", http.StatusInternalServerError)
		return
	}

	if len(keys) == 0 {
		json.NewEncoder(w).Encode([]Movie{})
		return
	}

	moviesData, err := rdb.MGet(ctx, keys...).Result()
	if err != nil {
		log.Printf("Failed to get movies from Redis: %v", err)
		http.Error(w, "Failed to retrieve movies", http.StatusInternalServerError)
		return
	}

	movies := make([]Movie, 0, len(moviesData))
	for _, movieJSON := range moviesData {
		if movieJSON == nil {
			continue
		}
		var movie Movie
		if err := json.Unmarshal([]byte(movieJSON.(string)), &movie); err != nil {
			log.Printf("Failed to unmarshal movie data: %v", err)
			continue
		}
		movies = append(movies, movie)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(movies)
}

func main() {
	http.HandleFunc("/", moviesHandler)

	log.Println("Movie Service started on :8082")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
