package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url" // newReverseProxy에서 사용됨
)

// newReverseProxy creates a new reverse proxy for the target URL.
func newReverseProxy(target string) *httputil.ReverseProxy {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Failed to parse target URL %s: %v", target, err)
	}
	return httputil.NewSingleHostReverseProxy(targetURL)
}

func main() {
	// UI Files
	http.Handle("/", http.FileServer(http.Dir("ui")))

	// User Service
	userServiceProxy := newReverseProxy("http://user-service:8081")
	http.Handle("/users/", http.StripPrefix("/users/", userServiceProxy))

	// Movie Service
	movieServiceProxy := newReverseProxy("http://movie-service:8082")
	http.Handle("/movies/", http.StripPrefix("/movies/", movieServiceProxy))

	// Booking Service
	bookingServiceProxy := newReverseProxy("http://booking-service:8083")
	http.Handle("/bookings/", http.StripPrefix("/bookings/", bookingServiceProxy))

	log.Println("API Gateway started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
