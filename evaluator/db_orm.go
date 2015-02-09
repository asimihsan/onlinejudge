package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	get "github.com/smugmug/godynamo/endpoints/get_item"
	put "github.com/smugmug/godynamo/endpoints/put_item"
	scan "github.com/smugmug/godynamo/endpoints/scan"
	"github.com/smugmug/godynamo/types/attributevalue"
)

func GetProblemSummaries() ([]Problem, error) {
	log.Printf("GetProblemSummaries() entry.")
	defer log.Printf("GetProblemSummaries() exit.")

	s := scan.NewScan()
	tablename := "problem_summary"
	s.TableName = tablename

	body, code, err := s.EndpointReq()
	if err != nil || code != http.StatusOK {
		log.Printf("scan failed %d %v %s\n", code, err, body)
		return nil, err
	}

	// Response is the full response from DynamoDB; Item and ConsumedCapacity
	// DyanmoDB does e.g. {"S": "foobar"} for strings etc. in Item.

	var resp scan.Response
	um_err := json.Unmarshal([]byte(body), &resp)
	if um_err != nil {
		e := fmt.Sprintf("unmarshal Response: %v", um_err)
		log.Printf("%s\n", e)
		return make([]Problem, 0), um_err
	}

	problems, err := ItemsToProblems(resp.Items)
	if err != nil {
		log.Printf("error while converting items to problems: %s", err)
		return problems, err
	}

	return problems, nil
}

func GetProblemSummary(problem_id string) (Problem, error) {
	log.Printf("GetProblemSummary() entry. problem_id: %s", problem_id)
	defer log.Printf("GetProblemSummary() exit.")
	var problem Problem

	get1 := get.NewGetItem()
	get1.TableName = "problem_summary"
	get1.Key["id"] = &attributevalue.AttributeValue{
		S: problem_id}
	body, code, err := get1.EndpointReq()
	if err != nil || code != http.StatusOK {
		log.Printf("get failed %d %v %s\n", code, err, body)
		return problem, err
	}

	// Response is the full response from DynamoDB; Item and ConsumedCapacity
	// DyanmoDB does e.g. {"S": "foobar"} for strings etc. in Item.
	resp := get.NewResponse()
	um_err := json.Unmarshal([]byte(body), resp)
	if um_err != nil {
		log.Printf("failed to unmarshal DynamoDB response (%s): %s", resp, um_err)
		return problem, um_err
	}

	problem, err = ItemToProblem(resp.Item)
	if err != nil {
		log.Printf("error while converting item to problem: %s", err)
		return problem, err
	}

	return problem, nil
}

func GetProblemDetails(problem_id string, language string) (Problem, error) {
	log.Printf("GetProblemDetails() entry. problem_id: %s, language: %s", problem_id, language)
	defer log.Printf("GetProblemDetails() exit.")
	var problem Problem

	get1 := get.NewGetItem()
	get1.TableName = "problem_details"
	id := fmt.Sprintf("%s#%s", problem_id, language)
	get1.Key["id"] = &attributevalue.AttributeValue{
		S: id}
	body, code, err := get1.EndpointReq()
	if err != nil || code != http.StatusOK {
		log.Printf("get failed %d %v %s\n", code, err, body)
		return problem, err
	}

	// Response is the full response from DynamoDB; Item and ConsumedCapacity
	// DyanmoDB does e.g. {"S": "foobar"} for strings etc. in Item.
	resp := get.NewResponse()
	um_err := json.Unmarshal([]byte(body), resp)
	if um_err != nil {
		log.Printf("failed to unmarshal DynamoDB response (%s): %s", resp, um_err)
		return problem, um_err
	}

	problem, err = ItemToProblem(resp.Item)
	if err != nil {
		log.Printf("error while converting item to problem: %s", err)
		return problem, err
	}

	return problem, nil

}

func GetProblemUnitTest(problem_id string, language string) (Problem, error) {
	log.Printf("GetProblemUnitTest() entry. problem_id: %s, language: %s", problem_id, language)
	defer log.Printf("GetProblemUnitTest() exit.")
	var problem Problem

	get1 := get.NewGetItem()
	get1.TableName = "unit_test"
	id := fmt.Sprintf("%s#%s", problem_id, language)
	get1.Key["id"] = &attributevalue.AttributeValue{
		S: id}
	body, code, err := get1.EndpointReq()
	if err != nil || code != http.StatusOK {
		log.Printf("get failed %d %v %s\n", code, err, body)
		return problem, err
	}

	// Response is the full response from DynamoDB; Item and ConsumedCapacity
	// DyanmoDB does e.g. {"S": "foobar"} for strings etc. in Item.
	resp := get.NewResponse()
	um_err := json.Unmarshal([]byte(body), resp)
	if um_err != nil {
		log.Printf("failed to unmarshal DynamoDB response (%s): %s", resp, um_err)
		return problem, um_err
	}

	problem, err = ItemToProblem(resp.Item)
	if err != nil {
		log.Printf("error while converting item to problem: %s", err)
		return problem, err
	}

	return problem, nil
}

