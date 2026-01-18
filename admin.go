package main

import (
	"log"
	"net/http"
)

func (a *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	if a.platform != "dev" {
		log.Printf("handlerReset: forbidden reset attempt on platform: %s", a.platform)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	err := a.dbQueries.ResetUsers(r.Context())
	if err != nil {
		log.Printf("handlerReset: failed to reset users: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	a.fileserverHits.Store(0)
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Application data reset\n"))
}