package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Gosewinckel/Chirpy/internal/auth"
)

func (conf *apiConfig) refresh(w http.ResponseWriter, r *http.Request) {
	type returnVals struct {
		Token string 	`json:"token"`
	}

	head := r.Header.Get("Authorization")
	if head == "" {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	token := strings.TrimPrefix(head, "Bearer ")
	refresh_token, err := conf.dbQueries.GetRefreshToken(context.Background(), token)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	if refresh_token.RevokedAt.Valid {
		respondWithError(w, 401, "revoked token")
		return
	}
	if refresh_token.ExpiresAt.Before(time.Now()) {
		respondWithError(w, 401, "expired")
		return
	}

	user, err := conf.dbQueries.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	authToken, err := auth.MakeJWT(user.ID, conf.secret, time.Duration(1) * time.Hour)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
	}

	payload := returnVals{authToken}
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}
