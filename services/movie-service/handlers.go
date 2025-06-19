package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

func moviesHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		if path == "" {
			getAllMoviesHandler(w, r)
		} else {
			getMovieHandler(w, r, path)
		}
	case http.MethodPost:
		if path == "" {
			createMovieHandler(w, r)
		} else {
			http.Error(w, "Method not allowed on specific resource", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var movie Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	movie.ID = uuid.New().String()

	if err := saveMovie(movie); err != nil {
		http.Error(w, "Failed to save movie", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(movie)
}

func getMovieHandler(w http.ResponseWriter, r *http.Request, id string) {
	movie, err := findMovieByID(id)
	if err == redis.Nil {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Failed to get movie from Redis: %v", err)
		http.Error(w, "Failed to get movie", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(movie)
}

func getAllMoviesHandler(w http.ResponseWriter, r *http.Request) {
	movies, err := findAllMovies()
	if err != nil {
		http.Error(w, "Failed to retrieve movies", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(movies)
}
