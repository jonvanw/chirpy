package main

import (
	"net/http"
)

func (a *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	if (a.platform != "dev") {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}

	err := a.dbQueries.ResetUsers(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	a.fileserverHits.Store(0)
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Application data reset\n"))
}