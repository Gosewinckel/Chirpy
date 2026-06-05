package main

import (
	"net/http"
	"context"
)

func (conf *apiConfig) resetHits(w  http.ResponseWriter, r *http.Request) {
	if conf.platform != "dev" {
		w.WriteHeader(403)
		return
	}
	conf.fileServerHits.Store(0)
	conf.dbQueries.ClearUsers(context.Background())
}
