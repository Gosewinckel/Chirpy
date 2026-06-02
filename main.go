package main

import (
	"net/http"
	"os"
)

func main() {
	// create and start server
	serverMux := http.NewServeMux()
	server := &http.Server{
		Addr: 		":8080",
		Handler: 	serverMux,
	}

	// register handler
	serverMux.HandleFunc("/healthz", serverHealth)

	// fileserver
	handler := http.FileServer(http.Dir("."))
	serverMux.Handle("/app/", http.StripPrefix("/app", handler))

	err := server.ListenAndServe()
	if err != nil {
		os.Exit(1)
	}
}

func serverHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}
