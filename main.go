package main

import (
	"log"
	"net/http"
)

func main() {
	const port ="8080"
	const appPath = "/app/"
	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}
	appConfig := apiConfig{}

	fileServerHandler := http.StripPrefix(appPath, http.FileServer(http.Dir(".")))
	mux.Handle(appPath, appConfig.middlewareMetricsInc(fileServerHandler))
	mux.HandleFunc("/healthz", readinessHandler)

	mux.HandleFunc("/metrics", appConfig.handlerMetrics)

	mux.HandleFunc("/reset", appConfig.handlerReset)

	log.Println("Starting server on localhost:8080")

	log.Fatal(server.ListenAndServe())
}
