package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/smugmug/godynamo/types/item"
)

type ProblemNotFoundError struct {
	Id string
}

func (e ProblemNotFoundError) Error() string {
	return fmt.Sprintf("problem ID '%s' not found", e.Id)
}

type Problem struct {
	Id                 string                  `json:"id"`
	Version            int                     `json:"version"`
	Title              string                  `json:"title,omitempty"`
	SupportedLanguages []string                `json:"supported_languages,omitempty"`
	CreationDate       *time.Time              `json:"creation_date,omitempty"`
	LastUpdatedDate    *time.Time              `json:"last_updated_date,omitempty"`
	Description        map[string]description  `json:"description,omitempty"`
	InitialCode        map[string]initial_code `json:"initial_code,omitempty"`
	UnitTest           map[string]unit_test    `json:"unit_test,omitempty"`
	Solution           map[string]solution     `json:"solution,omitempty"`
}

type description struct {
	Markdown string `json:"markdown,omitempty"`
}

type initial_code struct {
	Code string `json:"code,omitempty"`
}

type unit_test struct {
	Code string `json:"code,omitempty"`
}

type solution struct {
	Code string `json:"code,omitempty"`
}

func (p *Problem) GetDescription(language string) (string, bool) {
	value, present := p.Description[language]
	if present == true {
		return value.Markdown, present
	} else {
		value, present := p.Description["_all"]
		if present == true {
			return value.Markdown, present
		}
	}
	return "", false
}

func (p Problem) String() string {
	var (
		out []byte
		err error
	)
	if out, err = json.MarshalIndent(p, "", "  "); err != nil {
		log.Printf("could not marshal problem to JSON: %s", err)
		return "<could_not_marshal>"
	}
	return string(out)
}

func ParseProblems() (problems []Problem, err error) {
	problems = make([]Problem, 0)

	log.Printf("ParseProblems() entry.")
	defer log.Printf("ParseProblems() exit.")

	err = filepath.Walk("./problems/", func(filepath string, f os.FileInfo, err error) error {
		if f.IsDir() || !strings.HasSuffix(filepath, ".toml") {
			return nil
		}
		var problem Problem
		if _, err = toml.DecodeFile(filepath, &problem); err != nil {
			log.Printf("could not decode TOML file: %s", err)
			return err
		}
		log.Printf("found problem with Id: %s, Version: %d",
			problem.Id, problem.Version)
		problems = append(problems, problem)
		return nil
	})
	if err != nil {
		log.Printf("error during directory walk to find problems: %s", err)
		return
	}
	return
}

// An Item is a returned attributebaluemap from godynamo. This function
// deserializes an Item into a Problem. Note that we've split up a Problem
// into multiple tables. Hence the item will be incomplete and the parts
// of the problem you didn't request will be null.
func ItemToProblem(logger *log.Logger, input_id string, item item.Item) (Problem, error) {
	var problem Problem
	id, present := item["id"]
	if present == false {
		return problem, ProblemNotFoundError{input_id}
	}
	problem.Id = id.S
	if version, present := item["version"]; present == true {
		value, err := strconv.Atoi(version.N)
		if err != nil {
			logger.Printf("failed to deserialize number for problem: %s", err)
			return problem, err
		}
		problem.Version = value
	}
	if title, present := item["title"]; present == true {
		problem.Title = title.S
	}
	if supported_languages, present := item["supported_languages"]; present == true {
		problem.SupportedLanguages = supported_languages.SS
	}
	if creation_date, present := item["creation_date"]; present == true {
		creation_date_object, err := time.Parse(time.RFC3339, creation_date.S)
		if err != nil {
			logger.Printf("failed to parse creation_date: %s", err)
			return problem, err
		}
		problem.CreationDate = &creation_date_object
	}
	if last_updated_date, present := item["last_updated_date"]; present == true {
		last_updated_date_object, err := time.Parse(time.RFC3339, last_updated_date.S)
		if err != nil {
			logger.Printf("failed to parse last_updated_date: %s", err)
			return problem, err
		}
		problem.LastUpdatedDate = &last_updated_date_object
	}
	if description_encoded, present := item["description"]; present == true {
		description_decoded, err := decompressFromBase64(logger, description_encoded.B)
		if err != nil {
			logger.Printf("failed to decompress/decode description: %s", err)
			return problem, err
		}
		// Recall that for problem_details or unit_test the id is <problem_id>#<language>
		// Hence we know the language. For problem_summary we don't, but never return
		// anything language-specific.
		language := strings.Split(problem.Id, "#")[1]
		problem.Description = make(map[string]description)
		problem.Description[language] = description{Markdown: description_decoded}
	}
	if initial_code_encoded, present := item["initial_code"]; present == true {
		initial_code_decoded, err := decompressFromBase64(logger, initial_code_encoded.B)
		if err != nil {
			logger.Printf("failed to decompress/decode initial_code: %s", err)
			return problem, err
		}
		// Recall that for problem_details or unit_test the id is <problem_id>#<language>
		// Hence we know the language. For problem_summary we don't, but never return
		// anything language-specific.
		language := strings.Split(problem.Id, "#")[1]
		problem.InitialCode = make(map[string]initial_code)
		problem.InitialCode[language] = initial_code{Code: initial_code_decoded}
	}
	if unit_test_encoded, present := item["unit_test"]; present == true {
		unit_test_decoded, err := decompressFromBase64(logger, unit_test_encoded.B)
		if err != nil {
			logger.Printf("failed to decompress/decode unit_test: %s", err)
			return problem, err
		}
		// Recall that for problem_details or unit_test the id is <problem_id>#<language>
		// Hence we know the language. For problem_summary we don't, but never return
		// anything language-specific.
		language := strings.Split(problem.Id, "#")[1]
		problem.UnitTest = make(map[string]unit_test)
		problem.UnitTest[language] = unit_test{Code: unit_test_decoded}
	}
	if solution_encoded, present := item["solution"]; present == true {
		solution_decoded, err := decompressFromBase64(logger, solution_encoded.B)
		if err != nil {
			logger.Printf("failed to decompress/decode solution: %s", err)
			return problem, err
		}
		// Recall that for problem_details or solution the id is <problem_id>#<language>
		// Hence we know the language. For problem_summary we don't, but never return
		// anything language-specific.
		language := strings.Split(problem.Id, "#")[1]
		problem.Solution = make(map[string]solution)
		problem.Solution[language] = solution{Code: solution_decoded}
	}
	return problem, nil
}

func ItemsToProblems(logger *log.Logger, items []item.Item) ([]Problem, error) {
	logger.Printf("problem.ItemsToProblems() entry.")
	defer logger.Printf("problem.ItemsToProblems() exit.")
	problems := make([]Problem, 0)
	for _, item := range items {
		problem, err := ItemToProblem(logger, "", item)
		if err != nil {
			logger.Printf("error while parsing item: %s", err)
			return problems, err
		}
		problems = append(problems, problem)
	}
	return problems, nil
}