func LoadProblems() error {
	problems, err := ParseProblems()
	if err != nil {
		log.Printf("problem parsing problems: %s", err)
		return err
	}
	if err = PutProblems(problems); err != nil {
		log.Printf("error whilst putting problems: %s", err)
		return err
	}
	return nil
}

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
	if err := putProblemIntoProblemDetails(problem, "problem_details"); err != nil {
		log.Printf("failed to put problem into problem_details: %s", err)
		return err
	}
	if err := putProblemIntoUnitTest(problem, "unit_test"); err != nil {
		log.Printf("failed to put problem into unit_test: %s", err)
		return err
	}
	return nil
}

func putProblemIntoProblemSummary(problem *Problem, table_name string) error {
	log.Printf("putProblemIntoProblemSummary() entry. problem.Id: %s, "+
		"problem.Version: %s, table_name: %s",
		problem.Id, strconv.Itoa(problem.Version), table_name)
	defer log.Printf("putProblemIntoProblemSummary() exit.")

	var (
		err  error
		same bool
	)

	if same, err = isProblemNewer(problem, problem.Id, table_name); err != nil {
		log.Printf("error while checking if current problem is newer: %s", err)
		return err
	}
	if same == true {
		log.Printf("current problem is same version as existing problem, skip.")
		return nil
	}
	log.Printf("current problem is newer than existing problem, continue.")

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

func putProblemIntoProblemDetails(problem *Problem, table_name string) error {
	log.Printf("putProblemIntoProblemDetails() entry. problem.Id: %s, "+
		"problem.Version: %s, table_name: %s",
		problem.Id, strconv.Itoa(problem.Version), table_name)
	defer log.Printf("putProblemIntoProblemDetails() exit.")

	var (
		err  error
		same bool
	)

	for _, language := range problem.SupportedLanguages {
		log.Printf("handling language: %s", language)

		id := fmt.Sprintf("%s#%s", problem.Id, language)
		if same, err = isProblemNewer(problem, id, table_name); err != nil {
			log.Printf("error while checking if current problem is newer: %s", err)
			return err
		}
		if same == true {
			log.Printf("current problem is same version as existing problem, skip.")
			return nil
		}
		log.Printf("current problem is newer than existing problem, continue.")

		put1 := put.NewPutItem()
		put1.TableName = table_name

		put1.Item["id"] = &attributevalue.AttributeValue{
			S: id}
		put1.Item["version"] = &attributevalue.AttributeValue{
			N: strconv.Itoa(problem.Version)}
		description, present := problem.GetDescription(language)
		if present == false {
			log.Printf("no description present for language %s, problem.Id %s!",
				language, problem.Id)
			continue
		}
		compressed_description, err := compressToBase64(description)
		if err != nil {
			log.Printf("failed to compress description for language %s, problem.Id %s!", language, problem.Id)
			continue
		}
		put1.Item["description"] = &attributevalue.AttributeValue{
			B: compressed_description}
		initial_code, present := problem.InitialCode[language]
		if present == true {
			compressed_initial_code, err := compressToBase64(initial_code.Code)
			if err != nil {
				log.Printf("failed to compress initial_code for language %s, problem.Id %s!", language, problem.Id)
				continue
			}
			put1.Item["initial_code"] = &attributevalue.AttributeValue{
				B: compressed_initial_code}
		}
		body, code, err := put1.EndpointReq()
		if err != nil || code != http.StatusOK {
			log.Printf("put failed %d %v %s\n", code, err, body)
			return err
		}
	}
	return nil
}

func putProblemIntoUnitTest(problem *Problem, table_name string) error {
	log.Printf("putProblemIntoUnitTest() entry. problem.Id: %s, "+
		"problem.Version: %s, table_name: %s",
		problem.Id, strconv.Itoa(problem.Version), table_name)
	defer log.Printf("putProblemIntoUnitTest() exit.")

	var (
		err  error
		same bool
	)

	for _, language := range problem.SupportedLanguages {
		log.Printf("handling language: %s", language)

		id := fmt.Sprintf("%s#%s", problem.Id, language)
		if same, err = isProblemNewer(problem, id, table_name); err != nil {
			log.Printf("error while checking if current problem is newer: %s", err)
			return err
		}
		if same == true {
			log.Printf("current problem is same version as existing problem, skip.")
			return nil
		}
		log.Printf("current problem is newer than existing problem, continue.")

		put1 := put.NewPutItem()
		put1.TableName = table_name

		put1.Item["id"] = &attributevalue.AttributeValue{
			S: id}
		put1.Item["version"] = &attributevalue.AttributeValue{
			N: strconv.Itoa(problem.Version)}
		unit_test, present := problem.UnitTest[language]
		if present == true {
			compressed_unit_test, err := compressToBase64(unit_test.Code)
			if err != nil {
				log.Printf("failed to compress unit_test for language %s, problem.Id %s!", language, problem.Id)
				continue
			}
			put1.Item["unit_test"] = &attributevalue.AttributeValue{
				B: compressed_unit_test}
		}
		body, code, err := put1.EndpointReq()
		if err != nil || code != http.StatusOK {
			log.Printf("put failed %d %v %s\n", code, err, body)
			return err
		}
	}
	return nil
}

func isProblemNewer(problem *Problem, id string, table_name string) (bool, error) {
	log.Printf("isProblemNewer entry. problem.Id: %s, id: %s, problem.Version: %s, "+
		"table_name: %s", problem.Id, id, strconv.Itoa(problem.Version), table_name)
	defer log.Printf("isProblemNewer exit.")

	get1 := get.NewGetItem()
	get1.TableName = table_name
	get1.Key["id"] = &attributevalue.AttributeValue{
		S: id}
	get1.AttributesToGet = append(get1.AttributesToGet, "version")
	body, code, err := get1.EndpointReq()
	if err != nil || code != http.StatusOK {
		log.Printf("get failed %d %v %s\n", code, err, body)
	}

	// Response is the full response from DynamoDB; Item and ConsumedCapacity
	// DyanmoDB does e.g. {"S": "foobar"} for strings etc. in Item.
	resp := get.NewResponse()
	um_err := json.Unmarshal([]byte(body), resp)
	if um_err != nil {
		log.Printf("failed to unmarshal DynamoDB response (%s): %s", resp, um_err)
		return false, um_err
	}

	existing_problem, err := ItemToProblem(resp.Item)
	if err != nil {
		log.Printf("error while converting item to existing_problem: %s", err)
		return false, err
	}

	rc := existing_problem.Version == problem.Version
	log.Printf("returning: %t", rc)
	return rc, nil
}

func compressToBase64(input string) (string, error) {
	var b bytes.Buffer
	gz, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		log.Printf("failed to create gzip writer: %s", err)
		return "", err
	}
	if _, err := gz.Write([]byte(input)); err != nil {
		log.Printf("failed to compress string: %s", err)
		return "", err
	}
	if err := gz.Flush(); err != nil {
		log.Printf("failed to flush gzip: %s", err)
		return "", err
	}
	if err := gz.Close(); err != nil {
		log.Printf("failed to close gzip: %s", err)
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(b.Bytes())
	return encoded, nil
}

func decompressFromBase64(input string) (string, error) {
	log.Printf("decompressFromBase64() entry.")
	defer log.Printf("decompressFromBase64() exit.")

	// DecodedLen returns the maximum length in bytes of the decoded
	// data. But this is a maximum. You must use the 'n' return value
	// from the Decode call to know exactly how many bytes to use. If
	// you don't you'll feed the gzip reader garbage nulls at the end.
	// (Recall that base64 must be padded to the nearest 4 bytes).
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(input)))
	n, err := base64.StdEncoding.Decode(decoded, []byte(input))
	if err != nil {
		log.Printf("failed to base64 decode: %s", err)
		return "", err
	}
	gz, err := gzip.NewReader(bytes.NewBuffer(decoded[:n]))
	if err != nil {
		log.Printf("failed to create gzip reader: %s", err)
		return "", err
	}
	uncompressed, _ := ioutil.ReadAll(gz)
	if err != nil {
		log.Printf("failed to decompress string: %s", err)
		return "", err
	}
	if err := gz.Close(); err != nil {
		log.Printf("failed to close gzip: %s", err)
		return "", err
	}
	return string(uncompressed), nil
}
