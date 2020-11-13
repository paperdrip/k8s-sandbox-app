package main

import (
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	var router = newServer()
	log.Println("Starting up server and listening on port 3000")
	log.Fatal(http.ListenAndServe(":3000", handlers.CORS(handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"*"}))(router)))
}

func newServer() http.Handler {
	r := mux.NewRouter().StrictSlash(true)
	r.Use(commondMiddleware)
	r.HandleFunc("/api", handlePost).Methods("POST")
	r.HandleFunc("/api", handleQuery).Methods("GET")
	return r
}

func handlePost(w http.ResponseWriter, r *http.Request) {

}

func handleQuery(w http.ResponseWriter, r *http.Request) {

}

func commondMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
