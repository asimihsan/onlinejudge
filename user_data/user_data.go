package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/graceful"
)

var (
	logger = GetLogger("logger")
)

func main() {
	logger.Println("main() entry.")

	Initialize()
	//DeleteTables(logger)
	//CreateTables(logger)

	rand.Seed(time.Now().UTC().UnixNano())
	r := mux.NewRouter()

	r.HandleFunc("/user_data/auth/check", MakeGzipHandler(loginCheckHandler)).Methods("POST", "OPTIONS")
	r.HandleFunc("/user_data/auth/login", MakeGzipHandler(loginHandler)).Methods("POST", "OPTIONS")
	r.HandleFunc("/user_data/auth/logout", MakeGzipHandler(loginHandler)).Methods("POST", "OPTIONS")
	r.HandleFunc("/user_data/solution/submit", MakeGzipHandler(solutionSubmitHandler)).Methods("POST", "OPTIONS")
	r.HandleFunc("/user_data/solution/get/{problem_id:[a-z0-9_-]+}/{language:[a-z0-9_]+}", MakeGzipHandler(getSolutions)).Methods("GET", "OPTIONS")
	r.HandleFunc("/user_data/solution/vote/{solution_id:[a-z0-9_-]+}/{type:(up|down)}", MakeGzipHandler(solutionVoteHandler)).Methods("POST", "OPTIONS")
	http.Handle("/", r)

	logger.Printf("Starting HTTP server...")
	graceful.Run("localhost:9001", 10*time.Second, r)
}
