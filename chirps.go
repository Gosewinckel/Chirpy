package main

import (
	"fmt"
	"net/http"
	"strings"
	"slices"
	"github.com/google/uuid"
	"time"
	"encoding/json"

	"github.com/Gosewinckel/Chirpy/internal/database"
)

type chirpVals struct {
	ID 			uuid.UUID 	`json:"id"`
	CreatedAt 	time.Time 	`json:"created_at"`
	UpdatedAt 	time.Time 	`json:"updated_at"`
	Body 		string 		`json:"body"`
	UserId 		uuid.UUID 	`json:"user_id"`
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


func (conf *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body 	string 		`json:"body"`
		UserId 	uuid.UUID	`json:"user_id"`
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

	chirp, err := conf.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{Body: cleanOutput(params.Body), UserID: params.UserId})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("%s", err))	
	}
	
	payload := chirpVals{chirp.ID, chirp.CreatedAt, chirp.UpdatedAt, chirp.Body, chirp.UserID}
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
	chirps, err := conf.dbQueries.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	payload := []chirpVals{}
	for i := range(chirps) {
		payload = append(payload, 
			chirpVals{
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
	
	payload := chirpVals{
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
