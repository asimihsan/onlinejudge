package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET POST OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "text/plain")
	if r.Method == "OPTIONS" {
		return
	}
	session, _ := getCookieStore(r, "persona-session")
	email := session.Values["email"]
	if email != nil {
		fmt.Fprintf(w, email.(string))
	} else {
		fmt.Fprintf(w, "")
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := getCookieStore(r, "persona-session")
	session.Values["email"] = nil
	session.Save(r, w)
	w.Write([]byte("OK"))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET POST OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		return
	}

	decoder := json.NewDecoder(r.Body)
	var t loginRequest
	err := decoder.Decode(&t)
	if err != nil {
		log.Panicf("Could not decode JSON POST request")
	}

	log.Printf("host: %s, port: %s, assertion: %s", t.Host, t.Port, t.Assertion)

	data := url.Values{
		"assertion": {t.Assertion},
		"audience":  {fmt.Sprintf("%s:%d", t.Host, t.Port)},
	}

	resp, err := http.PostForm("https://verifier.login.persona.org/verify", data)
	if err != nil {
		log.Println(err)
		w.WriteHeader(400)
		w.Write([]byte("Bad Request."))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(400)
		w.Write([]byte("Bad Request."))
	}

	pr := &personaResponse{}
	err = json.Unmarshal(body, pr)
	if err != nil {
		log.Println(err)
		w.WriteHeader(400)
		w.Write([]byte("Bad Request."))
	}

	session, _ := getCookieStore(r, "persona-session")
	session.Values["email"] = pr.Email
	session.Save(r, w)

	w.Write(body)
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
