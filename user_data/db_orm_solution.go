package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/smugmug/godynamo/endpoints/batch_write_item"
	get "github.com/smugmug/godynamo/endpoints/get_item"
	put "github.com/smugmug/godynamo/endpoints/put_item"
	"github.com/smugmug/godynamo/types/attributevalue"
	"github.com/smugmug/godynamo/types/item"
)

func executePutItem(put_item *put.PutItem) error {
	body, code, err := put_item.EndpointReq()
	if err != nil || code != http.StatusOK {
		log.Printf("put failed %d %v %s\n", code, err, body)
		return err
	}
	return nil
}

func GetSolutionForProblemAndUser(logger *log.Logger, problem_id string, user_id string) (*Solution, error) {
	logger.Printf("db_orm_solution.GetSolutionForProblemAndUser() entry. problem_id: %s, user_id: %s", problem_id, user_id)
	defer logger.Printf("db_orm_solution.GetSolutionForProblemAndUser() exit.")

	solution := &Solution{}

	get1 := get.NewGetItem()
	get1.TableName = "solution"
	get1.Key["problem_id"] = &attributevalue.AttributeValue{S: problem_id}
	get1.Key["user_id"] = &attributevalue.AttributeValue{S: user_id}
	body, code, err := get1.EndpointReq()
	if err != nil || code != http.StatusOK {
		logger.Printf("get failed %d %v %s\n", code, err, body)
		return solution, err
	}
	// Response is the full response from DynamoDB; Item and ConsumedCapacity
	// DyanmoDB does e.g. {"S": "foobar"} for strings etc. in Item.
	resp := get.NewResponse()
	um_err := json.Unmarshal([]byte(body), resp)
	if um_err != nil {
		logger.Printf("failed to unmarshal DynamoDB response (%s): %s", resp, um_err)
		return solution, um_err
	}
	solution, err = ItemToSolution(logger, problem_id, user_id, resp.Item)
	if err != nil {
		logger.Printf("error while converting item to solution: %s", err)
		return solution, err
	}
	return solution, nil
}

func PutSolution(logger *log.Logger, solution *Solution) error {
	logger.Printf("db_orm_solution.PutSolution() entry. solution.SolutionId: %s", solution.SolutionId)
	defer logger.Printf("db_orm_solution.PutSolution() exit.")

	var (
		solution_table_name = "solution"
	)

	put1, err := getPutSolutionIntoSolution(logger, solution)
	if err != nil {
		logger.Printf("failed to get put item for putting solution into %s: %s", solution_table_name, err)
		return err
	}

	b := batch_write_item.NewBatchWriteItem()
	b.RequestItems[solution_table_name] = make([]batch_write_item.RequestInstance, 0)
	b.RequestItems[solution_table_name] = append(
		b.RequestItems[solution_table_name],
		batch_write_item.RequestInstance{PutRequest: &batch_write_item.PutRequest{Item: *put1}})
	bs, _ := batch_write_item.Split(b)
	for _, bsi := range bs {
		body, code, err := bsi.RetryBatchWrite(0)
		if err != nil || code != http.StatusOK {
			logger.Printf("error: %v\n%v\n%v\n", string(body), code, err)
			return err
		} else {
			logger.Printf("worked!: %v\n%v\n%v\n", string(body), code, err)
		}
	}

	return nil
}

func getPutSolutionIntoSolution(logger *log.Logger, solution *Solution) (*item.Item, error) {
	logger.Printf("db_orm_solution.getPutSolutionIntoSolution() entry. solution.SolutionId: %s", solution.SolutionId)
	defer logger.Printf("db_orm_solution.getPutSolutionIntoSolution() exit.")

	put_item := item.NewItem()
	put_item["solution_id"] = &attributevalue.AttributeValue{S: solution.SolutionId}
	put_item["problem_id"] = &attributevalue.AttributeValue{S: solution.ProblemId}
	put_item["user_id"] = &attributevalue.AttributeValue{S: solution.UserId}
	put_item["nickname"] = &attributevalue.AttributeValue{S: solution.Nickname}
	compressed_description, err := CompressToBase64(logger, solution.Description)
	if err != nil {
		logger.Printf("failed to compress description for solution.SolutionId %s!", solution.SolutionId)
		return &put_item, err
	}
	put_item["description"] = &attributevalue.AttributeValue{B: compressed_description}
	compressed_code, err := CompressToBase64(logger, solution.Code)
	if err != nil {
		logger.Printf("failed to compress code for solution.SolutionId %s!", solution.SolutionId)
		return &put_item, err
	}
	put_item["code"] = &attributevalue.AttributeValue{B: compressed_code}
	put_item["up"] = &attributevalue.AttributeValue{N: "0"}
	put_item["down"] = &attributevalue.AttributeValue{N: "0"}
	put_item["creation_date"] = &attributevalue.AttributeValue{S: solution.CreationDate.Format(time.RFC3339)}
	put_item["last_updated_date"] = &attributevalue.AttributeValue{S: solution.LastUpdatedDate.Format(time.RFC3339)}
	return &put_item, nil
}
