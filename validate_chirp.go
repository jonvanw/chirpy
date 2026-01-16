package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
)


func HandleValidateChirp(w http.ResponseWriter, r *http.Request) {
	type chirpRequet struct {
		Body string `json:"body"`
	}

	type chirpResponse struct {
		CleanedBody string `json:"cleaned_body"`
	}

	type chirpError struct {
		Error string `json:"error"`
	}

	decoder := json.NewDecoder(r.Body)
	var chirpReq chirpRequet
	err := decoder.Decode(&chirpReq)
	if err != nil {
		log.Printf("Failed to decode request body: %v", err)
		log.Printf("Request Body: %v", r.Body)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(chirpError{Error: "Invalid request payload"})
		return
	}

	if len(chirpReq.Body) > 140  {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(chirpError{Error: "Chirp is too long"})
		return
	}

	response := chirpResponse{
		CleanedBody: CleanChirp(chirpReq.Body),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func CleanChirp(text string) string {
	blackList := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(text, " ")
	for i, word := range words {
		if slices.Contains(blackList, strings.ToLower(word)) {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}

