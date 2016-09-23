package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/skeswa/gophr/common/config"
)

func main() {
	var (
		r        = mux.NewRouter()
		conf     = config.GetConfig()
		endpoint = fmt.Sprintf(
			"/api/repos/{%s}/{%s}/{%s}",
			urlVarAuthor,
			urlVarRepo,
			urlVarSHA)
	)

	// Register the status route.
	r.HandleFunc("/api/status", StatusHandler()).Methods("GET")
	// Register all the remaining routes for the main endpoint.
	r.HandleFunc(endpoint, RepoExistsHandler(conf)).Methods("GET")
	r.HandleFunc(endpoint, CreateRepoHandler(conf)).Methods("POST")
	r.HandleFunc(endpoint, DeleteRepoHandler(conf)).Methods("DELETE")

	// Start serving.
	log.Printf("Servicing HTTP requests on port %d.\n", conf.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), r)
}
