package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/smugmug/godynamo/endpoints/batch_write_item"
	_ "github.com/smugmug/godynamo/endpoints/get_item"
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

func PutNewSolution(logger *log.Logger, solution *Solution) error {
	logger.Printf("db_orm.PutSolution() entry. solution.SolutionId: %s", solution.SolutionId)
	defer logger.Printf("db_orm.PutSolution() exit.")

	var (
		solution_metadata_table_name = "solution_metadata"
		solution_content_table_name  = "solution_content"
		solution_vote_table_name     = "solution_vote"
	)

	put1, err := getPutSolutionIntoSolutionMetadata(logger, solution)
	if err != nil {
		logger.Printf("failed to get put item for putting solution into %s: %s", solution_metadata_table_name, err)
		return err
	}
	put2, err := getPutSolutionIntoSolutionContent(logger, solution)
	if err != nil {
		logger.Printf("failed to get put item for putting solution into %s: %s", solution_content_table_name, err)
		return err
	}
	put3, err := getPutSolutionIntoSolutionVote(logger, solution)
	if err != nil {
		logger.Printf("failed to get put item for putting solution into %s: %s", solution_vote_table_name, err)
		return err
	}

	b := batch_write_item.NewBatchWriteItem()
	b.RequestItems[solution_metadata_table_name] = make([]batch_write_item.RequestInstance, 0)
	b.RequestItems[solution_content_table_name] = make([]batch_write_item.RequestInstance, 0)
	b.RequestItems[solution_vote_table_name] = make([]batch_write_item.RequestInstance, 0)
	b.RequestItems[solution_metadata_table_name] = append(
		b.RequestItems[solution_metadata_table_name],
		batch_write_item.RequestInstance{PutRequest: &batch_write_item.PutRequest{Item: *put1}})
	b.RequestItems[solution_content_table_name] = append(
		b.RequestItems[solution_content_table_name],
		batch_write_item.RequestInstance{PutRequest: &batch_write_item.PutRequest{Item: *put2}})
	b.RequestItems[solution_vote_table_name] = append(
		b.RequestItems[solution_vote_table_name],
		batch_write_item.RequestInstance{PutRequest: &batch_write_item.PutRequest{Item: *put3}})
	bs, _ := batch_write_item.Split(b)
	for _, bsi := range bs {
		body, code, err := bsi.RetryBatchWrite(0)
		if err != nil || code != http.StatusOK {
			fmt.Printf("error: %v\n%v\n%v\n", string(body), code, err)
			return err
		} else {
			fmt.Printf("worked!: %v\n%v\n%v\n", string(body), code, err)
		}
	}

	return nil
}

func getPutSolutionIntoSolutionMetadata(logger *log.Logger, solution *Solution) (*item.Item, error) {
	log.Printf("db_orm.getPutSolutionIntoSolutionMetadata() entry. solution.SolutionId: %s", solution.SolutionId)
	defer log.Printf("getPutSolutionIntoSolutionMetadata() exit.")

	put_item := item.NewItem()
	put_item["solution_id"] = &attributevalue.AttributeValue{S: solution.SolutionId}
	put_item["problem_id"] = &attributevalue.AttributeValue{S: solution.ProblemId}
	put_item["user_id"] = &attributevalue.AttributeValue{S: solution.UserId}
	put_item["nickname"] = &attributevalue.AttributeValue{S: solution.Nickname}
	put_item["creation_date"] = &attributevalue.AttributeValue{S: solution.CreationDate.Format(time.RFC3339)}
	return &put_item, nil
}

func getPutSolutionIntoSolutionContent(logger *log.Logger, solution *Solution) (*item.Item, error) {
	log.Printf("db_orm.putSolutionIntoSolutionContent() entry. solution.Id: %s", solution.SolutionId)
	defer log.Printf("putSolutionIntoSolutionContent() exit.")

	put_item := item.NewItem()
	put_item["solution_id"] = &attributevalue.AttributeValue{S: solution.SolutionId}
	put_item["problem_id"] = &attributevalue.AttributeValue{S: solution.ProblemId}
	put_item["user_id"] = &attributevalue.AttributeValue{S: solution.UserId}
	compressed_description, err := CompressToBase64(logger, solution.Description)
	if err != nil {
		log.Printf("failed to compress description for solution.SolutionId %s!", solution.SolutionId)
		return &put_item, err
	}
	put_item["description"] = &attributevalue.AttributeValue{B: compressed_description}
	compressed_code, err := CompressToBase64(logger, solution.Code)
	if err != nil {
		log.Printf("failed to compress code for solution.SolutionId %s!", solution.SolutionId)
		return &put_item, err
	}
	put_item["code"] = &attributevalue.AttributeValue{B: compressed_code}
	return &put_item, nil
}

func getPutSolutionIntoSolutionVote(logger *log.Logger, solution *Solution) (*item.Item, error) {
	log.Printf("db_orm.getPutSolutionIntoSolutionVote() entry. solution.SolutionId: %s", solution.SolutionId)
	defer log.Printf("getPutSolutionIntoSolutionVote() exit.")

	put_item := item.NewItem()
	put_item["solution_id"] = &attributevalue.AttributeValue{S: solution.SolutionId}
	put_item["up"] = &attributevalue.AttributeValue{N: "0"}
	put_item["down"] = &attributevalue.AttributeValue{N: "0"}
	return &put_item, nil
}
