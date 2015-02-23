package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	get "github.com/smugmug/godynamo/endpoints/get_item"
	put "github.com/smugmug/godynamo/endpoints/put_item"
	"github.com/smugmug/godynamo/types/attributevalue"
)

func GetUserWithId(logger *log.Logger, user_id string) (User, error) {
	logger.Printf("db_orm.GetUserWithId() entry. user_id: %s", user_id)
	defer logger.Printf("db_orm.GetUserWithId() exit.")
	var user User

	get1 := get.NewGetItem()
	get1.TableName = "user"
	get1.Key["id"] = &attributevalue.AttributeValue{
		S: user_id}
	body, code, err := get1.EndpointReq()
	if err != nil || code != http.StatusOK {
		logger.Printf("get failed %d %v %s\n", code, err, body)
		return user, err
	}

	// Response is the full response from DynamoDB; Item and ConsumedCapacity
	// DyanmoDB does e.g. {"S": "foobar"} for strings etc. in Item.
	resp := get.NewResponse()
	um_err := json.Unmarshal([]byte(body), resp)
	if um_err != nil {
		logger.Printf("failed to unmarshal DynamoDB response (%s): %s", resp, um_err)
		return user, um_err
	}

	user, err = ItemToUser(logger, user_id, resp.Item)
	if err != nil {
		logger.Printf("error while converting item to user: %s", err)
		return user, err
	}

	return user, nil
}

func PutUserIntoUser(logger *log.Logger, user *User, table_name string) error {
	log.Printf("db_orm.PutUserIntoUser() entry. user.Id: %s, table_name: %s",
		user.Id, table_name)
	defer log.Printf("PutUserIntoUser() exit.")

	var (
		err error
	)

	put1 := put.NewPutItem()
	put1.TableName = table_name

	put1.Item["id"] = &attributevalue.AttributeValue{
		S: user.Id}
	put1.Item["email"] = &attributevalue.AttributeValue{
		S: user.Email}
	put1.Item["nickname"] = &attributevalue.AttributeValue{
		S: user.Nickname}
	put1.Item["role"] = &attributevalue.AttributeValue{
		S: user.Role}
	put1.Item["creation_date"] = &attributevalue.AttributeValue{
		S: user.CreationDate.Format(time.RFC3339)}
	put1.Item["last_updated_date"] = &attributevalue.AttributeValue{
		S: user.LastUpdatedDate.Format(time.RFC3339)}

	body, code, err := put1.EndpointReq()
	if err != nil || code != http.StatusOK {
		log.Printf("put failed %d %v %s\n", code, err, body)
		return err
	}
	return nil
}
