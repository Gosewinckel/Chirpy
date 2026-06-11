package main

import (
	"net/http"
	"strings"
)

func (conf *apiConfig) revoke(w http.ResponseWriter, r *http.Request) {
	head := r.Header.Get("Authorization")
	if head == "" {
		respondWithError(w, 500, "Somethingh went wrong1")
		return
	}
	token := strings.TrimPrefix(head, "Bearer ")
	err := conf.dbQueries.RevokeToken(r.Context(), token)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	w.WriteHeader(204)
}
