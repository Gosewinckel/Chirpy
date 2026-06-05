package main

import (
	"net/http"
	"github.com/google/uuid"
	"time"
	"encoding/json"
	"log"
	"github.com/Gosewinckel/Chirpy/internal/database"
	"github.com/Gosewinckel/Chirpy/internal/auth"
)

func (conf *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email 		string `json:"email"`
		Password 	string `json:"password"`
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
	
	hashed_passwd, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
	}
	sqlPayload := database.CreateUserParams{Email: params.Email, HashedPassword: hashed_passwd}
	user, err := conf.dbQueries.CreateUser(r.Context(), sqlPayload)
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
