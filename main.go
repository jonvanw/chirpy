package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/jonvanw/chirpy/internal/database"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	const port ="8080"
	const appPrefix = "/app/"

	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	appConfig := apiConfig{
		dbQueries: database.New(db),
		platform: os.Getenv("PLATFORM"),
		jwtAuthSecret: os.Getenv("JWT_AUTH_SECRET"),
	}

	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}
	

	fileServerHandler := http.StripPrefix(appPrefix, http.FileServer(http.Dir(".")))
	mux.Handle(appPrefix, appConfig.middlewareMetricsInc(fileServerHandler))

	mux.HandleFunc("GET /admin/metrics", appConfig.handlerMetrics)

	mux.HandleFunc("POST /admin/reset", appConfig.handlerReset)

	mux.HandleFunc("GET /api/healthz", readinessHandler)

	mux.HandleFunc("POST /api/chirps", appConfig.handleAddChirp)

	mux.HandleFunc("GET /api/chirps", appConfig.handlerGetChirps)

	mux.HandleFunc("GET /api/chirps/{chirpId}", appConfig.handlerGetChirpById)

	mux.HandleFunc("POST /api/users", appConfig.handleAddUser)

	mux.HandleFunc("POST /api/login", appConfig.handleLogin)

	mux.HandleFunc("POST /api/refresh", appConfig.handleRefreshAuthToken)

	mux.HandleFunc("POST /api/revoke", appConfig.handleRevokeRefreshToken)

	log.Println("Starting server on localhost:8080")

	log.Fatal(server.ListenAndServe())
}

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries 	*database.Queries
	platform string
	jwtAuthSecret string
}
