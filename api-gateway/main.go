package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	// UI Files
	fs := http.FileServer(http.Dir("../ui"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "../ui/index.html")
			return
		}
		http.NotFound(w, r)
	})

	// User Service
	userURL, err := url.Parse("http://127.0.0.1:8081")
	if err != nil {
		log.Fatalf("Failed to parse user service URL: %v", err)
	}
	userProxy := httputil.NewSingleHostReverseProxy(userURL)
	http.Handle("/users/", http.StripPrefix("/users/", userProxy))

	// Movie Service
	movieURL, err := url.Parse("http://127.0.0.1:8082")
	if err != nil {
		log.Fatalf("Failed to parse movie service URL: %v", err)
	}
	movieProxy := httputil.NewSingleHostReverseProxy(movieURL)
	http.Handle("/movies/", http.StripPrefix("/movies/", movieProxy))

	log.Println("API Gateway started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
