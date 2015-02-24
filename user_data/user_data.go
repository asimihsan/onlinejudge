package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
)

var (
	logger  = getLogger("logger")
	letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
)

func getLogger(prefix string) *log.Logger {
	paddedPrefix := fmt.Sprintf("%-8s: ", prefix)
	return log.New(os.Stdout, paddedPrefix,
		log.Ldate|log.Ltime|log.Lmicroseconds)
}

func getLogPill() string {
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	Initialize()
	//DeleteTables()
	//CreateTables()

	/*
		user, err := NewUser(logger)
		if err != nil {
			log.Printf("failed to new user")
		}
		user.Email = "user@host.com"
		user.Nickname = "user"
		PutUser(logger, user, "user", "user_email_to_id", "user_nickname_to_id")
	*/

	/*
		user, err := GetUserWithEmail(logger, "user@host.com")
		if err != nil {
			log.Printf("failed to get user: %s", err)
		}
		log.Printf("%s", user)

		user, err = GetUserWithNickname(logger, "user")
		if err != nil {
			log.Printf("failed to get user: %s", err)
		}
		log.Printf("%s", user)
	*/

	http.HandleFunc("/auth/check", loginCheckHandler)
	http.HandleFunc("/auth/login", loginHandler)
	http.HandleFunc("/auth/logout", logoutHandler)

	log.Printf("Starting HTTP server...")
	log.Fatal(http.ListenAndServe("localhost:9001", nil))
}
