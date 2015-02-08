package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/smugmug/godynamo/types/item"
)

type Problem struct {
	Id                 string
	Version            int
	Title              string                  `json:",omitempty"`
	SupportedLanguages []string                `json:",omitempty"`
	CreationDate       time.Time               `json:",omitempty"`
	LastUpdatedDate    time.Time               `json:",omitempty"`
	Description        map[string]description  `json:",omitempty"`
	InitialCode        map[string]initial_code `json:",omitempty"`
	UnitTest           map[string]unit_test    `json:",omitempty"`
	Solution           map[string]solution     `json:",omitempty"`
}

type description struct {
	Markdown string
}

type initial_code struct {
	Code string
}

type unit_test struct {
	Code string
}

type solution struct {
	Code string
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
func ItemToProblem(item item.Item) (Problem, error) {
	var problem Problem
	if id, present := item["id"]; present == true {
		problem.Id = id.S
	}
	if version, present := item["version"]; present == true {
		value, err := strconv.Atoi(version.N)
		if err != nil {
			log.Printf("failed to deserialize number for problem: %s", err)
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
			log.Printf("failed to parse creation_date: %s", err)
			return problem, err
		}
		problem.CreationDate = creation_date_object
	}
	if last_updated_date, present := item["last_updated_date"]; present == true {
		last_updated_date_object, err := time.Parse(time.RFC3339, last_updated_date.S)
		if err != nil {
			log.Printf("failed to parse last_updated_date: %s", err)
			return problem, err
		}
		problem.LastUpdatedDate = last_updated_date_object
	}
	return problem, nil
}

func ItemsToProblems(items []item.Item) ([]Problem, error) {
	log.Printf("ItemsToProblems() entry.")
	defer log.Printf("ItemsToProblems() exit.")
	problems := make([]Problem, 0)
	for _, item := range items {
		problem, err := ItemToProblem(item)
		if err != nil {
			log.Printf("error while parsing item: %s", err)
			return problems, err
		}
		problems = append(problems, problem)
	}
	return problems, nil
}

/*
	Id                 string
	Version            int
	Title              string
	SupportedLanguages []string
	CreationDate       time.Time
	LastUpdatedDate    time.Time
	Description        map[string]description
	InitialCode        map[string]initial_code
	UnitTest           map[string]unit_test
	Solution           map[string]solution
*/
