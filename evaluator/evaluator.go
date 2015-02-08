package main

import (
	"log"
)

func main() {
	var (
		err      error
		problems = make([]Problem, 0)
	)

	log.Printf("main() entry.")
	defer log.Printf("main() exit.")

	Initialize()
	//CreateTables()
	//DeleteTables()

	if problems, err = ParseProblems(); err != nil {
		log.Printf("problem parsing problems: %s", err)
		return
	}
	if err = PutProblems(problems); err != nil {
		log.Printf("error whilst putting problems: %s", err)
		return
	}
}
