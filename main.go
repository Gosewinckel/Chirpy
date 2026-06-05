package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"time"
	"context"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/Gosewinckel/Chirpy/internal/database"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	dbQueries database.Queries
	platform string
}

func main() {
	// load .env
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
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

	// register handlers
	serverMux.HandleFunc("GET /api/healthz", serverHealth)
	serverMux.HandleFunc("GET /admin/metrics", conf.numRequests)
	serverMux.HandleFunc("POST /admin/reset", conf.resetHits)
	serverMux.HandleFunc("POST /api/validate_chirp", validateChirp)
	serverMux.HandleFunc("POST /api/users", conf.createUser)
	serverMux.HandleFunc("POST /api/chirps", conf.createChirp)
	serverMux.HandleFunc("GET /api/chirps", conf.getAllChirps)
	serverMux.HandleFunc("GET /api/chirps/{chirpID}", conf.getChirp)

	// fileserver
	handler := http.FileServer(http.Dir("."))
	serverMux.Handle("/app/", conf.middlewareMetricsInc(http.StripPrefix("/app", handler)))

	err = server.ListenAndServe()
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
	if conf.platform != "dev" {
		w.WriteHeader(403)
		return
	}
	conf.fileServerHits.Store(0)
	conf.dbQueries.ClearUsers(context.Background())
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

func (conf *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}	
	type returnVals struct {
		ID 			uuid.UUID 	`json:"id"`	
		CreatedAt 	time.Time 	`json:"created_at"`
		UpdatedAt	time.Time 	`json:"updated_at"`
		Email 		string 		`json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding request body")
		w.WriteHeader(500)
		return
	}

	user, err := conf.dbQueries.CreateUser(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error creating user")
		w.WriteHeader(500)
		return
	}
	
	payload := returnVals{user.ID, user.CreatedAt, user.UpdatedAt, user.Email}
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling payload")
		w.WriteHeader(500)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)
}

func (conf *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body 	string 		`json:"body"`
		UserId 	uuid.UUID	`json:"user_id"`
	}
	type returnVals struct {
		ID 			uuid.UUID 	`json:"id"`
		CreatedAt 	time.Time 	`json:"created_at"`
		UpdatedAt 	time.Time 	`json:"updated_at"`
		Body 		string 		`json:"body"`
		UserId 		uuid.UUID 	`json:"user_id"`
	}
	
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong 1")
	}
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
	}

	chirp, err := conf.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{Body: params.Body, UserID: params.UserId})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("%s", err))	
	}
	
	payload := returnVals{chirp.ID, chirp.CreatedAt, chirp.UpdatedAt, chirp.Body, chirp.UserID}
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 500, "Something went wrong 3")
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)
}

func (conf *apiConfig) getAllChirps(w http.ResponseWriter, r *http. Request) {
	type parameters struct {
		ID 			uuid.UUID 	`json:"id"`
		CreatedAt 	time.Time 	`json:"created_at"`
		UpdatedAt 	time.Time 	`json:"updated_at"`
		Body 		string 		`json:"body"`
		UserId 		uuid.UUID 	`json:"user_id"`
	}

	chirps, err := conf.dbQueries.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	payload := []parameters{}
	for i := range(chirps) {
		payload = append(payload, 
			parameters{
				ID: chirps[i].ID,
				CreatedAt: chirps[i].CreatedAt,
				UpdatedAt: chirps[i].UpdatedAt,
				Body: chirps[i].Body,
				UserId: chirps[i].UserID,
			},
		)
	}
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
	} 
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (conf *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		ID 			uuid.UUID 	`json:"id"`
		CreatedAt 	time.Time 	`json:"created_at"`
		UpdatedAt 	time.Time 	`json:"updated_at"`
		Body 		string 		`json:"body"`
		UserId 		uuid.UUID 	`json:"user_id"`
	}

	id := r.PathValue("chirpID")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	chirp, err := conf.dbQueries.GetChirp(r.Context(), parsedID)
	if err != nil {
		respondWithError(w, 404, "Chirp not found")
		return
	}
	
	payload := parameters{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserId: chirp.UserID,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 500, ":Something went wrong 2")
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}
