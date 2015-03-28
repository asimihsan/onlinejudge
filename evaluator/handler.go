package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func MakeGzipHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			fn(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		fn(gzw, r)
	}
}

func getProblemSummaries(w http.ResponseWriter, r *http.Request) {
	commonHandlerSetup(w)
	if r.Method == "OPTIONS" {
		return
	}

	logger = getLogger(getLogPill())
	logger.Println("handler.getProblemSummaries() entry.")
	defer logger.Println("handler.getProblemSummaries() exit.")

	problems, err := GetProblemSummaries(logger)
	if err != nil {
		log.Panic(err)
	}
	responseEncoded, _ := json.Marshal(problems)
	io.WriteString(w, string(responseEncoded))
}

func getProblemSummary(w http.ResponseWriter, r *http.Request) {
	commonHandlerSetup(w)
	if r.Method == "OPTIONS" {
		return
	}
	vars := mux.Vars(r)
	problem_id := vars["problem_id"]

	logger = getLogger(getLogPill())
	logger.Printf("handler.getProblemSummary() entry. problem_id: %s", problem_id)
	defer logger.Println("handler.getProblemSummary() exit.")

	problem, err := GetProblemSummary(logger, problem_id)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	responseEncoded, _ := json.Marshal(problem)
	io.WriteString(w, string(responseEncoded))
}

func getProblemDetails(w http.ResponseWriter, r *http.Request) {
	commonHandlerSetup(w)
	if r.Method == "OPTIONS" {
		return
	}
	vars := mux.Vars(r)
	problem_id := vars["problem_id"]
	language := vars["language"]

	logger = getLogger(getLogPill())
	logger.Printf("handler.getProblemDetails() entry. problem_id: %s, language: %s",
		problem_id, language)
	defer logger.Println("handler.getProblemDetails() exit.")

	problem, err := GetProblemDetails(logger, problem_id, language)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	responseEncoded, _ := json.Marshal(problem)
	io.WriteString(w, string(responseEncoded))
}

type evaluate_struct struct {
	Code string `json:"code,omitempty"`
}

type runner_response_struct struct {
	Success bool   `json:"success,omitempty"`
	Output  string `json:"output,omitempty"`
}

func evaluate(w http.ResponseWriter, r *http.Request) {
	commonHandlerSetup(w)
	if r.Method == "OPTIONS" {
		return
	}
	vars := mux.Vars(r)
	problem_id := vars["problem_id"]
	language := vars["language"]

	logger = getLogger(getLogPill())
	logger.Printf("handler.evaluate() entry. problem_id: %s, language: %s",
		problem_id, language)
	defer logger.Println("handler.evaluate() exit.")

	// -------------------------------------------------------------------------
	//   Get unit test.
	// -------------------------------------------------------------------------
	problem, err := GetProblemUnitTest(logger, problem_id, language)
	if err != nil {
		http.Error(w, err.Error(), 404)
	}

	// -------------------------------------------------------------------------
	//   Response always written out as JSON.
	// -------------------------------------------------------------------------
	response := map[string]interface{}{}
	defer writeJSONResponse(logger, response, w)
	response["success"] = false

	// -------------------------------------------------------------------------
	//   Decode JSON body.
	// -------------------------------------------------------------------------
	decoder := json.NewDecoder(r.Body)
	var t evaluate_struct
	err = decoder.Decode(&t)
	if err != nil {
		response["output"] = "<could not decode JSON POST request>"
		logger.Printf("Could not decode JSON POST request")
		return
	}

	runner_response, err := CallRunner(language, t.Code, problem.UnitTest[language].Code)
	if err != nil {
		msg := fmt.Sprintf("failed during CallRunner: %s", err)
		response["output"] = msg
		logger.Printf(msg)
		return
	}
	response["success"] = runner_response.Success
	response["output"] = runner_response.Output
}

func CallRunner(language string, code string, unit_test string) (*runner_response_struct, error) {
	logger.Printf("CallRunner entry. language: %s", language)
	defer logger.Printf("CallRunner exit.")
	data := make(map[string]string)
	data["code"] = code
	data["unit_test"] = unit_test

	uri := fmt.Sprintf("https://www.runsomecode.com/run/%s", language)
	j, jerr := json.Marshal(data)
	if jerr != nil {
		return nil, jerr
	}
	request, err := http.NewRequest("POST", uri, bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("HTTP POST not 200: %s\n", resp))
	}

	decoder := json.NewDecoder(resp.Body)
	var t2 runner_response_struct
	err = decoder.Decode(&t2)
	if err != nil {
		return nil, err
	}
	return &t2, nil
}

func CheckProblem(filepath string, language string) error {
	logger.Printf("CheckProblem entry. filepath: %s, language: %s", filepath, language)
	defer logger.Printf("CheckProblem exit.")
	problem, err := ParseProblem(filepath)
	if err != nil {
		log.Printf("failed to load problem: %s", err)
		return err
	}
	code := problem.Solution[language].Code
	unit_test := problem.UnitTest[language].Code
	runner_response, err := CallRunner(language, code, unit_test)
	if err != nil {
		msg := fmt.Sprintf("failed during CallRunner: %s", err)
		logger.Printf(msg)
		return nil
	}
	logger.Printf("runner success: %b", runner_response.Success)
	logger.Printf("runner output: \n%s", runner_response.Output)
	return nil
}

func commonHandlerSetup(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET POST OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
}

func writeJSONResponse(logger *log.Logger, response map[string]interface{}, w http.ResponseWriter) {
	logger.Println("writeJSONResponse() entry")
	defer logger.Println("writeJSONResponse() exit")
	responseEncoded, _ := json.Marshal(response)
	io.WriteString(w, string(responseEncoded))
}
