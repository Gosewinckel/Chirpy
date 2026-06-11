package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Gosewinckel/Chirpy/internal/auth"
	"github.com/Gosewinckel/Chirpy/internal/database"
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
		Token 		string 		`json:"token"`
		RefreshToken string 	`json:"refresh_token"`
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
	
	token, err := auth.MakeJWT(user.ID, conf.secret, time.Duration(3600) * time.Second)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	refreshToken := auth.MakeRefreshToken()
	refreshParams := database.CreateRefreshTokenParams{Token: refreshToken, UserID: user.ID, ExpiresAt: time.Now().Add(60 * 24 * time.Hour)}
	_, err = conf.dbQueries.CreateRefreshToken(r.Context(), refreshParams)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	payload := returnVals{user.ID, user.CreatedAt, user.UpdatedAt, user.Email, token, refreshToken}
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}
