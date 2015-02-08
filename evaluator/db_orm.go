package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	get "github.com/smugmug/godynamo/endpoints/get_item"
	put "github.com/smugmug/godynamo/endpoints/put_item"
	"github.com/smugmug/godynamo/types/attributevalue"
)

func PutProblems(problems []Problem) error {
	log.Printf("PutProblems() entry.")
	defer log.Printf("PutProblems() exit.")
	for _, problem := range problems {
		putProblem(&problem)
	}
	return nil
}

func putProblem(problem *Problem) error {
	log.Printf("putProblem() entry. problem.Id: %s", problem.Id)
	defer log.Printf("putProblem() exit.")

	if err := putProblemIntoProblemSummary(problem, "problem_summary"); err != nil {
		log.Printf("failed to put problem into problem_summary: %s", err)
		return err
	}
	/*
		if err = putProblemIntoProblemDetails(problem, "problem_details"); err != nil {
			log.Printf("failed to put problem into problem_details: %s", err)
			return err
		}
		if err = putProblemIntoUnitTest(problem, "unit_test"); err != nil {
			log.Printf("failed to put problem into unit_test: %s", err)
			return err
		}
	*/
	return nil
}

func putProblemIntoProblemSummary(problem *Problem, table_name string) error {
	log.Printf("putProblemIntoProblemSummary() entry. problem.Id: %s, "+
		"problem.Version: %s, table_name: %s",
		problem.Id, problem.Version, table_name)
	defer log.Printf("putProblemIntoProblemSummary() exit.")

	if same, err := isProblemNewer(problem, table_name); err != nil {
		log.Printf("error while checking if current problem is newer: %s", err)
		return err
	}
	if same == true {
		log.Printf("current problem is same version as existing problem, skip.")
		return nil
	}

	put1 := put.NewPutItem()
	put1.TableName = table_name

	put1.Item["id"] = &attributevalue.AttributeValue{
		S: problem.Id}
	put1.Item["version"] = &attributevalue.AttributeValue{
		N: strconv.Itoa(problem.Version)}
	put1.Item["title"] = &attributevalue.AttributeValue{
		S: problem.Title}
	av := attributevalue.NewAttributeValue()
	for _, language := range problem.SupportedLanguages {
		av.InsertSS(language)
	}
	put1.Item["supported_languages"] = av
	put1.Item["creation_date"] = &attributevalue.AttributeValue{
		S: problem.CreationDate.Format(time.RFC3339)}
	put1.Item["last_updated_date"] = &attributevalue.AttributeValue{
		S: problem.LastUpdatedDate.Format(time.RFC3339)}

	body, code, err := put1.EndpointReq()
	if err != nil || code != http.StatusOK {
		log.Printf("put failed %d %v %s\n", code, err, body)
		return err
	}
	return nil
}

func isProblemNewer(problem *problem, table_name string) (bool, error) {
	log.Printf("isProblemNewer entry. problem.Id: %s, problem.Version: %s, "+
		"table_name: %s", problem.Id, problem.Version, table_name)
	defer log.Printf("isProblemNewer exit.")

	get1 = get.NewGetItem()
	get1.TableName = table_name
	get1.Key["id"] = &attributevalue.AttributeValue{
		S: problem.Id}
	body, code, err := get1.EndpointReq()
	if err != nil || code != http.StatusOK {
		fmt.Printf("get failed %d %v %s\n", code, err, body)
	}

	return true, nil
}
