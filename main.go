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

	mux.Handle(appPath, http.StripPrefix(appPath, http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", statusHandler)

	log.Println("Starting server on localhost:8080")

	log.Fatal(server.ListenAndServe())
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}