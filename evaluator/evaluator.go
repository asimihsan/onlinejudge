package main

import (
	"log"
)

func main() {
	log.Printf("main() entry.")
	defer log.Printf("main() exit.")

	Initialize()
	//DeleteTables()
	//CreateTables()
	LoadProblems()

	/*
		problems, err := GetProblemSummaries()
		if err != nil {
			log.Panic(err)
		}
		log.Printf("%s", problems)
	*/

	/*
		problem, err := GetProblemSummary("fizz_buzz")
		if err != nil {
			log.Panic(err)
		}
		log.Printf("%s", problem)
	*/
	/*
		problem, err := GetProblemDetails("fizz_buzz", "python")
		if err != nil {
			log.Panic(err)
		}
		log.Printf("%s", problem)
	*/
	/*
		problem, err := GetProblemUnitTest("fizz_buzz", "python")
		if err != nil {
			log.Panic(err)
		}
		log.Printf("%s", problem)
	*/
}
