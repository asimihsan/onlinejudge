package main

import (
	"encoding/json"
	"log"
	"net/http"

	ep "github.com/smugmug/godynamo/endpoint"
	"github.com/smugmug/godynamo/endpoints/create_table"
	delete_table "github.com/smugmug/godynamo/endpoints/delete_table"
	desc "github.com/smugmug/godynamo/endpoints/describe_table"
	"github.com/smugmug/godynamo/endpoints/list_tables"
	"github.com/smugmug/godynamo/types/attributedefinition"
	"github.com/smugmug/godynamo/types/aws_strings"
	"github.com/smugmug/godynamo/types/keydefinition"
	"github.com/smugmug/godynamo/types/localsecondaryindex"
)

func CreateTables() (err error) {
	if err = createProblemSummaryTable("problem_summary"); err != nil {
		log.Printf("could not create problem_summary table: %s", err)
		return
	}
	if err = createProblemDetailsTable("problem_details"); err != nil {
		log.Printf("could not create problem_details table: %s", err)
		return
	}
	if err = createUnitTestTable("unit_test"); err != nil {
		log.Printf("could not create unit_test table: %s", err)
		return
	}
	return
}

func DeleteTables() (err error) {
	if err = deleteTable("problem_summary"); err != nil {
		log.Printf("could not delete problem summary table: %s", err)
		return
	}
	if err = deleteTable("problem_details"); err != nil {
		log.Printf("could not delete problem details table: %s", err)
		return
	}
	if err = deleteTable("unit_test"); err != nil {
		log.Printf("could not delete unit_test table: %s", err)
		return
	}
	return
}

func deleteTable(table_name string) (err error) {
	log.Printf("deleteTable() entry. table_name: %s", table_name)
	defer log.Printf("deleteTable() exit.")

	var (
		code int
		body []byte
	)

	// DELETE THE TABLE
	del_table1 := delete_table.NewDeleteTable()
	del_table1.TableName = table_name
	body, code, err = del_table1.EndpointReq()
	if err != nil || code != http.StatusOK {
		log.Printf("fail delete %d %v %s\n", code, err, string(body))
		return
	}
	return
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

func createProblemSummaryTable(table_name string) (err error) {
	log.Printf("createProblemSummaryTable() entry.")
	defer log.Printf("createProblemSummaryTable() exit.")

	var (
		code   int
		body   []byte
		exists bool
	)

	if exists, err = doesTableExist(table_name); err != nil {
		log.Printf("unable to check for table existence")
		return
	}
	if exists == true {
		log.Printf("table %s already exists.", table_name)
		return
	}

	create1 := create_table.NewCreateTable()
	create1.TableName = table_name
	create1.ProvisionedThroughput.ReadCapacityUnits = 10
	create1.ProvisionedThroughput.WriteCapacityUnits = 1

	create1.AttributeDefinitions = append(create1.AttributeDefinitions,
		attributedefinition.AttributeDefinition{AttributeName: "id", AttributeType: ep.S})
	create1.AttributeDefinitions = append(create1.AttributeDefinitions,
		attributedefinition.AttributeDefinition{AttributeName: "title", AttributeType: ep.S})
	create1.AttributeDefinitions = append(create1.AttributeDefinitions,
		attributedefinition.AttributeDefinition{AttributeName: "last_updated_date", AttributeType: ep.S})
	create1.KeySchema = append(create1.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "id", KeyType: ep.HASH})
	create1.KeySchema = append(create1.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "title", KeyType: ep.RANGE})

	lsi := localsecondaryindex.NewLocalSecondaryIndex()
	lsi.IndexName = "last_updated_date"
	lsi.Projection.ProjectionType = aws_strings.KEYS_ONLY
	lsi.KeySchema = append(lsi.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "id", KeyType: ep.HASH})
	lsi.KeySchema = append(lsi.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "last_updated_date", KeyType: ep.RANGE})
	create1.LocalSecondaryIndexes = append(create1.LocalSecondaryIndexes, *lsi)

	// Prepare JSON request
	_, create_json_err := json.Marshal(create1)
	if create_json_err != nil {
		log.Printf("%v\n", create_json_err)
		return err
	}

	// Execute JSON request
	body, code, err = create1.EndpointReq()
	if err != nil || code != http.StatusOK {
		log.Printf("create failed %d %v %s\n", code, err, string(body))
		return err
	}

	log.Printf("checking for ACTIVE status for table...")
	_, poll_err := desc.PollTableStatus(table_name, desc.ACTIVE, 100)
	if poll_err != nil {
		log.Printf("poll1:%v\n", poll_err)
		return err
	}

	return
}

