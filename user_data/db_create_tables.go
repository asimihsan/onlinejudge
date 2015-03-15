package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	ep "github.com/smugmug/godynamo/endpoint"
	"github.com/smugmug/godynamo/endpoints/create_table"
	delete_table "github.com/smugmug/godynamo/endpoints/delete_table"
	desc "github.com/smugmug/godynamo/endpoints/describe_table"
	"github.com/smugmug/godynamo/endpoints/list_tables"
	"github.com/smugmug/godynamo/types/attributedefinition"
	"github.com/smugmug/godynamo/types/keydefinition"
)

var (
	CannotCheckForTableExistence = errors.New("Cannot check for table existence.")
	TableAlreadyExists           = errors.New("Table already exists.")
	DeleteTableTimeOut           = errors.New("Delete table time out.")
	tables                       = []string{
		"user", "user_email_to_id", "user_nickname_to_id",
		"solution", "user_vote"}
)

func CreateTables(logger *log.Logger) error {
	logger.Printf("CreateTables() entry.")
	defer logger.Printf("CreateTables() exit.")
	if err := createUserTable(logger, "user"); err != nil {
		logger.Printf("could not create user table: %s", err)
	}
	if err := createUserEmailToIdTable(logger, "user_email_to_id"); err != nil {
		logger.Printf("could not create user_email_to_id table: %s", err)
	}
	if err := createUserNicknameToIdTable(logger, "user_nickname_to_id"); err != nil {
		logger.Printf("could not create user_nickname_to_id table: %s", err)
	}
	if err := createSolutionTable(logger, "solution"); err != nil {
		logger.Printf("could not create solution table: %s", err)
	}
	if err := createUserVoteTable(logger, "user_vote"); err != nil {
		logger.Printf("could not create user_vote table: %s", err)
	}
	time.Sleep(5 * time.Second)
	for _, table := range tables {
		logger.Printf("checking for ACTIVE status for table %s...", table)
		_, poll_err := desc.PollTableStatus(table, desc.ACTIVE, 100)
		if poll_err != nil {
			logger.Printf("failed to poll for table %s to create", poll_err)
			return poll_err
		}
	}
	return nil
}

func DeleteTables(logger *log.Logger) error {
	logger.Printf("DeleteTables() entry.")
	defer logger.Printf("DeleteTables() exit.")

	var (
		exists bool
		err    error
	)

	for _, table := range tables {
		if err := deleteTable(logger, table); err != nil {
			logger.Printf("could not delete %s table: %s", table, err)
			return err
		}
	}
	time.Sleep(5 * time.Second)
	for _, table := range tables {
		cnt := 0
		for cnt < 10 {
			if exists, err = doesTableExist(logger, table); err != nil {
				logger.Printf("unable to check for table %s existence", table)
				return CannotCheckForTableExistence
			}
			if exists == false {
				logger.Printf("table %s no longer exists.", table)
				break
			}
			cnt += 1
			if cnt == 20 {
				logger.Printf("timed out waiting for table %s to delete", table)
				return DeleteTableTimeOut
			}
			time.Sleep(3 * time.Second)
		}
	}
	return nil
}

func deleteTable(logger *log.Logger, table_name string) error {
	logger.Printf("deleteTable() entry. table_name: %s", table_name)
	defer logger.Printf("deleteTable() exit.")

	var (
		code int
		body []byte
	)

	// DELETE THE TABLE
	del_table1 := delete_table.NewDeleteTable()
	del_table1.TableName = table_name
	body, code, err := del_table1.EndpointReq()
	if err != nil || code != http.StatusOK {
		logger.Printf("fail delete %d %v %s\n", code, err, string(body))
		return err
	}

	return nil
}

