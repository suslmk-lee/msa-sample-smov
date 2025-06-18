package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func bookingsHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
		if path == "" {
			createBookingHandler(w, r)
		} else {
			http.Error(w, "Invalid path for POST", http.StatusBadRequest)
		}
	case http.MethodGet:
		if len(parts) == 2 && parts[0] == "user" {
			getUserBookingsHandler(w, r, parts[1])
		} else {
			http.Error(w, "Invalid path for GET", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func createBookingHandler(w http.ResponseWriter, r *http.Request) {
	var booking Booking
	if err := json.NewDecoder(r.Body).Decode(&booking); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	booking.ID = uuid.New().String()

	if err := saveBooking(booking); err != nil {
		http.Error(w, "Failed to save booking", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(booking)
}

func getUserBookingsHandler(w http.ResponseWriter, r *http.Request, userID string) {
	bookings, err := findUserBookings(userID)
	if err != nil {
		http.Error(w, "Failed to retrieve bookings", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bookings)
}
