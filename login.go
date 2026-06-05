package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Gosewinckel/Chirpy/internal/auth"
	"github.com/google/uuid"
)

func (conf *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password 	string 	`json:"password"`
		Email 		string 	`json:"email"`
	}
	type returnVals struct {
		ID 			uuid.UUID 	`json:"id"`
		CreatedAt	time.Time	`json:"created_at"`
		UpdatedAt 	time.Time	`json:"updated_at"`
		Email 		string 		`json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	user, err := conf.dbQueries.GetUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}
	ok, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || ok != true {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}
	
	payload := returnVals{user.ID, user.CreatedAt, user.UpdatedAt, user.Email}
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}
