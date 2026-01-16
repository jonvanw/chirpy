package main

import (
	"log"
	"net/http"
)

func main() {
	const port ="8080"
	const appPrefix = "/app/"
	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}
	appConfig := apiConfig{}

	fileServerHandler := http.StripPrefix(appPrefix, http.FileServer(http.Dir(".")))
	mux.Handle(appPrefix, appConfig.middlewareMetricsInc(fileServerHandler))

	mux.HandleFunc("GET /api/healthz", readinessHandler)

	mux.HandleFunc("POST /api/validate_chirp", HandleValidateChirp)

	mux.HandleFunc("GET /admin/metrics", appConfig.handlerMetrics)

	mux.HandleFunc("POST /admin/reset", appConfig.handlerReset)

	log.Println("Starting server on localhost:8080")

	log.Fatal(server.ListenAndServe())
}