func doesTableExist(logger *log.Logger, table_name string) (exists bool, err error) {
	logger.Printf("doesTableExist() entry. table_name: %s", table_name)
	defer logger.Printf("doesTableExist() exit.")
	var (
		code        int
		body        []byte
		l           list_tables.List
		data        map[string][]string
		table_names []string
	)

	l.ExclusiveStartTableName = ""
	l.Limit = 100
	body, code, err = l.EndpointReq()
	if err != nil || code != http.StatusOK {
		logger.Printf("list failed %d %v %s\n", code, err, string(body))
		return
	}
	if err = json.Unmarshal(body, &data); err != nil {
		logger.Printf("could not decode JSON response (%s): %s", string(body), err)
		return
	}
	if table_names, exists = data["TableNames"]; exists != true {
		logger.Printf("Could not find 'TableNames' key in response: %s", data)
		return
	}
	for _, b := range table_names {
		if b == table_name {
			logger.Printf("table %s exists, returning true", table_name)
			return true, nil
		}
	}
	logger.Printf("table %s does not exist, returning false", table_name)
	return false, nil
}

func executeCreateTable(logger *log.Logger, create1 *create_table.CreateTable) error {
	logger.Printf("executeCreateTable() entry. table_name: %s",
		create1.TableName)
	defer logger.Printf("executeCreateTable() exit.")

	var (
		code int
		body []byte
	)

	// Prepare JSON request
	_, create_json_err := json.Marshal(create1)
	if create_json_err != nil {
		logger.Printf("%v\n", create_json_err)
		return create_json_err
	}

	// Execute JSON request
	body, code, err := create1.EndpointReq()
	if err != nil || code != http.StatusOK {
		logger.Printf("create failed %d %v %s\n", code, err, string(body))
		return err
	}

	return nil
}

func createUserTable(logger *log.Logger, table_name string) error {
	logger.Printf("createUserTable() entry.")
	defer logger.Printf("createUserTable() exit.")

	var (
		exists bool
		err    error
	)

	if exists, err = doesTableExist(logger, table_name); err != nil {
		logger.Printf("unable to check for table existence")
		return CannotCheckForTableExistence
	}
	if exists == true {
		logger.Printf("table %s already exists.", table_name)
		return TableAlreadyExists
	}

	create1 := create_table.NewCreateTable()
	create1.TableName = table_name
	create1.ProvisionedThroughput.ReadCapacityUnits = 5
	create1.ProvisionedThroughput.WriteCapacityUnits = 1

	create1.AttributeDefinitions = append(create1.AttributeDefinitions,
		attributedefinition.AttributeDefinition{AttributeName: "user_id", AttributeType: ep.S})
	create1.KeySchema = append(create1.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "user_id", KeyType: ep.HASH})

	if err := executeCreateTable(logger, create1); err != nil {
		logger.Printf("failed to create table: %s", err)
		return err
	}

	return nil
}

func createUserEmailToIdTable(logger *log.Logger, table_name string) error {
	logger.Printf("createUserEmailToIdTable() entry.")
	defer logger.Printf("createUserEmailToIdTable() exit.")

	var (
		exists bool
		err    error
	)

	if exists, err = doesTableExist(logger, table_name); err != nil {
		logger.Printf("unable to check for table existence")
		return CannotCheckForTableExistence
	}
	if exists == true {
		logger.Printf("table %s already exists.", table_name)
		return TableAlreadyExists
	}

	create1 := create_table.NewCreateTable()
	create1.TableName = table_name
	create1.ProvisionedThroughput.ReadCapacityUnits = 5
	create1.ProvisionedThroughput.WriteCapacityUnits = 1

	create1.AttributeDefinitions = append(create1.AttributeDefinitions,
		attributedefinition.AttributeDefinition{AttributeName: "email", AttributeType: ep.S})
	create1.KeySchema = append(create1.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "email", KeyType: ep.HASH})

	if err := executeCreateTable(logger, create1); err != nil {
		logger.Printf("failed to create table: %s", err)
		return err
	}

	return nil
}

