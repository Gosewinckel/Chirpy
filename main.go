package main

import (
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func main() {
	// create and start server
	serverMux := http.NewServeMux()
	server := &http.Server{
		Addr: 		":8080",
		Handler: 	serverMux,
	}
	conf := apiConfig{}

	// register handler
	serverMux.HandleFunc("GET /api/healthz", serverHealth)
	serverMux.HandleFunc("GET /api/metrics", conf.numRequests)
	serverMux.HandleFunc("POST /api/reset", conf.resetHits)

	// fileserver
	handler := http.FileServer(http.Dir("."))
	serverMux.Handle("/app/", conf.middlewareMetricsInc(http.StripPrefix("/app", handler)))

	err := server.ListenAndServe()
	if err != nil {
		os.Exit(1)
	}
}

// is server running
func serverHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

// responds with total numer of requests since server  turned on
func (conf *apiConfig) numRequests(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Hits: %v", conf.fileServerHits.Load())))
}

func (conf *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conf.fileServerHits.Add(1) 
		next.ServeHTTP(w, r)
	})
}

func (conf *apiConfig) resetHits(w  http.ResponseWriter, r *http.Request) {
	conf.fileServerHits.Store(0)
}
