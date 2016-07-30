package main

import (
	"fmt"
	"github.com/jsutton9/preflight/api/errors"
	"github.com/jsutton9/preflight/commands"
	"github.com/jsutton9/preflight/persistence"
	"github.com/jsutton9/preflight/security"
	//"html"
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

	http.HandleFunc("/users", handleUsers)
	http.HandleFunc("/users/", handleUsers)
	http.HandleFunc("/checklists", handleChecklists)
	http.HandleFunc("/checklists/", handleChecklists)
	http.HandleFunc("/tokens", handleTokens)
	http.HandleFunc("/tokens/", handleTokens)
	http.HandleFunc("/settings", handleSettings)
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
	logger := log.New(os.Stderr, "", log.Ldate | log.Ltime)
	persister, err := persistence.New("localhost", "users")
	if err != nil {
		err = err.Prepend("api.handleUsers: error getting persister: ")
		logger.Println(err.Error())
		err.WriteResponse(w)
		return
	}
	defer persister.Close()

	pathWords := getPathWords(r)
	//query := r.URL.Query()

	if strings.EqualFold(r.Method, "POST") && len(pathWords) == 1 {
		//TODO: verify server token
		body, pErr := readBody(r, 1000)
		if pErr != nil {
			logger.Println(pErr.Error())
			pErr.WriteResponse(w)
			return
		}
		id, pErr := commands.AddUser(body, persister)
		if pErr != nil {
			logger.Println(pErr.Error())
			pErr.WriteResponse(w)
			return
		}
		w.WriteHeader(201)
		w.Write([]byte(id))
	} else if strings.EqualFold(r.Method, "DELETE") && len(pathWords) == 2 {
		//id := pathWords[1]
		//TODO
	} else {
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
	logger := log.New(os.Stderr, "", log.Ldate | log.Ltime)
	persister, err := persistence.New("localhost", "users")
	if err != nil {
		err = err.Prepend("api.handleChecklists: error getting persister: ")
		logger.Println(err.Error())
		err.WriteResponse(w)
		return
	}
	defer persister.Close()

	pathWords := getPathWords(r)
	secret, err := getToken(r)
	if err != nil {
		err = err.Prepend("api.handleChecklists: error getting token: ")
		logger.Println(err.Error())
		err.WriteResponse(w)
		return
	}
	id, err := commands.GetUserIdFromToken(secret, persister)
	if err != nil {
		err = err.Prepend("api.handleChecklists: error getting id: ")
		logger.Println(err.Error())
		err.WriteResponse(w)
		return
	}

	if strings.EqualFold(r.Method, "GET") && len(pathWords) == 1 {
		permissions := security.PermissionFlags{ChecklistRead: true}
		err = commands.ValidateToken(id, secret, permissions, persister)
		if err != nil {
			err.Prepend("api.handleChecklists: error validating token: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		checklistsString, err := commands.GetChecklistsString(id, persister)
		if err != nil {
			err = err.Prepend("api.handleChecklists: error getting checklists: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(checklistsString))
	} else if strings.EqualFold(r.Method, "GET") && len(pathWords) == 2 {
		permissions := security.PermissionFlags{ChecklistRead: true}
		err = commands.ValidateToken(id, secret, permissions, persister)
		if err != nil {
			err.Prepend("api.handleChecklists: error validating token: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		checklistName := pathWords[1]
		checklistString, err := commands.GetChecklistString(id, checklistName, persister)
		if err != nil {
			err = err.Prepend("api.handleChecklists: error getting checklist: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(checklistString))
	} else if strings.EqualFold(r.Method, "POST") && len(pathWords) == 1 {
		//TODO
	} else if strings.EqualFold(r.Method, "POST") && len(pathWords) == 3 {
		//TODO
	} else if strings.EqualFold(r.Method, "PUT") && len(pathWords) == 2 {
		//TODO
	} else if strings.EqualFold(r.Method, "DELETE") && len(pathWords) == 2 {
		//TODO
	} else {
		//TODO
	}
}

func handleTokens(w http.ResponseWriter, r *http.Request) {
	//TODO: handle:
	//	POST /tokens - add token
	//	DELETE /tokens/{id} - delete token
	//	GET /tokens - get all tokens
	logger := log.New(os.Stderr, "", log.Ldate | log.Ltime)
	persister, err := persistence.New("localhost", "users")
	if err != nil {
		err = err.Prepend("api.handleTokens: error getting persister: ")
		logger.Println(err.Error())
		err.WriteResponse(w)
		return
	}
	defer persister.Close()

	pathWords := getPathWords(r)

	if strings.EqualFold(r.Method, "GET") && len(pathWords) == 1 {
		//TODO
	} else if strings.EqualFold(r.Method, "POST") && len(pathWords) == 1 {
		username, password, ok := r.BasicAuth()
		if ! ok {
			err = &errors.PreflightError{
				Status: 401,
				InternalMessage: "api.handleTokens: no basic auth header",
				ExternalMessage: "Basic authentication is required to add a token.",
			}
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}
		id, err := commands.GetUserIdFromEmail(username, persister)
		if err != nil {
			err = err.Prepend("api.handleTokens: error getting user: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}
		err = commands.ValidatePassword(id, password, persister)
		if err != nil {
			err = err.Prepend("api.handleTokens: error validating password: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		body, err := readBody(r, 10000)
		if err != nil {
			err = err.Prepend("api.handleTokens: error reading body: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}
		tokenString, err := commands.AddToken(id, body, persister)
		if err != nil {
			err = err.Prepend("api.handleTokens: error adding token: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		w.WriteHeader(201)
		w.Write([]byte(tokenString))
	} else if strings.EqualFold(r.Method, "DELETE") && len(pathWords) == 2 {
		//TODO
	} else {
		//TODO
	}
}

func handleSettings(w http.ResponseWriter, r *http.Request) {
	//TODO: handle:
	//	PUT /settings/{setting-name} - update setting
	//	GET /settings - get settings
	logger := log.New(os.Stderr, "", log.Ldate | log.Ltime)
	persister, err := persistence.New("localhost", "users")
	if err != nil {
		err = err.Prepend("api.handleSettings: error getting persister: ")
		logger.Println(err.Error())
		err.WriteResponse(w)
		return
	}
	defer persister.Close()

	pathWords := getPathWords(r)
	//query := r.URL.Query()

	if strings.EqualFold(r.Method, "GET") && len(pathWords) == 1 {
	} else if strings.EqualFold(r.Method, "PUT") && len(pathWords) == 2 {
	} else {
	}
}

func readBody(r *http.Request, limit int) (string, *errors.PreflightError) {
	bodyBytes := make([]byte, limit)
	n, err := r.Body.Read(bodyBytes)
	if err == nil {
		pErr := &errors.PreflightError{
			Status: 413,
			InternalMessage: fmt.Sprintf("api.readBody: " +
				"request body exceeded size limit of %d", limit),
			ExternalMessage: fmt.Sprintf("The request body was too large. " +
				"This type of request is limited to %d bytes", limit),
		}
		return "", pErr
	} else if err.Error() != "EOF" {
		pErr := &errors.PreflightError{
			Status: 500,
			InternalMessage: "api.readBody: error reading request body: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error reading the request body.",
		}
		return "", pErr
	}

	return string(bodyBytes[:n]), nil
}

func getPathWords(r *http.Request) []string {
	pathWords := strings.Split(r.URL.Path, "/")[1:]
	if pathWords[len(pathWords)-1] == "" {
		pathWords = pathWords[:len(pathWords)-1]
	}

	return pathWords
}

func getToken(r *http.Request) (string, *errors.PreflightError) {
	query := r.URL.Query()
	secret := query.Get("token")
	if secret == "" {
		err := &errors.PreflightError{
			Status: 401,
			InternalMessage: "api.getUser: no token",
			ExternalMessage: "A security token is required.",
		}
		return "", err
	}

	return secret, nil
}