func createProblemDetailsTable(table_name string) (err error) {
	log.Printf("createProblemDetailsTable() entry.")
	defer log.Printf("createProblemDetailsTable() exit.")

	var (
		code   int
		body   []byte
		exists bool
	)

	if exists, err = doesTableExist(table_name); err != nil {
		log.Printf("unable to check for table existence")
		return
	}
	if exists == true {
		log.Printf("table %s already exists.", table_name)
		return
	}

	create1 := create_table.NewCreateTable()
	create1.TableName = table_name
	create1.ProvisionedThroughput.ReadCapacityUnits = 10
	create1.ProvisionedThroughput.WriteCapacityUnits = 1

	create1.AttributeDefinitions = append(create1.AttributeDefinitions,
		attributedefinition.AttributeDefinition{AttributeName: "id", AttributeType: ep.S})
	create1.KeySchema = append(create1.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "id", KeyType: ep.HASH})

	// Prepare JSON request
	_, create_json_err := json.Marshal(create1)
	if create_json_err != nil {
		log.Printf("%v\n", create_json_err)
		return err
	}

	// Execute JSON request
	body, code, err = create1.EndpointReq()
	if err != nil || code != http.StatusOK {
		log.Printf("create failed %d %v %s\n", code, err, string(body))
		return err
	}

	log.Printf("checking for ACTIVE status for table...")
	_, poll_err := desc.PollTableStatus(table_name, desc.ACTIVE, 100)
	if poll_err != nil {
		log.Printf("poll1:%v\n", poll_err)
		return err
	}

	return
}

func createUnitTestTable(table_name string) (err error) {
	log.Printf("createUnitTestTable() entry.")
	defer log.Printf("createUnitTestTable() exit.")

	var (
		code   int
		body   []byte
		exists bool
	)

	if exists, err = doesTableExist(table_name); err != nil {
		log.Printf("unable to check for table existence")
		return
	}
	if exists == true {
		log.Printf("table %s already exists.", table_name)
		return
	}

	create1 := create_table.NewCreateTable()
	create1.TableName = table_name
	create1.ProvisionedThroughput.ReadCapacityUnits = 10
	create1.ProvisionedThroughput.WriteCapacityUnits = 1

	create1.AttributeDefinitions = append(create1.AttributeDefinitions,
		attributedefinition.AttributeDefinition{AttributeName: "id", AttributeType: ep.S})
	create1.KeySchema = append(create1.KeySchema,
		keydefinition.KeyDefinition{AttributeName: "id", KeyType: ep.HASH})

	// Prepare JSON request
	_, create_json_err := json.Marshal(create1)
	if create_json_err != nil {
		log.Printf("%v\n", create_json_err)
		return err
	}

	// Execute JSON request
	body, code, err = create1.EndpointReq()
	if err != nil || code != http.StatusOK {
		log.Printf("create failed %d %v %s\n", code, err, string(body))
		return err
	}

	log.Printf("checking for ACTIVE status for table...")
	_, poll_err := desc.PollTableStatus(table_name, desc.ACTIVE, 100)
	if poll_err != nil {
		log.Printf("poll1:%v\n", poll_err)
		return err
	}

	return
}
