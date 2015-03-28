package main

/*
// JSON from browser does OPTIONS before GET/POST
curl -X OPTIONS -H "Content-Type: application/json" --compressed \
    http://localhost:8081/evaluator/get_problem_summaries

curl -X GET -H "Content-Type: application/json" --compressed \
    http://localhost:8081/evaluator/get_problem_summaries

curl -X GET -H "Content-Type: application/json" --compressed \
    http://localhost:8081/evaluator/get_problem_summary/fizz_buzz

// Will return 404
curl -X GET -H "Content-Type: application/json" --compressed \
    http://localhost:8081/evaluator/get_problem_summary/fizz_buzz2

curl -X GET -H "Content-Type: application/json" --compressed \
    http://localhost:8081/evaluator/get_problem_details/fizz_buzz/python

curl -X OPTIONS -H "Content-Type: application/json" --compressed \
    --data-binary @foo.py http://localhost:8081/evaluator/evaluate/fizz_buzz/python

curl -X POST -H "Content-Type: application/json" --compressed \
    --data-binary @/tmp/foo.json http://localhost:8081/evaluator/evaluate/fizz_buzz/python
*/

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/graceful"
)

var (
	logger               = getLogger("logger")
	letters              = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	usage                = "CRUD for problems, evaluate problem solutions with unit tests."
	loadProblems         = flag.Bool("load-problems", false, "Load problems to DynamoDB.")
	recreateTables       = flag.Bool("recreate-tables", false, "Delete then create tables in DynamoDB")
	checkProblemFilepath = flag.String("check-problem-filepath", "", "Check problem at specific filepath")
	checkProblemLanguage = flag.String("check-problem-language", "", "Check problem using specific language")
)

func getLogger(prefix string) *log.Logger {
	paddedPrefix := fmt.Sprintf("%-8s: ", prefix)
	return log.New(os.Stdout, paddedPrefix,
		log.Ldate|log.Ltime|log.Lmicroseconds)
}

func getLogPill() string {
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	logger.Println("main() entry.")
	flag.Parse()

	if *checkProblemFilepath != "" {
		if *checkProblemLanguage == "" {
			logger.Printf("Both check-problem-filepath and check-problem-language are required.")
			return
		}
		CheckProblem(*checkProblemFilepath, *checkProblemLanguage)
		return
	}

	Initialize()

	if *recreateTables {
		DeleteTables()
		CreateTables()
		return
	}

	if *loadProblems {
		LoadProblems(logger)
		return
	}

	rand.Seed(time.Now().UTC().UnixNano())
	r := mux.NewRouter()

	r.HandleFunc("/evaluator/get_problem_summaries",
		MakeGzipHandler(getProblemSummaries)).Methods("GET", "OPTIONS")
	r.HandleFunc("/evaluator/get_problem_summary/{problem_id:[a-z0-9_]+}",
		MakeGzipHandler(getProblemSummary)).Methods("GET", "OPTIONS")
	r.HandleFunc("/evaluator/get_problem_details/{problem_id:[a-z0-9_]+}/{language:[a-z0-9_]+}",
		MakeGzipHandler(getProblemDetails)).Methods("GET", "OPTIONS")
	r.HandleFunc("/evaluator/evaluate/{problem_id:[a-z0-9_]+}/{language:[a-z0-9_]+}",
		MakeGzipHandler(evaluate)).Methods("POST", "OPTIONS")
	http.Handle("/", r)

	graceful.Run("localhost:8081", 10*time.Second, r)
}
