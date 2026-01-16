package main

import (
	"fmt"
	"net/http"
)



func (a *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (a *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	msg := fmt.Sprintf(metricsHtmlTemplate, a.fileserverHits.Load())
	w.Write([]byte(msg))
}

func (a *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	a.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hit counter reset\n"))
}

const metricsHtmlTemplate = 
`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`