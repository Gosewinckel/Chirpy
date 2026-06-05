package main

import (
	"fmt"
	"net/http"
)

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
