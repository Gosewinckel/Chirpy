package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Gosewinckel/Chirpy/internal/auth"
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

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 500, "Somethin went wrong")
		return
	}
	user, err := auth.ValidateJWT(token, conf.secret)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	chirp, err := conf.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{Body: cleanOutput(params.Body), UserID: user})
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
	// query author
	author := r.URL.Query().Get("author_id")
	var chirps []database.Chirp
	var err error
	if author != "" {
		authorID, err := uuid.Parse(author)
		if err != nil {
			respondWithError(w, 500, "Something wemt wrong")
			return
		}
		chirps, err = conf.dbQueries.GetChirpByAuthor(r.Context(), authorID)
		if err != nil {
			respondWithError(w, 500, "Something went wrong")
			return
		}
	} else {
		chirps, err = conf.dbQueries.GetAllChirps(r.Context())
		if err != nil {
			respondWithError(w, 500, "Something went wrong")
			return
		}
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

	// query sort
	sortOrder := r.URL.Query().Get("sort")
	if sortOrder == "desc" {
		sort.Slice(payload, func(i, j int) bool {return payload[i].CreatedAt.After(payload[j].CreatedAt)})
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

func (conf *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	head := r.Header.Get("Authorization")
	if head == "" {
		respondWithError(w, 401, "Unauthorized")
		return
	}
	token := strings.TrimPrefix(head, "Bearer ")
	identification, err := auth.ValidateJWT(token, conf.secret)
	if err != nil {
		respondWithError(w, 403, "Unauthorized")
		return
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
	
	if identification != chirp.UserID {
		respondWithError(w, 403, "Unauthorized")
		return
	}

	err = conf.dbQueries.DeleteChirp(r.Context(), chirp.ID)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	w.WriteHeader(204)
}
