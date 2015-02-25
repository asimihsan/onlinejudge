package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	_ "strings"

	"github.com/gorilla/sessions"
)

var (
	secretKey = []byte("acaec567-a14b-449a-8910-1aad305ce6ad")
	store     = sessions.NewCookieStore(secretKey)
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
	defer writeJSONResponse(logger, response, w)
	response["success"] = false

	session, _ := getCookieStore(r, "persona-session")
	email := session.Values["email"]
	if email != nil {
		logger.Printf("user has valid secure cookie set with email: %s", email.(string))
		response["email"] = email.(string)
		response["success"] = true
	} else {
		logger.Printf("user does not have valid secure cookie set.")
		w.WriteHeader(401)
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
	defer writeJSONResponse(logger, response, w)
	response["success"] = false

	session, _ := getCookieStore(r, "persona-session")
	session.Values["email"] = nil
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
	defer writeJSONResponse(logger, response, w)
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
	session.Values["email"] = pr.Email
	session.Save(r, w)
	response["email"] = pr.Email
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

func writeJSONResponse(logger *log.Logger, response map[string]interface{}, w http.ResponseWriter) {
	logger.Println("writeJSONResponse() entry")
	defer logger.Println("writeJSONResponse() exit")
	responseEncoded, _ := json.Marshal(response)
	io.WriteString(w, string(responseEncoded))
}
