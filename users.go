package main

import (
	"encoding/json"
	"net/http"
)

func (a *apiConfig) handleAddUser(w http.ResponseWriter, r *http.Request) {
	type requestPayload struct {
		Email string `json:"email"`
	}

	// type responsePayload struct {
	// 	ID        uuid.UUID `json:"id"`
	// 	CreatedAt time.Time `json:"created_at"`
	// 	UpdatedAt time.Time `json:"updated_at"`
	// 	Email     string    `json:"email"`
	// }

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload requestPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	user, err := a.dbQueries.CreateUser(r.Context(), payload.Email)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// response := responsePayload{
	// 	ID:        user.ID,
	// 	CreatedAt: user.CreatedAt,
	// 	UpdatedAt: user.UpdatedAt,
	// 	Email:     user.Email,
	// }

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)	
}
