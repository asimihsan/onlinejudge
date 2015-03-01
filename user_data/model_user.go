package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nu7hatch/gouuid"
	"github.com/smugmug/godynamo/types/item"
)

type UserEmailNotFoundError struct {
	Email string
}

func (e UserEmailNotFoundError) Error() string {
	return fmt.Sprintf("user email '%s' not found", e.Email)
}

type UserIdNotFoundError struct {
	UserId string
}

func (e UserIdNotFoundError) Error() string {
	return fmt.Sprintf("user id '%s' not found", e.UserId)
}

type User struct {
	UserId          string    `json:"user_id"`
	Email           string    `json:"email,omitempty"`
	Nickname        string    `json:"nickname,omitempty"`
	Role            string    `json:"role,omitempty"`
	CreationDate    time.Time `json:"creation_date,omitempty"`
	LastUpdatedDate time.Time `json:"last_updated_date,omitempty"`
}

func (u User) String() string {
	var (
		out []byte
		err error
	)
	if out, err = json.MarshalIndent(u, "", "  "); err != nil {
		log.Printf("could not marshal problem to JSON: %s", err)
		return "<could_not_marshal>"
	}
	return string(out)
}

// Returns a new User object with Id, Role, CreationDate, and LastUpdatedDate
// set up for you. You'll still need to fill in Email and Nickname.
func NewUser(logger *log.Logger) (User, error) {
	var user User
	new_uuid, err := uuid.NewV4()
	if err != nil {
		log.Printf("failed to create new UUID.")
		return user, err
	}
	user = User{
		UserId:          new_uuid.String(),
		Role:            "regular",
		CreationDate:    time.Now(),
		LastUpdatedDate: time.Now(),
	}
	return user, nil
}

// An Item is a returned attributebaluemap from godynamo. This function
// deserializes an Item into a User.
func ItemToUser(logger *log.Logger, input_id string, item item.Item) (User, error) {
	var user User
	user_id, present := item["user_id"]
	if present == false {
		return user, UserIdNotFoundError{input_id}
	}
	user.UserId = user_id.S
	if email, present := item["email"]; present == true {
		user.Email = email.S
	}
	if nickname, present := item["nickname"]; present == true {
		user.Nickname = nickname.S
	}
	if role, present := item["role"]; present == true {
		user.Role = role.S
	}
	if creation_date, present := item["creation_date"]; present == true {
		creation_date_object, err := time.Parse(time.RFC3339, creation_date.S)
		if err != nil {
			logger.Printf("failed to parse creation_date: %s", err)
			return user, err
		}
		user.CreationDate = creation_date_object
	}
	if last_updated_date, present := item["last_updated_date"]; present == true {
		last_updated_date_object, err := time.Parse(time.RFC3339, last_updated_date.S)
		if err != nil {
			logger.Printf("failed to parse last_updated_date: %s", err)
			return user, err
		}
		user.LastUpdatedDate = last_updated_date_object
	}
	return user, nil
}

func ItemsToUsers(logger *log.Logger, items []item.Item) ([]User, error) {
	logger.Printf("model_user.ItemsToUsers() entry.")
	defer logger.Printf("model_user.ItemsToUsers() exit.")
	users := make([]User, 0)
	for _, item := range items {
		user, err := ItemToUser(logger, "", item)
		if err != nil {
			logger.Printf("error while parsing item: %s", err)
			return users, err
		}
		users = append(users, user)
	}
	return users, nil
}