func createUserNicknameToIdTable(logger *log.Logger, table_name string) error {
	logger.Printf("createUserNicknameToIdTable() entry.")
	defer logger.Printf("createUserNicknameToIdTable() exit.")

	var (
		exists bool
		err    error
	)

	if exists, err = doesTableExist(logger, table_name); err != nil {
		logger.Printf("unable to check for table existence")
		return CannotCheckForTableExistence
	}
	if exists == true {
		logger.Printf("table %s already exists.", table_name)
		return TableAlreadyExists
	}

	create1 := create_table.NewCreateTable()
	create1.TableName = table_name
	create1.ProvisionedThroughput.ReadCapacityUnits = 5
	create1.ProvisionedThroughput.WriteCapacityUnits = 1

	create1.AttributeDefinitions = append(create1.AttributeDefinitions,
		attributedefinition.AttributeDefinition{AttributeName: "nickname", AttributeType: ep.S})
	create1.KeySchema = append(create1.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "nickname", KeyType: ep.HASH})

	if err := executeCreateTable(logger, create1); err != nil {
		logger.Printf("failed to create table: %s", err)
		return err
	}

	return nil
}

func createSolutionTable(logger *log.Logger, table_name string) error {
	logger.Printf("createSolutionTable() entry.")
	defer logger.Printf("createSolutionTable() exit.")

	var (
		exists bool
		err    error
	)

	if exists, err = doesTableExist(logger, table_name); err != nil {
		logger.Printf("unable to check for table existence")
		return CannotCheckForTableExistence
	}
	if exists == true {
		logger.Printf("table %s already exists.", table_name)
		return TableAlreadyExists
	}

	create1 := create_table.NewCreateTable()
	create1.TableName = table_name
	create1.ProvisionedThroughput.ReadCapacityUnits = 5
	create1.ProvisionedThroughput.WriteCapacityUnits = 1

	create1.AttributeDefinitions = append(create1.AttributeDefinitions,
		attributedefinition.AttributeDefinition{AttributeName: "problem_id", AttributeType: ep.S})
	create1.AttributeDefinitions = append(create1.AttributeDefinitions,
		attributedefinition.AttributeDefinition{AttributeName: "user_id", AttributeType: ep.S})
	create1.KeySchema = append(create1.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "problem_id", KeyType: ep.HASH})
	create1.KeySchema = append(create1.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "user_id", KeyType: ep.RANGE})

	if err := executeCreateTable(logger, create1); err != nil {
		logger.Printf("failed to create table: %s", err)
		return err
	}

	return nil
}

func createUserVoteTable(logger *log.Logger, table_name string) error {
	logger.Printf("createUserVoteTable() entry.")
	defer logger.Printf("createUserVoteTable() exit.")

	var (
		exists bool
		err    error
	)

	if exists, err = doesTableExist(logger, table_name); err != nil {
		logger.Printf("unable to check for table existence")
		return CannotCheckForTableExistence
	}
	if exists == true {
		logger.Printf("table %s already exists.", table_name)
		return TableAlreadyExists
	}

	create1 := create_table.NewCreateTable()
	create1.TableName = table_name
	create1.ProvisionedThroughput.ReadCapacityUnits = 5
	create1.ProvisionedThroughput.WriteCapacityUnits = 1

	create1.AttributeDefinitions = append(create1.AttributeDefinitions,
		attributedefinition.AttributeDefinition{AttributeName: "user_vote_id", AttributeType: ep.S})
	create1.AttributeDefinitions = append(create1.AttributeDefinitions,
		attributedefinition.AttributeDefinition{AttributeName: "solution_id", AttributeType: ep.S})
	create1.KeySchema = append(create1.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "user_vote_id", KeyType: ep.HASH})
	create1.KeySchema = append(create1.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "solution_id", KeyType: ep.RANGE})

	if err := executeCreateTable(logger, create1); err != nil {
		logger.Printf("failed to create table: %s", err)
		return err
	}

	return nil
}
