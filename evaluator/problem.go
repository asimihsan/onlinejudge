package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Problem struct {
	Id                 string
	Version            int
	Title              string
	SupportedLanguages []string
	CreationDate       time.Time
	LastUpdatedDate    time.Time
	Description        string
	InitialCode        map[string]initial_code
	UnitTest           map[string]unit_test
	Solution           map[string]solution
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
