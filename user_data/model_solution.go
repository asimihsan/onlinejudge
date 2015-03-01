package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nu7hatch/gouuid"
	"github.com/smugmug/godynamo/types/item"
)

type SolutionIdNotFoundError struct {
	SolutionId string
}

func (e SolutionIdNotFoundError) Error() string {
	return fmt.Sprintf("solution ID '%s' not found", e.SolutionId)
}

type Solution struct {
	SolutionId   string    `json:"solution_id"`
	ProblemId    string    `json:"problem_id,omitempty"`
	UserId       string    `json:"user_id,omitempty"`
	Nickname     string    `json:"nickname,omitempty"`
	CreationDate time.Time `json:"creation_date,omitempty"`
	Code         string    `json:"code,omitempty"`
	Description  string    `json:"code,omitempty"`
}

func (s Solution) String() string {
	var (
		out []byte
		err error
	)
	if out, err = json.MarshalIndent(s, "", "  "); err != nil {
		log.Printf("could not marshal problem to JSON: %s", err)
		return "<could_not_marshal>"
	}
	return string(out)
}

func NewSolution(logger *log.Logger) (*Solution, error) {
	var solution Solution
	new_uuid, err := uuid.NewV4()
	if err != nil {
		log.Printf("failed to create new UUID.")
		return &solution, err
	}
	solution = Solution{
		SolutionId: new_uuid.String(),
	}
	return &solution, nil
}

func ItemToSolution(logger *log.Logger, input_solution_id string, item item.Item) (*Solution, error) {
	var solution Solution
	solution_id, present := item["solution_id"]
	if present == false {
		return &solution, SolutionIdNotFoundError{input_solution_id}
	}
	solution.SolutionId = solution_id.S
	if problem_id, present := item["problem_id"]; present == true {
		solution.ProblemId = problem_id.S
	}
	if user_id, present := item["user_id"]; present == true {
		solution.UserId = user_id.S
	}
	if nickname, present := item["nickname"]; present == true {
		solution.Nickname = nickname.S
	}
	if creation_date, present := item["creation_date"]; present == true {
		creation_date_object, err := time.Parse(time.RFC3339, creation_date.S)
		if err != nil {
			logger.Printf("failed to parse creation_date: %s", err)
			return &solution, err
		}
		solution.CreationDate = creation_date_object
	}
	if code_encoded, present := item["code"]; present == true {
		code_decoded, err := DecompressFromBase64(logger, code_encoded.B)
		if err != nil {
			logger.Printf("failed to decompress/decode code: %s", err)
			return &solution, err
		}
		solution.Code = code_decoded
	}
	if description_encoded, present := item["description"]; present == true {
		description_decoded, err := DecompressFromBase64(logger, description_encoded.B)
		if err != nil {
			logger.Printf("failed to decompress/decode description: %s", err)
			return &solution, err
		}
		solution.Description = description_decoded
	}
	return &solution, nil
}

func ItemsToSolutions(logger *log.Logger, items []item.Item) ([]*Solution, error) {
	logger.Printf("model_user.ItemsToSolutions() entry.")
	defer logger.Printf("model_user.ItemsToSolutions() exit.")
	solutions := make([]*Solution, 0)
	for _, item := range items {
		solution, err := ItemToSolution(logger, "", item)
		if err != nil {
			logger.Printf("error while parsing item: %s", err)
			return solutions, err
		}
		solutions = append(solutions, solution)
	}
	return solutions, nil
}
