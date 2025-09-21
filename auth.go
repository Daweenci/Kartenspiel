package main

import (
	"encoding/json"
	"net/http"
)

type AuthResponse struct {
	Token   string      `json:"token,omitempty"`
	Message string      `json:"message,omitempty"`
	Type    MessageType `json:"type,omitempty"`
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{Type: ResponseLoginFailed, Message: "Invalid request"})
		return
	}

	if req.Name == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{Type: ResponseLoginFailed, Message: "Username and password required"})
		return
	}

	playerID, err := authenticatePlayer(req.Name, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthResponse{Type: ResponseLoginFailed, Message: "Invalid username or password"})
		return
	}

	token, err := generateJWT(playerID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AuthResponse{Type: ResponseLoginFailed, Message: "Failed to generate token"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{Token: token, Type: ResponseLoginSuccessful, Message: "Login successful"})
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{Type: ResponseRegisterFailed, Message: "Invalid request"})
		return
	}

	if req.Name == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{Type: ResponseRegisterFailed, Message: "Username and password required"})
		return
	}

	playerID, err := registerPlayer(req.Name, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{Type: ResponseRegisterFailed, Message: err.Error()})
		return
	}

	token, err := generateJWT(playerID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AuthResponse{Type: ResponseRegisterFailed, Message: "Failed to generate token"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{Token: token, Type: ResponseRegisterSuccessful, Message: "Registration successful"})
}
