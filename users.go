package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jonvanw/chirpy/internal/auth"
	"github.com/jonvanw/chirpy/internal/database"
)

const jwtDuration = time.Hour

type userInfoRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userInfoResponse struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Email          string    `json:"email"`
	Token          string    `json:"token,omitempty"`
	RefreshToken   string    `json:"refresh_token,omitempty"`
	IsChirpyRed    bool      `json:"is_chirpy_red"`
}

func (a *apiConfig) handleAddUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var payload userInfoRequest
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("handleAddUser: failed to decode request body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	args, err := payload.ToInsertDbArgs()
	if err != nil {
		log.Printf("handleAddUser: failed to convert to db args: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	userRaw, err := a.dbQueries.CreateUser(r.Context(), args)
	if err != nil {
		log.Printf("handleAddUser: failed to create user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user := userInfoResponse{
		ID:        userRaw.ID,
		CreatedAt: userRaw.CreatedAt,
		UpdatedAt: userRaw.UpdatedAt,
		Email:     userRaw.Email,
		IsChirpyRed: userRaw.IsChirpyRed,
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

	var payload userInfoRequest
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("handleLogin: failed to decode request body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	userRaw, err := a.dbQueries.GetUserByEmail(r.Context(), payload.Email)
	if err != nil {
		log.Printf("handleLogin: failed to get user by email: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	authorized, err := auth.CheckPasswordHash(payload.Password, userRaw.HashedPassword)
	if err != nil  {
		log.Printf("handleLogin: failed to check password hash: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	if !authorized {
		log.Printf("handleLogin: unauthorized login attempt for email: %s", payload.Email)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	jwt, err := auth.MakeJWT(
		userRaw.ID,
		a.jwtAuthSecret,
		jwtDuration,
	)
	if err != nil {
		log.Printf("handleLogin: failed to create JWT: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("handleLogin: failed to create refresh token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_,  err = a.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token: refreshToken,
		UserID: userRaw.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour), // 60 days
	})
	if err != nil {
		log.Printf("handleLogin: failed to save refresh token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user := userInfoResponse{
		ID:        userRaw.ID,
		CreatedAt: userRaw.CreatedAt,
		UpdatedAt: userRaw.UpdatedAt,
		Email:     userRaw.Email,
		Token:     jwt,
		RefreshToken: refreshToken,
		IsChirpyRed: userRaw.IsChirpyRed,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)	
}

func (a *apiConfig) handleRefreshAuthToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("handleRefreshAuthToken: failed to get bearer token: %v", err)
		http.Error(w, "Unauthorized, no user token provided.", http.StatusUnauthorized)
		return
	}
	refreshTokenRecord, err := a.dbQueries.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		log.Printf("handleRefreshAuthToken: failed to get refresh token from db: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Unauthorized, invalid refresh token.", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	if refreshTokenRecord.RevokedAt.Valid {
		log.Printf("handleRefreshAuthToken: refresh token revoked for user: %s", refreshTokenRecord.UserID)
		http.Error(w, "Unauthorized, refresh token revoked.", http.StatusUnauthorized)
		return
	}
	if refreshTokenRecord.ExpiresAt.Before(time.Now()) {
		log.Printf("handleRefreshAuthToken: refresh token expired for user: %s", refreshTokenRecord.UserID)
		http.Error(w, "Unauthorized, refresh token expired.", http.StatusUnauthorized)
		return
	}

	jwt, err := auth.MakeJWT(
		refreshTokenRecord.UserID,
		a.jwtAuthSecret,
		jwtDuration,
	)
	if err != nil {
		log.Printf("handleRefreshAuthToken: failed to create JWT: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	res := struct {
		Token string `json:"token"`
	}{
		Token: jwt,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (a *apiConfig) handleRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("handleRevokeRefreshToken: failed to get bearer token: %v", err)
		http.Error(w, "Unauthorized, no refresh token provided.", http.StatusUnauthorized)
		return
	}

	refreshTokenRecord, err := a.dbQueries.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		log.Printf("handleRevokeRefreshToken: failed to get refresh token from db: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Unauthorized, invalid refresh token.", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	if refreshTokenRecord.RevokedAt.Valid {
		w.WriteHeader(http.StatusNoContent)  // already revoked so treat as no op and report success to user
		return
	}

	if err = a.dbQueries.RevokeRefreshToken(r.Context(), refreshToken); err != nil {
		log.Printf("handleRevokeRefreshToken: failed to revoke refresh token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
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

	var payload userInfoRequest
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("handleUpdateUser: failed to decode request body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	args, err := payload.ToUpdateDbArgs(userId)
	if err != nil {
		log.Printf("handleUpdateUser: failed to convert to db args: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	userRaw, err := a.dbQueries.UpdateUser(r.Context(), args)
	if err != nil {
		log.Printf("handleUpdateUser: failed to update user: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Unauthorized: unknown user", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	user := userInfoResponse{
		ID:        userRaw.ID,
		CreatedAt: userRaw.CreatedAt,
		UpdatedAt: userRaw.UpdatedAt,
		Email:     userRaw.Email,
		IsChirpyRed: userRaw.IsChirpyRed,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (u *userInfoRequest) ToInsertDbArgs() (database.CreateUserParams, error) {
	hashed, err := auth.HashPassword(u.Password)
	if err != nil {
		return database.CreateUserParams{}, err
	}
	return database.CreateUserParams{
		Email: u.Email,
		HashedPassword: hashed,
	}, nil
}

func (u *userInfoRequest) ToUpdateDbArgs(userId uuid.UUID) (database.UpdateUserParams, error) {
	hashed, err := auth.HashPassword(u.Password)
	if err != nil {
		return database.UpdateUserParams{}, err
	}
	return database.UpdateUserParams{
		ID: userId,
		Email: u.Email,
		HashedPassword: hashed,
	}, nil
}