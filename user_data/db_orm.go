package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	batch_write_item "github.com/smugmug/godynamo/endpoints/batch_write_item"
	get "github.com/smugmug/godynamo/endpoints/get_item"
	_ "github.com/smugmug/godynamo/endpoints/put_item"
	"github.com/smugmug/godynamo/types/attributevalue"
	"github.com/smugmug/godynamo/types/item"
)

type UserWithEmailAlreadyExists struct {
	Email string
}

func (e UserWithEmailAlreadyExists) Error() string {
	return fmt.Sprintf("user with email '%s' already exists", e.Email)
}

func executeGetItem(logger *log.Logger, get_request *get.GetItem) (*get.Response, error) {
	body, code, err := get_request.EndpointReq()
	if err != nil || code != http.StatusOK {
		logger.Printf("get failed %d %v %s\n", code, err, body)
		return nil, err
	}

	// Response is the full response from DynamoDB; Item and ConsumedCapacity
	// DyanmoDB does e.g. {"S": "foobar"} for strings etc. in Item.
	resp := get.NewResponse()
	um_err := json.Unmarshal([]byte(body), resp)
	if um_err != nil {
		logger.Printf("failed to unmarshal DynamoDB response (%s): %s", resp, um_err)
		return nil, um_err
	}

	return resp, nil
}

func GetUserWithId(logger *log.Logger, user_id string) (User, error) {
	logger.Printf("db_orm.GetUserWithId() entry. user_id: %s", user_id)
	defer logger.Printf("db_orm.GetUserWithId() exit.")

	var user User

	get1 := get.NewGetItem()
	get1.TableName = "user"
	get1.Key["id"] = &attributevalue.AttributeValue{
		S: user_id}

	resp, err := executeGetItem(logger, get1)
	if err != nil {
		log.Printf("failed to execute get item: %s", err)
		return user, err
	}

	user, err = ItemToUser(logger, user_id, resp.Item)
	if err != nil {
		logger.Printf("error while converting item to user: %s", err)
		return user, err
	}

	return user, nil
}

func GetUserWithEmail(logger *log.Logger, user_email string) (User, error) {
	logger.Printf("db_orm.GetUserWithEmail() entry. user_email: %s", user_email)
	defer logger.Printf("db_orm.GetUserWithEmail() exit.")

	var user User

	get1 := get.NewGetItem()
	get1.TableName = "user_email_to_id"
	get1.Key["email"] = &attributevalue.AttributeValue{
		S: user_email}

	resp, err := executeGetItem(logger, get1)
	if err != nil {
		log.Printf("failed to execute get item: %s", err)
		return user, err
	}

	user_id, present := resp.Item["id"]
	if present == false {
		log.Printf("could not find user with user_email: %s", user_email)
		return user, UserEmailNotFoundError{user_email}
	}

	user, err = GetUserWithId(logger, user_id.S)
	if err != nil {
		log.Printf("Failed to get user with ID %s: %s", user_id, err)
		return user, err
	}

	return user, nil
}

func GetUserWithNickname(logger *log.Logger, user_nickname string) (User, error) {
	logger.Printf("db_orm.GetUserWithNickname() entry. user_nickname: %s", user_nickname)
	defer logger.Printf("db_orm.GetUserWithNickname() exit.")

	var user User

	get1 := get.NewGetItem()
	get1.TableName = "user_nickname_to_id"
	get1.Key["nickname"] = &attributevalue.AttributeValue{
		S: user_nickname}

	resp, err := executeGetItem(logger, get1)
	if err != nil {
		log.Printf("failed to execute get item: %s", err)
		return user, err
	}

	user_id, present := resp.Item["id"]
	if present == false {
		log.Printf("could not find user with user_nickname: %s", user_nickname)
		return user, UserEmailNotFoundError{user_nickname}
	}

	user, err = GetUserWithId(logger, user_id.S)
	if err != nil {
		log.Printf("Failed to get user with ID %s: %s", user_id, err)
		return user, err
	}

	return user, nil
}

func PutUser(logger *log.Logger, user User,
	user_table_name string, user_email_to_id_table_name string,
	user_nickname_to_id_table_name string) error {
	log.Printf("db_orm.PutUser() entry. user.Id: %s, user_table_name: %s, "+
		"user_email_to_id_table_name: %s, user_nickname_to_id_table_name: %s",
		user.Id, user_table_name, user_email_to_id_table_name,
		user_nickname_to_id_table_name)
	defer log.Printf("PutUser() exit.")

	var (
		p1 batch_write_item.PutRequest
		p2 batch_write_item.PutRequest
		p3 batch_write_item.PutRequest
	)

	_, err := GetUserWithEmail(logger, user.Email)
	if _, ok := err.(UserEmailNotFoundError); !ok {
		log.Printf("user with email address %s already exists: %s", user.Email, err)
		return UserWithEmailAlreadyExists{user.Email}
	}

	b := batch_write_item.NewBatchWriteItem()
	b.RequestItems[user_table_name] = make([]batch_write_item.RequestInstance, 0)
	b.RequestItems[user_email_to_id_table_name] = make([]batch_write_item.RequestInstance, 0)
	b.RequestItems[user_nickname_to_id_table_name] = make([]batch_write_item.RequestInstance, 0)

	logger.Printf("user to create in DB: %s", user)

	p1.Item = item.NewItem()
	p1.Item["id"] = &attributevalue.AttributeValue{S: user.Id}
	p1.Item["email"] = &attributevalue.AttributeValue{S: user.Email}
	p1.Item["nickname"] = &attributevalue.AttributeValue{S: user.Nickname}
	p1.Item["role"] = &attributevalue.AttributeValue{S: user.Role}
	p1.Item["creation_date"] = &attributevalue.AttributeValue{S: user.CreationDate.Format(time.RFC3339)}
	p1.Item["last_updated_date"] = &attributevalue.AttributeValue{S: user.LastUpdatedDate.Format(time.RFC3339)}
	b.RequestItems[user_table_name] = append(
		b.RequestItems[user_table_name],
		batch_write_item.RequestInstance{PutRequest: &p1})

	p2.Item = item.NewItem()
	p2.Item["email"] = &attributevalue.AttributeValue{S: user.Email}
	p2.Item["id"] = &attributevalue.AttributeValue{S: user.Id}
	b.RequestItems[user_email_to_id_table_name] = append(
		b.RequestItems[user_email_to_id_table_name],
		batch_write_item.RequestInstance{PutRequest: &p2})

	p3.Item = item.NewItem()
	p3.Item["nickname"] = &attributevalue.AttributeValue{S: user.Nickname}
	p3.Item["id"] = &attributevalue.AttributeValue{S: user.Id}
	b.RequestItems[user_nickname_to_id_table_name] = append(
		b.RequestItems[user_nickname_to_id_table_name],
		batch_write_item.RequestInstance{PutRequest: &p3})

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
