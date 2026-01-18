package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jonvanw/chirpy/internal/auth"
	"github.com/jonvanw/chirpy/internal/database"
)

func (a *apiConfig) handleAddChirp(w http.ResponseWriter, r *http.Request) {	
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Unauthorized, no user token provided.", http.StatusUnauthorized)
		return
	}

	userId, err := auth.ValidateJWT(token, a.jwtAuthSecret)
	if err != nil || userId == uuid.Nil {
		http.Error(w, "Unauthorized. Invalid user token.", http.StatusUnauthorized)
		return
	}

	var payload database.CreateChirpParams
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	cleanedBody, err := ValidateChirp(payload.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	payload.Body = cleanedBody
	payload.UserID = userId

	chirp, err := a.dbQueries.CreateChirp(r.Context(), payload)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chirp)
}

func (a *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := a.dbQueries.GetChirpsByCreation(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chirps)
}

func (a *apiConfig) handlerGetChirpById(w http.ResponseWriter, r *http.Request) {
	idText := r.PathValue("chirpId")
	if idText == "" {
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(idText)
	if err != nil {
		http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
		return
	}
	

	chirp, err := a.dbQueries.GetChirpById(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, fmt.Sprintf("Chirp with ID %s not found", id.String()), http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chirp)
}