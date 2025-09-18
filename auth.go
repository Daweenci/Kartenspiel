package main

import (
	"encoding/json"
	"net/http"
)

type AuthResponse struct {
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{Message: "Invalid request"})
		return
	}

	// Validate input
	if req.Name == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{Message: "Username and password required"})
		return
	}

	// Check credentials (replace with your DB lookup)
	if !checkCredentials(req.Name, req.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthResponse{Message: "Invalid username or password"})
		return
	}

	// Generate JWT
	playerID := "examplePlayerID" // replace with real player ID from DB
	token, err := generateJWT(playerID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AuthResponse{Message: "Failed to generate token"})
		return
	}

	// Send JSON response with token
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{Token: token, Message: "Login successful"})
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON request body
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{Message: "Invalid request"})
		return
	}

	// Validate input
	if req.Name == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{Message: "Username and password required"})
		return
	}

	// Check if username already exists
	if !registerPlayer(req.Name, req.Password) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{Message: "Username already exists"})
		return
	}

	// Generate JWT for new user
	playerID := "exampleNewPlayerID" // replace with real player ID from DB
	token, err := generateJWT(playerID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AuthResponse{Message: "Failed to generate token"})
		return
	}

	// Send JSON response with token
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{Token: token, Message: "Registration successful"})
}
