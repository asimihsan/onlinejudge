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
		"solution_metadata", "solution_content", "solution_vote"}
)

func CreateTables() (err error) {
	log.Printf("CreateTables() entry.")
	defer log.Printf("CreateTables() exit.")
	if err = createUserTable("user"); err != nil {
		log.Printf("could not create user table: %s", err)
	}
	if err = createUserEmailToIdTable("user_email_to_id"); err != nil {
		log.Printf("could not create user_email_to_id table: %s", err)
	}
	if err = createUserNicknameToIdTable("user_nickname_to_id"); err != nil {
		log.Printf("could not create user_nickname_to_id table: %s", err)
	}
	if err = createSolutionMetadataTable("solution_metadata"); err != nil {
		log.Printf("could not create solution_metadata table: %s", err)
	}
	if err = createSolutionContentTable("solution_content"); err != nil {
		log.Printf("could not create solution_content table: %s", err)
	}
	if err = createSolutionVoteTable("solution_vote"); err != nil {
		log.Printf("could not create user_vote table: %s", err)
	}
	for _, table := range tables {
		log.Printf("checking for ACTIVE status for table %s...", table)
		_, poll_err := desc.PollTableStatus(table, desc.ACTIVE, 100)
		if poll_err != nil {
			log.Printf("failed to poll for table %s to create", poll_err)
			return poll_err
		}
	}
	return
}

func DeleteTables() error {
	log.Printf("DeleteTables() entry.")
	defer log.Printf("DeleteTables() exit.")

	var (
		exists bool
		err    error
	)

	for _, table := range tables {
		if err := deleteTable(table); err != nil {
			log.Printf("could not delete %s table: %s", table, err)
			return err
		}
	}
	for _, table := range tables {
		cnt := 0
		for cnt < 10 {
			if exists, err = doesTableExist(table); err != nil {
				log.Printf("unable to check for table %s existence", table)
				return CannotCheckForTableExistence
			}
			if exists == false {
				log.Printf("table %s no longer exists.", table)
				break
			}
			cnt += 1
			if cnt == 10 {
				log.Printf("timed out waiting for table %s to delete", table)
				return DeleteTableTimeOut
			}
			time.Sleep(3 * time.Second)
		}
	}
	return nil
}

func deleteTable(table_name string) error {
	log.Printf("deleteTable() entry. table_name: %s", table_name)
	defer log.Printf("deleteTable() exit.")

	var (
		code int
		body []byte
	)

	// DELETE THE TABLE
	del_table1 := delete_table.NewDeleteTable()
	del_table1.TableName = table_name
	body, code, err := del_table1.EndpointReq()
	if err != nil || code != http.StatusOK {
		log.Printf("fail delete %d %v %s\n", code, err, string(body))
		return err
	}

	return nil
}

func doesTableExist(table_name string) (exists bool, err error) {
	log.Printf("doesTableExist() entry. table_name: %s", table_name)
	defer log.Printf("doesTableExist() exit.")
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
		log.Printf("list failed %d %v %s\n", code, err, string(body))
		return
	}
	if err = json.Unmarshal(body, &data); err != nil {
		log.Printf("could not decode JSON response (%s): %s", string(body), err)
		return
	}
	if table_names, exists = data["TableNames"]; exists != true {
		log.Printf("Could not find 'TableNames' key in response: %s", data)
		return
	}
	for _, b := range table_names {
		if b == table_name {
			log.Printf("table %s exists, returning true", table_name)
			return true, nil
		}
	}
	log.Printf("table %s does not exist, returning false", table_name)
	return false, nil
}

func executeCreateTable(create1 *create_table.CreateTable) error {
	log.Printf("executeCreateTable() entry. table_name: %s",
		create1.TableName)
	defer log.Printf("executeCreateTable() exit.")

	var (
		code int
		body []byte
	)

	// Prepare JSON request
	_, create_json_err := json.Marshal(create1)
	if create_json_err != nil {
		log.Printf("%v\n", create_json_err)
		return create_json_err
	}

	// Execute JSON request
	body, code, err := create1.EndpointReq()
	if err != nil || code != http.StatusOK {
		log.Printf("create failed %d %v %s\n", code, err, string(body))
		return err
	}

	return nil
}

func createUserTable(table_name string) error {
	log.Printf("createUserTable() entry.")
	defer log.Printf("createUserTable() exit.")

	var (
		exists bool
		err    error
	)

	if exists, err = doesTableExist(table_name); err != nil {
		log.Printf("unable to check for table existence")
		return CannotCheckForTableExistence
	}
	if exists == true {
		log.Printf("table %s already exists.", table_name)
		return TableAlreadyExists
	}

	create1 := create_table.NewCreateTable()
	create1.TableName = table_name
	create1.ProvisionedThroughput.ReadCapacityUnits = 5
	create1.ProvisionedThroughput.WriteCapacityUnits = 1

	create1.AttributeDefinitions = append(create1.AttributeDefinitions,
		attributedefinition.AttributeDefinition{AttributeName: "id", AttributeType: ep.S})
	create1.KeySchema = append(create1.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "id", KeyType: ep.HASH})

	if err := executeCreateTable(create1); err != nil {
		log.Printf("failed to create table: %s", err)
		return err
	}

	return nil
}

