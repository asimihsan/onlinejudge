package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type runner_response_struct struct {
	Success bool   `json:"success,omitempty"`
	Output  string `json:"output,omitempty"`
}

func example() {
	log.Println("example() entry")
	defer log.Printf("example() exit.")

	// Get code from user
	code, _ := ioutil.ReadFile("/tmp/foo.py")

	// Get unit test from DynamoDB
	problem, err := GetProblemUnitTest("fizz_buzz", "python")
	if err != nil {
		log.Panic(err)
	}

	// Prepare data
	data := make(map[string]string)
	data["code"] = string(code)
	data["unit_test"] = problem.UnitTest["python"].Code

	uri := "http://www.runsomecode.com/run/python"
	j, jerr := json.Marshal(data)
	if jerr != nil {
		log.Panic(jerr)
	}
	request, err := http.NewRequest("POST", uri, bytes.NewBuffer(j))
	if err != nil {
		log.Println("Failed to create HTTP POST")
		return
	}
	request.Header.Set("Content-Type", "application/json; charse=utf-8")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Println("Failed during HTTP POST")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("HTTP GET to Google reCAPTCHA not 200: %s\n", resp)
	}

	decoder := json.NewDecoder(resp.Body)
	var t runner_response_struct
	err = decoder.Decode(&t)
	if err != nil {
		log.Panicf("Could not decode Google reCAPTCHA resposne")
	}

	log.Printf("%s", t)
}

func main() {
	log.Printf("main() entry.")
	defer log.Printf("main() exit.")

	Initialize()
	//DeleteTables()
	//CreateTables()
	//LoadProblems()

	example()

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
