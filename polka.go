package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/jonvanw/chirpy/internal/auth"
	"github.com/jonvanw/chirpy/internal/database"
)

func (a *apiConfig) handlePolkaEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		log.Printf("handlePolkaEvent: failed to get API key: %v", err)
		http.Error(w, "Unauthorized, API key missing.", http.StatusUnauthorized)
		return
	}
	if apiKey != a.pokaApiKey {
		log.Printf("handlePolkaEvent: unauthorized API key: %s", apiKey)
		http.Error(w, "Unauthorized, invalid API key.", http.StatusUnauthorized)
		return
	}

	var payload struct {
		Event string `json:"event"`
		Data struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("handlePolkaEvent: failed to decode request body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if payload.Event != "user.upgraded" {
		log.Printf("handlePolkaEvent: unhandled event: %s", payload.Event)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	
	_, err = a.dbQueries.UpdateUserIsChirpyRed(r.Context(), database.UpdateUserIsChirpyRedParams{
		ID: payload.Data.UserID,
		IsChirpyRed: true,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("handlePolkaEvent: user not found: %v\n", payload.Data.UserID)
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Printf("handlePolkaEvent: failed to update user is chirpy red: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}