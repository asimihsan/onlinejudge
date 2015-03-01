package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	_ "strings"

	"github.com/gorilla/sessions"
)

var (
	store = sessions.NewCookieStore(
		[]byte("5bf9c796032240f1b846ffe751278a7108292c09c96d46468fd6ff2925bd643b"),
	)
)

type personaResponse struct {
	Status   string `json: "status"`
	Email    string `json: "email"`
	Audience string `json: "audience"`
	Expires  int64  `json: "expires"`
	Issuer   string `json: "issuer"`
}

type loginRequest struct {
	Assertion string `json:"assertion"`
	Host      string `json:"host"`
	Port      int64  `json:"port"`
}

func loginCheckHandler(w http.ResponseWriter, r *http.Request) {
	logger = getLogger(getLogPill())
	logger.Printf("handler_login.loginCheckHandler() entry. method: %s", r.Method)
	defer logger.Printf("handler_login.loginCheckHandler() exit.")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET POST OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "text/plain")
	if r.Method == "OPTIONS" {
		return
	}

	response := map[string]interface{}{}
	defer WriteJSONResponse(logger, response, w)
	response["success"] = false

	session, _ := getCookieStore(r, "persona-session")
	user_id := session.Values["user_id"]
	if user_id != nil {
		email := session.Values["email"]
		role := session.Values["role"]
		logger.Printf("user has valid secure cookie set with user_id: %s, email: %s, role: %s",
			user_id.(string), email.(string), role.(string))
		response["email"] = email.(string)
		response["user_id"] = user_id.(string)
		response["role"] = role.(string)
		response["success"] = true
	} else {
		logger.Printf("user does not have valid secure cookie set.")
		w.WriteHeader(401)
		session.Values["email"] = nil
		session.Values["user_id"] = nil
		session.Values["role"] = nil
		session.Save(r, w)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	logger = getLogger(getLogPill())
	logger.Printf("handler_login.logoutHandler() entry. method: %s", r.Method)
	defer logger.Printf("handler_login.logoutHandler() exit.")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET POST OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "text/plain")
	if r.Method == "OPTIONS" {
		return
	}

	response := map[string]interface{}{}
	defer WriteJSONResponse(logger, response, w)
	response["success"] = false

	session, _ := getCookieStore(r, "persona-session")
	session.Values["email"] = nil
	session.Values["user_id"] = nil
	session.Values["role"] = nil
	session.Save(r, w)

	response["success"] = true
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	logger = getLogger(getLogPill())
	logger.Printf("handler_login.loginHandler() entry. method: %s", r.Method)
	defer logger.Printf("handler_login.loginHandler() exit.")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET POST OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		return
	}

	response := map[string]interface{}{}
	defer WriteJSONResponse(logger, response, w)
	response["success"] = false

	decoder := json.NewDecoder(r.Body)
	var t loginRequest
	err := decoder.Decode(&t)
	if err != nil {
		error := fmt.Sprintf("could not decode JSON post request: %s", err)
		response["error"] = error
		logger.Printf(error)
		w.WriteHeader(400)
		return
	}

	logger.Printf("host: %s, port: %s, assertion: %s", t.Host, t.Port, t.Assertion)

	data := url.Values{
		"assertion": {t.Assertion},
		"audience":  {fmt.Sprintf("%s:%d", t.Host, t.Port)},
	}

	resp, err := http.PostForm("https://verifier.login.persona.org/verify", data)
	if err != nil {
		error_msg := fmt.Sprintf("Persona response returned error: %s", err)
		response["error"] = error_msg
		logger.Printf(error_msg)
		w.WriteHeader(500)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		error_msg := fmt.Sprintf("Failed to read body from Persona response: %s", err)
		response["error"] = error_msg
		logger.Println(error_msg)
		w.WriteHeader(500)
		return
	}

	pr := &personaResponse{}
	err = json.Unmarshal(body, pr)
	if err != nil {
		error_msg := fmt.Sprintf("Failed to JSON decode Persona response: %s", err)
		response["error"] = error_msg
		logger.Println(error_msg)
		w.WriteHeader(500)
		return
	}

	logger.Printf("user passes Persona verification with email: %s", pr.Email)

	user, err := GetUserWithEmail(logger, pr.Email)
	if _, ok := err.(UserEmailNotFoundError); ok {
		log.Printf("user with email address %s does not exist, create.", pr.Email)
		user, err = NewUser(logger)
		if err != nil {
			error_msg := fmt.Sprintf("Failed to create new user object: %s", err)
			response["error"] = error_msg
			logger.Println(error_msg)
			w.WriteHeader(500)
			return
		}
		user.Email = pr.Email
		user.Nickname = pr.Email
		log.Printf("new user: %s", user)
		err = PutUser(logger, user, "user", "user_email_to_id", "user_nickname_to_id")
		if err != nil {
			error_msg := fmt.Sprintf("Failed to create new user in backend: %s", err)
			response["error"] = error_msg
			logger.Println(error_msg)
			w.WriteHeader(500)
			return
		}
	}

	session, _ := getCookieStore(r, "persona-session")
	session.Values["email"] = user.Email
	session.Values["user_id"] = user.UserId
	session.Values["role"] = user.Role
	session.Save(r, w)
	response["email"] = user.Email
	response["user_id"] = user.UserId
	response["role"] = user.Role
	response["success"] = true
}

func getCookieStore(r *http.Request, name string) (*sessions.Session, error) {
	session, err := store.Get(r, name)
	session.Options = &sessions.Options{
		Domain:   "runsomecode.com",
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
	return session, err
}
