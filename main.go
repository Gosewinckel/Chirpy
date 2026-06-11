package main

import (
	"database/sql"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/Gosewinckel/Chirpy/internal/database"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	dbQueries database.Queries
	platform string
	secret string
}

func main() {
	// load .env
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		os.Exit(1)
	}
	dbQueries := database.New(db)

	// create and start server
	serverMux := http.NewServeMux()
	server := &http.Server{
		Addr: 		":8080",
		Handler: 	serverMux,
	}
	conf := apiConfig{}
	conf.dbQueries = *dbQueries
	conf.platform = platform
	conf.secret = secret

	// register handlers
	serverMux.HandleFunc("GET /api/healthz", serverHealth)
	serverMux.HandleFunc("GET /admin/metrics", conf.numRequests)
	serverMux.HandleFunc("POST /admin/reset", conf.resetHits)
	serverMux.HandleFunc("POST /api/users", conf.createUser)
	serverMux.HandleFunc("PUT /api/users", conf.updateUser)
	serverMux.HandleFunc("POST /api/chirps", conf.createChirp)
	serverMux.HandleFunc("GET /api/chirps", conf.getAllChirps)
	serverMux.HandleFunc("GET /api/chirps/{chirpID}", conf.getChirp)
	serverMux.HandleFunc("POST /api/login", conf.login)
	serverMux.HandleFunc("POST /api/refresh", conf.refresh)
	serverMux.HandleFunc("POST /api/revoke", conf.revoke)
	serverMux.HandleFunc("DELETE /api/chirps/{chirpID}", conf.deleteChirp)

	// fileserver
	handler := http.FileServer(http.Dir("."))
	serverMux.Handle("/app/", conf.middlewareMetricsInc(http.StripPrefix("/app", handler)))

	err = server.ListenAndServe()
	if err != nil {
		os.Exit(1)
	}
}

