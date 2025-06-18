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
	userServiceProxy := newReverseProxy("http://user-service:8081")
	http.Handle("/users/", http.StripPrefix("/users/", userServiceProxy))

	// Movie Service
	movieServiceProxy := newReverseProxy("http://movie-service:8082")
	http.Handle("/movies/", http.StripPrefix("/movies/", movieServiceProxy))

	// Booking Service
	bookingServiceProxy := newReverseProxy("http://booking-service:8083")
	http.Handle("/bookings/", http.StripPrefix("/bookings/", bookingServiceProxy))
	if err != nil {
		log.Fatalf("Failed to parse booking service URL: %v", err)
	}
	bookingProxy := httputil.NewSingleHostReverseProxy(bookingURL)
	http.Handle("/bookings/", http.StripPrefix("/bookings/", bookingProxy))

	log.Println("API Gateway started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
