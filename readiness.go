package main

import "net/http"

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	w.WriteHeader(status)
	w.Write([]byte(http.StatusText(status)))
}