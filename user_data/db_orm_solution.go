package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	ep "github.com/smugmug/godynamo/endpoint"
	"github.com/smugmug/godynamo/endpoints/batch_write_item"
	get "github.com/smugmug/godynamo/endpoints/get_item"
	put "github.com/smugmug/godynamo/endpoints/put_item"
	"github.com/smugmug/godynamo/endpoints/query"
	"github.com/smugmug/godynamo/types/attributevalue"
	"github.com/smugmug/godynamo/types/condition"
	"github.com/smugmug/godynamo/types/item"
)

var (
	VoteOfSameTypeAlreadyExists = errors.New("Vote of same type already exists.")
)

func PutSolutionVote(logger *log.Logger, user_id string, problem_id string, solution_id string, vote_type string) (bool, error) {
	logger.Printf("db_orm_solution.PutSolutionVote() entry. user_id: %s, problem_id: %s, solution_id: %s, vote_type: %s", user_id, problem_id, solution_id, vote_type)
	defer logger.Printf("db_orm_solution.PutSolutionVote() exit.")

	put_item, err := getPutVoteIntoUserVote(logger, user_id, problem_id, solution_id, vote_type)
	if err != nil {
		logger.Printf("failed to build put_item for putting into user_vote: %s", err)
		return false, err
	}
	response, err := executePut(logger, put_item)
	if err != nil {
		logger.Printf("failed to execute put_item for putting into user_vote: %s", err)
		if err == VoteOfSameTypeAlreadyExists {
			logger.Printf("Vote of same type exists isn't a true error, tell client call succeeded")
			return true, nil
		}
		return false, err
	}

	// Since we requested ReturnValues = ALL_OLD on Put, response.Attributes
	// will be a map of the old item. If this is empty there was no old item.
	logger.Printf("length of resp attrs: %s", len(response.Attributes))

	return true, nil
}

func executePut(logger *log.Logger, put_item *put.PutItem) (*put.Response, error) {
	body, code, err := put_item.EndpointReq()
	if err != nil || code != http.StatusOK {
		if strings.Contains(string(body), "ConditionalCheckFailedException") {
			logger.Printf("Vote of same type already exists.")
			return nil, VoteOfSameTypeAlreadyExists
		}
		logger.Printf("failed to execute put into user_vote: %s, %s, %s", body, code, err)
		return nil, err
	}

	// Response is the full response from DynamoDB; Item and ConsumedCapacity
	// DyanmoDB does e.g. {"S": "foobar"} for strings etc. in Item.
	var resp put.Response
	um_err := json.Unmarshal([]byte(body), &resp)
	if um_err != nil {
		e := fmt.Sprintf("unmarshal Response: %v", um_err)
		logger.Printf("%s\n", e)
		return nil, um_err
	}

	return &resp, nil
}

// Put a row into user_vote to record that the user made a vote
// -	Set ReturnValues to ALL_OLD to return the old vote. If no vote was ever
// 		made then response.Attributes will be an empty map, else will be a
// 		non-empty map.
// -	Set a condition for any existing vote to not be the same. That way if
// 		the user vote again in the same way we don't update the main solution
// 		vote count.
func getPutVoteIntoUserVote(logger *log.Logger, user_id string, problem_id string, solution_id string, vote_type string) (*put.PutItem, error) {
	logger.Printf("db_orm_solution.getPutVoteIntoUserVote() entry. user_id: %s, problem_id: %s, solution_id: %s, vote_type: %s", user_id, problem_id, solution_id, vote_type)
	defer logger.Printf("db_orm_solution.getPutVoteIntoUserVote() exit.")

	if vote_type != "u" && vote_type != "d" {
		error_msg := "vote_type must be 'u' or 'd'"
		logger.Printf(error_msg)
		return nil, errors.New(error_msg)
	}

	put_item := put.NewPutItem()
	put_item.TableName = "user_vote"
	put_item.Item["user_vote_id"] = &attributevalue.AttributeValue{S: fmt.Sprintf("%s#%s", user_id, problem_id)}
	put_item.Item["solution_id"] = &attributevalue.AttributeValue{S: solution_id}
	put_item.Item["vote"] = &attributevalue.AttributeValue{S: vote_type}
	put_item.ReturnValues = put.RETVAL_ALL_OLD
	put_item.ConditionExpression = fmt.Sprintf("vote <> :v")
	put_item.ExpressionAttributeValues[":v"] = &attributevalue.AttributeValue{S: vote_type}
	return put_item, nil
}

func GetSolutions(logger *log.Logger, problem_id string, table_name string) ([]*Solution, error) {
	logger.Printf("db_orm_solution.GetSolutions() entry. problem_id: %s, table_name: %s", table_name, problem_id)
	defer logger.Printf("db_orm_solution.GetSolutions() exit.")

	q := query.NewQuery()
	q.TableName = table_name
	q.Select = ep.SELECT_ALL
	q.Limit = 100
	kc := condition.NewCondition()
	kc.AttributeValueList = make([]*attributevalue.AttributeValue, 1)
	kc.AttributeValueList[0] = &attributevalue.AttributeValue{S: problem_id}
	kc.ComparisonOperator = query.OP_EQ
	q.KeyConditions["problem_id"] = kc

	body, code, err := q.EndpointReq()
	if err != nil || code != http.StatusOK {
		logger.Printf("scan failed %d %v %s\n", code, err, body)
		return nil, err
	}
	// Response is the full response from DynamoDB; Item and ConsumedCapacity
	// DyanmoDB does e.g. {"S": "foobar"} for strings etc. in Item.
	var resp query.Response
	um_err := json.Unmarshal([]byte(body), &resp)
	if um_err != nil {
		e := fmt.Sprintf("unmarshal Response: %v", um_err)
		logger.Printf("%s\n", e)
		return make([]*Solution, 0), um_err
	}
	solutions, err := ItemsToSolutions(logger, resp.Items)
	if err != nil {
		logger.Printf("error while converting items to solutions: %s", err)
		return solutions, err
	}
	return solutions, nil
}

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
