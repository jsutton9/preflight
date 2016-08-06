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
	persister, err := persistence.New("localhost", "users")
	if err != nil {
		err = err.Prepend("api.handleUsers: error getting persister: ")
		logger.Println(err.Error())
		return
	}
	defer persister.Close()

	certFile := os.Args[1]
	keyFile := os.Args[2]

	//TODO: cert file, key file, port from config file

	e_handleUsers := encloseHandler(handleUsers, logger, persister)
	e_handleChecklists := encloseHandler(handleChecklists, logger, persister)
	e_handleTokens := encloseHandler(handleTokens, logger, persister)
	e_handleSettings := encloseHandler(handleSettings, logger, persister)

	http.HandleFunc("/users", e_handleUsers)
	http.HandleFunc("/users/", e_handleUsers)
	http.HandleFunc("/checklists", e_handleChecklists)
	http.HandleFunc("/checklists/", e_handleChecklists)
	http.HandleFunc("/tokens", e_handleTokens)
	http.HandleFunc("/tokens/", e_handleTokens)
	http.HandleFunc("/settings", e_handleSettings)
	http.HandleFunc("/settings/", e_handleSettings)

	log.Fatal(http.ListenAndServeTLS(":443", certFile, keyFile, nil))
}

func handleUsers(w http.ResponseWriter, r *http.Request, logger *log.Logger, persister *persistence.Persister) {
	pathWords := getPathWords(r)

	//TODO: verify server token

	if strings.EqualFold(r.Method, "POST") && len(pathWords) == 1 {
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
		w.WriteHeader(404)
	}
}

func handleChecklists(w http.ResponseWriter, r *http.Request, logger *log.Logger, persister *persistence.Persister) {
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
		permissions := security.PermissionFlags{ChecklistWrite: true}
		err = commands.ValidateToken(id, secret, permissions, persister)
		if err != nil {
			err.Prepend("api.handleChecklists: error validating token: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		body, err := readBody(r, 10000)
		if err != nil {
			err = err.Prepend("api.handleChecklists: error reading body: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}
		checklistName, err := commands.AddChecklist(id, body, persister)
		if err != nil {
			err = err.Prepend("api.handleChecklists: error adding checklist: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}
		newUrl := "https://" + r.Host + "/checklists/" + checklistName

		w.Header().Add("Location", newUrl)
		w.WriteHeader(201)
	} else if strings.EqualFold(r.Method, "POST") && len(pathWords) == 3 &&
			strings.EqualFold(pathWords[2], "invoke") {
		permissions := security.PermissionFlags{ChecklistInvoke: true}
		err = commands.ValidateToken(id, secret, permissions, persister)
		if err != nil {
			err.Prepend("api.handleChecklists: error validating token: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		checklistName := pathWords[1]
		err = commands.Invoke(id, checklistName, "", persister) //TODO: app trello key
		if err != nil {
			err = err.Prepend("api.handleChecklists: error invoking checklist: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		w.WriteHeader(204)
	} else if strings.EqualFold(r.Method, "PUT") && len(pathWords) == 2 {
		permissions := security.PermissionFlags{ChecklistWrite: true}
		err = commands.ValidateToken(id, secret, permissions, persister)
		if err != nil {
			err.Prepend("api.handleChecklists: error validating token: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		checklistName := pathWords[1]
		body, err := readBody(r, 10000)
		if err != nil {
			err = err.Prepend("api.handleChecklists: error reading body: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}
		err = commands.UpdateChecklist(id, checklistName, body, persister)
		if err != nil {
			err = err.Prepend("api.handleChecklists: error updating checklist: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}
		w.WriteHeader(204)
	} else if strings.EqualFold(r.Method, "DELETE") && len(pathWords) == 2 {
		permissions := security.PermissionFlags{ChecklistWrite: true}
		err = commands.ValidateToken(id, secret, permissions, persister)
		if err != nil {
			err.Prepend("api.handleChecklists: error validating token: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		checklistName := pathWords[1]
		err = commands.DeleteChecklist(id, checklistName, persister)
		if err != nil {
			err = err.Prepend("api.handleChecklists: error deleting checklist: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}
		w.WriteHeader(204)
	} else {
		w.WriteHeader(404)
	}
}

func handleTokens(w http.ResponseWriter, r *http.Request, logger *log.Logger, persister *persistence.Persister) {
	pathWords := getPathWords(r)

	if strings.EqualFold(r.Method, "GET") && len(pathWords) == 1 {
		//TODO
	} else if strings.EqualFold(r.Method, "POST") && len(pathWords) == 1 {
		username, password, ok := r.BasicAuth()
		if ! ok {
			err := &errors.PreflightError{
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
		w.WriteHeader(404)
	}
}

func handleSettings(w http.ResponseWriter, r *http.Request, logger *log.Logger, persister *persistence.Persister) {
	pathWords := getPathWords(r)
	secret, err := getToken(r)
	if err != nil {
		err = err.Prepend("api.handleSettings: error getting token: ")
		logger.Println(err.Error())
		err.WriteResponse(w)
		return
	}
	id, err := commands.GetUserIdFromToken(secret, persister)
	if err != nil {
		err = err.Prepend("api.handleSettings: error getting id: ")
		logger.Println(err.Error())
		err.WriteResponse(w)
		return
	}

	if strings.EqualFold(r.Method, "GET") && len(pathWords) == 1 {
		permissions := security.PermissionFlags{GeneralRead: true}
		err = commands.ValidateToken(id, secret, permissions, persister)
		if err != nil {
			err.Prepend("api.handleSettings: error validating token: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		settingsString, err := commands.GetGeneralSettings(id, persister)
		if err != nil {
			err = err.Prepend("api.handleSettings: error getting settings: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(settingsString))
	} else if strings.EqualFold(r.Method, "PUT") && len(pathWords) == 2 {
		permissions := security.PermissionFlags{GeneralWrite: true}
		err = commands.ValidateToken(id, secret, permissions, persister)
		if err != nil {
			err.Prepend("api.handleSettings: error validating token: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		settingName := pathWords[1]
		settingValue, err := readBody(r, 10000)
		if err != nil {
			err = err.Prepend("api.handleSettings: error reading body: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		err = commands.SetGeneralSetting(id, settingName, settingValue, persister)
		if err != nil {
			err = err.Prepend("api.handleChecklists: error setting setting: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		w.WriteHeader(204)
	} else {
		w.WriteHeader(404)
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

func encloseHandler(f func(http.ResponseWriter, *http.Request, *log.Logger, *persistence.Persister), logger *log.Logger, persister *persistence.Persister) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		pCopy := persister.Copy()
		defer pCopy.Close()
		f(w, r, logger, pCopy)
	}
}
