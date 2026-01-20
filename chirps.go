package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"

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
		log.Printf("handleAddChirp: failed to get bearer token: %v", err)
		http.Error(w, "Unauthorized, no user token provided.", http.StatusUnauthorized)
		return
	}

	userId, err := auth.ValidateJWT(token, a.jwtAuthSecret)
	if err != nil || userId == uuid.Nil {
		log.Printf("handleAddChirp: failed to validate JWT: %v", err)
		http.Error(w, "Unauthorized. Invalid user token.", http.StatusUnauthorized)
		return
	}

	var payload database.CreateChirpParams
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("handleAddChirp: failed to decode request body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	cleanedBody, err := ValidateChirp(payload.Body)
	if err != nil {
		log.Printf("handleAddChirp: chirp validation failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	payload.Body = cleanedBody
	payload.UserID = userId

	chirp, err := a.dbQueries.CreateChirp(r.Context(), payload)
	if err != nil {
		log.Printf("handleAddChirp: failed to create chirp: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chirp)
}

func (a *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var chirps []database.Chirp
	var err error
	if userIdText := r.URL.Query().Get("author_id"); userIdText != "" {
		userId, err := uuid.Parse(userIdText)
		if err != nil {
			log.Printf("handlerGetChirps: invalid author_id parameter: %v", err)
			http.Error(w, "Invalid author_id parameter", http.StatusBadRequest)
			return
		}
		chirps, err = a.dbQueries.GetChirpsByUserId(r.Context(), userId)
		if err != nil {
			log.Printf("handlerGetChirps: failed to get chirps by user ID: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		chirps, err = a.dbQueries.GetChirpsByCreation(r.Context())
		if err != nil {
			log.Printf("handlerGetChirps: failed to get chirps: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	sort := strings.ToLower(r.URL.Query().Get("sort"))
	if sort == "desc" {
		slices.Reverse(chirps)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chirps)
}

func (a *apiConfig) handlerGetChirpById(w http.ResponseWriter, r *http.Request) {
	idText := r.PathValue("chirpId")
	if idText == "" {
		log.Printf("handlerGetChirpById: missing ID parameter")
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(idText)
	if err != nil {
		log.Printf("handlerGetChirpById: invalid ID parameter: %v", err)
		http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
		return
	}

	chirp, err := a.dbQueries.GetChirpById(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("handlerGetChirpById: chirp not found: %s", id.String())
			http.Error(w, fmt.Sprintf("Chirp with ID %s not found", id.String()), http.StatusNotFound)
			return
		}
		log.Printf("handlerGetChirpById: failed to get chirp: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chirp)
}

func (a *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("handleAddChirp: failed to get bearer token: %v", err)
		http.Error(w, "Unauthorized, no user token provided.", http.StatusUnauthorized)
		return
	}

	userId, err := auth.ValidateJWT(token, a.jwtAuthSecret)
	if err != nil || userId == uuid.Nil {
		log.Printf("handleAddChirp: failed to validate JWT: %v", err)
		http.Error(w, "Unauthorized. Invalid user token.", http.StatusUnauthorized)
		return
	}
	
	idText := r.PathValue("chirpId")
	if idText == "" {
		log.Printf("handlerDeleteChirp: missing ID parameter")
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(idText)
	if err != nil {
		log.Printf("handlerDeleteChirp: invalid ID parameter: %v", err)
		http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
		return
	}

	chirp, err := a.dbQueries.GetChirpById(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("handlerGetChirpById: chirp not found: %s", id.String())
			http.Error(w, fmt.Sprintf("Chirp with ID %s not found", id.String()), http.StatusNotFound)
			return
		}
		log.Printf("handlerGetChirpById: failed to get chirp: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if chirp.UserID != userId {
		log.Printf("handlerDeleteChirp: user %s unauthorized to delete chirp %s", userId.String(), id.String())
		http.Error(w, "Forbidden: you can only delete your own chirps", http.StatusForbidden)
		return
	}

	err = a.dbQueries.DeleteChirp(r.Context(), id)
	if err != nil {
		log.Printf("handlerDeleteChirp: failed to delete chirp: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}