func createUserEmailToIdTable(table_name string) error {
	log.Printf("createUserEmailToIdTable() entry.")
	defer log.Printf("createUserEmailToIdTable() exit.")

	var (
		exists bool
		err    error
	)

	if exists, err = doesTableExist(table_name); err != nil {
		log.Printf("unable to check for table existence")
		return CannotCheckForTableExistence
	}
	if exists == true {
		log.Printf("table %s already exists.", table_name)
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

	if err := executeCreateTable(create1); err != nil {
		log.Printf("failed to create table: %s", err)
		return err
	}

	return nil
}

func createUserNicknameToIdTable(table_name string) error {
	log.Printf("createUserNicknameToIdTable() entry.")
	defer log.Printf("createUserNicknameToIdTable() exit.")

	var (
		exists bool
		err    error
	)

	if exists, err = doesTableExist(table_name); err != nil {
		log.Printf("unable to check for table existence")
		return CannotCheckForTableExistence
	}
	if exists == true {
		log.Printf("table %s already exists.", table_name)
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

	if err := executeCreateTable(create1); err != nil {
		log.Printf("failed to create table: %s", err)
		return err
	}

	return nil
}

func createSolutionMetadataTable(table_name string) error {
	log.Printf("createSolutionMetadataTable() entry.")
	defer log.Printf("createSolutionMetadataTable() exit.")

	var (
		exists bool
		err    error
	)

	if exists, err = doesTableExist(table_name); err != nil {
		log.Printf("unable to check for table existence")
		return CannotCheckForTableExistence
	}
	if exists == true {
		log.Printf("table %s already exists.", table_name)
		return TableAlreadyExists
	}

	create1 := create_table.NewCreateTable()
	create1.TableName = table_name
	create1.ProvisionedThroughput.ReadCapacityUnits = 5
	create1.ProvisionedThroughput.WriteCapacityUnits = 1

	create1.AttributeDefinitions = append(create1.AttributeDefinitions,
		attributedefinition.AttributeDefinition{AttributeName: "problem_id", AttributeType: ep.S})
	create1.AttributeDefinitions = append(create1.AttributeDefinitions,
		attributedefinition.AttributeDefinition{AttributeName: "id", AttributeType: ep.S})
	create1.KeySchema = append(create1.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "problem_id", KeyType: ep.HASH})
	create1.KeySchema = append(create1.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "id", KeyType: ep.RANGE})

	if err := executeCreateTable(create1); err != nil {
		log.Printf("failed to create table: %s", err)
		return err
	}

	return nil
}

func createSolutionContentTable(table_name string) error {
	log.Printf("createSolutionContentTable() entry.")
	defer log.Printf("createSolutionContentTable() exit.")

	var (
		exists bool
		err    error
	)

	if exists, err = doesTableExist(table_name); err != nil {
		log.Printf("unable to check for table existence")
		return CannotCheckForTableExistence
	}
	if exists == true {
		log.Printf("table %s already exists.", table_name)
		return TableAlreadyExists
	}

	create1 := create_table.NewCreateTable()
	create1.TableName = table_name
	create1.ProvisionedThroughput.ReadCapacityUnits = 5
	create1.ProvisionedThroughput.WriteCapacityUnits = 1

	create1.AttributeDefinitions = append(create1.AttributeDefinitions,
		attributedefinition.AttributeDefinition{AttributeName: "id", AttributeType: ep.S})
	create1.KeySchema = append(create1.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "id", KeyType: ep.HASH})

	if err := executeCreateTable(create1); err != nil {
		log.Printf("failed to create table: %s", err)
		return err
	}

	return nil
}

func createSolutionVoteTable(table_name string) error {
	log.Printf("createSolutionVoteTable() entry.")
	defer log.Printf("createSolutionVoteTable() exit.")

	var (
		exists bool
		err    error
	)

	if exists, err = doesTableExist(table_name); err != nil {
		log.Printf("unable to check for table existence")
		return CannotCheckForTableExistence
	}
	if exists == true {
		log.Printf("table %s already exists.", table_name)
		return TableAlreadyExists
	}

	create1 := create_table.NewCreateTable()
	create1.TableName = table_name
	create1.ProvisionedThroughput.ReadCapacityUnits = 5
	create1.ProvisionedThroughput.WriteCapacityUnits = 1

	create1.AttributeDefinitions = append(create1.AttributeDefinitions,
		attributedefinition.AttributeDefinition{AttributeName: "solution_id", AttributeType: ep.S})
	create1.AttributeDefinitions = append(create1.AttributeDefinitions,
		attributedefinition.AttributeDefinition{AttributeName: "user_id", AttributeType: ep.S})
	create1.KeySchema = append(create1.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "solution_id", KeyType: ep.HASH})
	create1.KeySchema = append(create1.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "user_id", KeyType: ep.RANGE})

	if err := executeCreateTable(create1); err != nil {
		log.Printf("failed to create table: %s", err)
		return err
	}

	return nil
}
