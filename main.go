package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"strings"
	"slices"
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

	// register handlers
	serverMux.HandleFunc("GET /api/healthz", serverHealth)
	serverMux.HandleFunc("GET /admin/metrics", conf.numRequests)
	serverMux.HandleFunc("POST /admin/reset", conf.resetHits)
	serverMux.HandleFunc("POST /api/validate_chirp", validateChirp)

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
	w.Header().Add("Content-Type", "text/html")
	html := fmt.Sprintf(`
	<html>
	  <body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	  </body>
	</html>
	`, conf.fileServerHits.Load())
	w.Write([]byte(html))
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

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		Cleaned_Body string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
	} else {
		vals := returnVals{
			Cleaned_Body: cleanOutput(params.Body),
		}
		respondWithJSON(w, 200, vals)
	}
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type returnVals struct {
		Error string `json:"error"`
	}
	respBody := returnVals{
		Error: msg,
	}
	data, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marchaling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func cleanOutput(s string) string {
	curses := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(s, " ")
	for i := range(words) {
		if slices.Contains(curses, strings.ToLower(words[i])) {
			words[i] = "****"
		}
	}
	out := strings.Join(words, " ")
	return out
}
