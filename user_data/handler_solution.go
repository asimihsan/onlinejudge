package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type solutionSubmitRequest struct {
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
	ProblemId   string `json:"problem_id"`
	Language    string `json:"language"`
}

type evaluatorResponse struct {
	Success bool   `json:"success,omitempty"`
	Output  string `json:"output,omitempty"`
}

func solutionSubmitHandler(w http.ResponseWriter, r *http.Request) {
	logger = GetLogger(GetLogPill())
	logger.Printf("handler_solution.solutionSubmitHandler() entry. method: %s", r.Method)
	defer logger.Printf("handler_solution.solutionSubmitHandler() exit.")

	SetCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	response := map[string]interface{}{}
	defer WriteJSONResponse(logger, response, w)
	response["success"] = false

	session, _ := GetCookieStore(r, "persona-session")
	user_id := session.Values["user_id"]
	if user_id == nil {
		error_msg := "user does not have a valid secure cookie set."
		logger.Printf(error_msg)
		w.WriteHeader(401)
		response["error"] = error_msg
		return
	}
	nickname := session.Values["nickname"]
	user_id_value, _ := user_id.(string)
	nickname_value, _ := nickname.(string)
	logger.Printf("user has valid secure cookie set, id: %s, nickname: %s, email: %s",
		user_id_value, nickname_value, session.Values["email"])

	// -------------------------------------------------------------------------
	//   Decode JSON body for request, send request to evaluator, get response.
	// -------------------------------------------------------------------------
	decoder := json.NewDecoder(r.Body)
	var request solutionSubmitRequest
	err := decoder.Decode(&request)
	if err != nil {
		response["success"] = false
		response["output"] = "<could not decode JSON POST request>"
		logger.Panicf("Could not decode JSON POST request")
	}
	logger.Printf("problem_id: %s, language: %s", request.ProblemId, request.Language)

	evaluator_response, err := sendSolutionToEvaluator(logger, &request)
	if err != nil {
		error_msg := fmt.Sprintf("evaluator failed to evaluate the solution: %s", err)
		logger.Printf(error_msg)
		response["error"] = error_msg
		return
	}
	logger.Printf("evaluator returns success: %t", evaluator_response.Success)
	response["success"] = evaluator_response.Success
	if evaluator_response.Success == false {
		return
	}

	if err := putOrUpdateSolution(logger, &request, user_id_value, nickname_value); err != nil {
		error_msg := fmt.Sprintf("user_data failed to put or update solution: %s", err)
		response["error"] = error_msg
		logger.Printf(error_msg)
		return
	}

	return
}

func putOrUpdateSolution(logger *log.Logger, request *solutionSubmitRequest, user_id string, nickname string) error {
	logger.Printf("putOrUpdateSolution() entry.")
	defer logger.Printf("putOrUpdateSolution() exit.")

	solution, err := NewSolution(logger)
	if err != nil {
		error_msg := fmt.Sprintf("Failed to creation new solution during putOrUpdateSolution: %s", err)
		logger.Printf(error_msg)
		return errors.New(error_msg)
	}
	solution.ProblemId = fmt.Sprintf("%s#%s", request.ProblemId, request.Language)
	solution.UserId = user_id
	solution.Nickname = nickname
	solution.Code = request.Code
	solution.Description = request.Description

	existing_solution, err := GetSolutionForProblemAndUser(logger, solution.ProblemId, user_id)
	if _, ok := err.(SolutionForProblemAndUserNotFoundError); ok {
		logger.Printf("user does not have an existing solution.")
	} else if err != nil {
		error_msg := fmt.Sprintf("unexpected error checking if user has an existing solution: %s", err)
		logger.Printf(error_msg)
		return errors.New(error_msg)
	} else {
		logger.Printf("user has an existing solution with ID: %s", existing_solution.SolutionId)
		solution.SolutionId = existing_solution.SolutionId
		solution.CreationDate = existing_solution.CreationDate
	}
	if err := PutSolution(logger, solution); err != nil {
		error_msg := fmt.Sprintf("failed to put solution: %s", err)
		logger.Printf(error_msg)
		return errors.New(error_msg)
	}
	return nil
}

func sendSolutionToEvaluator(logger *log.Logger, request *solutionSubmitRequest) (*evaluatorResponse, error) {
	var (
		response evaluatorResponse
	)

	uri := fmt.Sprintf("http://runsomecode.com/evaluator/evaluate/%s/%s",
		request.ProblemId, request.Language)
	data := make(map[string]string)
	data["code"] = request.Code
	j, jerr := json.Marshal(data)
	if jerr != nil {
		logger.Printf("failed to encode JSON to send to evaluator: %s", jerr)
		return &response, jerr
	}
	post_request, err := http.NewRequest("POST", uri, bytes.NewBuffer(j))
	if err != nil {
		logger.Printf("Failed to create HTTP POST: %s", err)
		return &response, jerr
	}
	post_request.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(post_request)
	if err != nil {
		logger.Println("Failed during HTTP POST")
		return &response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		error_msg := fmt.Sprintf("HTTP GET not 200: %s", resp)
		logger.Printf(error_msg)
		return &response, errors.New(error_msg)
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&response)
	if err != nil {
		logger.Printf("Could not decode JSON response")
		return &response, err
	}
	return &response, err
}
