package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	logger := log.New(os.Stderr, "", log.Ldate | log.Ltime)
	if len(os.Args) != 3 {
		logger.Println("Usage: preflight-api CERT_FILE KEY_FILE")
		return
	}

	certFile := os.Args[1]
	keyFile := os.Args[2]

	//TODO: cert file, key file, port from config file

	http.HandleFunc("/users/", handleUsers)
	http.HandleFunc("/checklists/", handleChecklists)
	http.HandleFunc("/tokens/", handleTokens)
	http.HandleFunc("/settings/", handleSettings)

	log.Fatal(http.ListenAndServeTLS(":443", certFile, keyFile, nil))
}

//TODO: handler flow:
//	1. instantiate persister, defer close (keep between calls?)
//	2. parse path, branch
//	3. extract token, verify permissions
//	4. call command
//	5. write response

func handleUsers(w http.ResponseWriter, r *http.Request) {
	//TODO: handle:
	//	POST /users - add user
	//	DELETE /users/{id} - delete user
	pathWords := strings.Split(r.URL.Path, "/")[1:]
	if pathWords[len(pathWords)-1] == "" {
		pathWords = pathWords[:len(pathWords)-1]
	}
	//query := r.URL.Query()

	if strings.EqualFold(r.Method, "POST") && len(pathWords) == 1 {
		//TODO
	}
}

func handleChecklists(w http.ResponseWriter, r *http.Request) {
	//TODO: handle:
	//	POST /checklists/{name}/invoke - invoke checklist
	//	POST /checklists - add checklist
	//	DELETE /checklists/{name} - delete checklist
	//	PUT /checklists/{name} - update checklist
	//	GET /checklists/{name} - get checklist
	//	GET /checklists - get all checklists
}

func handleTokens(w http.ResponseWriter, r *http.Request) {
	//TODO: handle:
	//	POST /tokens - add token
	//	DELETE /tokens/{id} - delete token
	//	GET /tokens - get all tokens
}

func handleSettings(w http.ResponseWriter, r *http.Request) {
	//TODO: handle:
	//	PUT /settings/{setting-name} - update setting
	//	GET /settings - get settings
}
