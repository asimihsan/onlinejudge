package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
	logger = GetLogger(GetLogPill())
	logger.Printf("handler_login.loginCheckHandler() entry. method: %s", r.Method)
	defer logger.Printf("handler_login.loginCheckHandler() exit.")

	SetCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	response := map[string]interface{}{}
	defer WriteJSONResponse(logger, response, w)
	response["success"] = false

	session, _ := GetCookieStore(r, "persona-session")
	user_id := session.Values["user_id"]
	if user_id != nil {
		email := session.Values["email"]
		nickname := session.Values["nickname"]
		role := session.Values["role"]
		logger.Printf("user has valid secure cookie set with user_id: %s, email: %s, role: %s",
			user_id.(string), email.(string), role.(string))
		response["email"] = email.(string)
		response["nickname"] = nickname.(string)
		response["user_id"] = user_id.(string)
		response["role"] = role.(string)
		response["success"] = true
	} else {
		logger.Printf("user does not have valid secure cookie set.")
		w.WriteHeader(401)
		session.Values["email"] = nil
		session.Values["nickname"] = nil
		session.Values["user_id"] = nil
		session.Values["role"] = nil
		session.Save(r, w)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	logger = GetLogger(GetLogPill())
	logger.Printf("handler_login.logoutHandler() entry. method: %s", r.Method)
	defer logger.Printf("handler_login.logoutHandler() exit.")

	SetCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	response := map[string]interface{}{}
	defer WriteJSONResponse(logger, response, w)
	response["success"] = false

	session, _ := GetCookieStore(r, "persona-session")
	session.Values["email"] = nil
	session.Values["nickname"] = nil
	session.Values["user_id"] = nil
	session.Values["role"] = nil
	session.Save(r, w)

	response["success"] = true
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	logger = GetLogger(GetLogPill())
	logger.Printf("handler_login.loginHandler() entry. method: %s", r.Method)
	defer logger.Printf("handler_login.loginHandler() exit.")

	SetCORS(w)
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

	session, _ := GetCookieStore(r, "persona-session")
	session.Values["email"] = user.Email
	session.Values["nickname"] = user.Nickname
	session.Values["user_id"] = user.UserId
	session.Values["role"] = user.Role
	session.Save(r, w)
	response["email"] = user.Email
	response["user_id"] = user.UserId
	response["nickname"] = user.Nickname
	response["role"] = user.Role
	response["success"] = true
}
