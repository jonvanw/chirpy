package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jonvanw/chirpy/internal/auth"
	"github.com/jonvanw/chirpy/internal/database"
)

type userInfoPost struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	ExpiresInSeconds int `json:"expires_in_seconds"`
}

type userInfoResponse struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Email          string    `json:"email"`
	Token          string    `json:"token,omitempty"`
}

func (a *apiConfig) handleAddUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var payload userInfoPost
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	args, err := payload.ToDbArgs()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	userRaw, err := a.dbQueries.CreateUser(r.Context(), args)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user := userInfoResponse{
		ID:        userRaw.ID,
		CreatedAt: userRaw.CreatedAt,
		UpdatedAt: userRaw.UpdatedAt,
		Email:     userRaw.Email,
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)	
}

func (a *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload userInfoPost
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	userRaw, err := a.dbQueries.GetUserByEmail(r.Context(), payload.Email)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	authorized, err := auth.CheckPasswordHash(payload.Password, userRaw.HashedPassword)
	if err != nil  {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	if !authorized {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	const hourInSeconds = 60 * 60 // min/hr * sec/min
	if payload.ExpiresInSeconds <= 0 || payload.ExpiresInSeconds > hourInSeconds {
		payload.ExpiresInSeconds = hourInSeconds
	}
	token, err := auth.MakeJWT(
		userRaw.ID,
		a.jwtAuthSecret,
		time.Duration(payload.ExpiresInSeconds) * time.Second,
	)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user := userInfoResponse{
		ID:        userRaw.ID,
		CreatedAt: userRaw.CreatedAt,
		UpdatedAt: userRaw.UpdatedAt,
		Email:     userRaw.Email,
		Token:     token,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)	
}

func (u *userInfoPost) ToDbArgs() (database.CreateUserParams, error) {
	hashed, err := auth.HashPassword(u.Password)
	if err != nil {
		return database.CreateUserParams{}, err
	}
	return database.CreateUserParams{
		Email: u.Email,
		HashedPassword: hashed,
	}, nil
}