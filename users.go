package main

import (
	"net/http"
	"github.com/google/uuid"
	"time"
	"encoding/json"
	"log"
	"strings"

	"github.com/Gosewinckel/Chirpy/internal/database"
	"github.com/Gosewinckel/Chirpy/internal/auth"
)

func (conf *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email 		string 	`json:"email"`
		Password 	string 	`json:"password"`
	}	
	type returnVals struct {
		ID 			uuid.UUID 	`json:"id"`	
		CreatedAt 	time.Time 	`json:"created_at"`
		UpdatedAt	time.Time 	`json:"updated_at"`
		Email 		string 		`json:"email"`
		IsChirpyRed bool 		`json:"is_chirpy_red"`
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
	
	payload := returnVals{user.ID, user.CreatedAt, user.UpdatedAt, user.Email, user.IsChirpyRed.Bool}
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

func (conf *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email 		string `json:"email"`
		Password 	string `json:"password"`
	}	
	type returnVals struct {
		ID 			uuid.UUID 	`json:"id"`
		CreatedAt	time.Time 	`json:"created_at"`
		UpdatedAt	time.Time 	`json:"updated_at"`
		Email 		string 		`json:"email"`
	}

	head := r.Header.Get("Authorization")
	if head == "" {
		respondWithError(w, 401, "Unauthorized")
		return
	}
	token := strings.TrimPrefix(head, "Bearer ")

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
		return
	}
	
	ident, err := auth.ValidateJWT(token, conf.secret)
	if err != nil {
		respondWithError(w, 401, "Something went wrong")
		return
	}

	sqlPayload := database.UpdateUserParams{Email: params.Email, HashedPassword: hashed_passwd, ID: ident}
	newUser, err := conf.dbQueries.UpdateUser(r.Context(), sqlPayload)
	if err != nil {
		respondWithError(w, 401, "Something went wrong")
		return
	}

	payload := returnVals{newUser.ID, newUser.CreatedAt, newUser.UpdatedAt, newUser.Email}
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (conf *apiConfig) upgradeUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event 	string 	`json:"event"`
		Data struct {	
			UserId 	uuid.UUID	`json:"user_id"`
		}	`json:"data"`
	}

	authorized, err := auth.GetAPIKey(r.Header)
	if authorized != conf.polkaKey {
		respondWithError(w, 401, "")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	if params.Event != "user.upgraded" {
		respondWithError(w, 204, "")
		return
	}

	err = conf.dbQueries.UpgradeRed(r.Context(), params.Data.UserId)
	if err != nil {
		respondWithError(w, 404, "Not found")
		return
	}

	w.WriteHeader(204)
}
