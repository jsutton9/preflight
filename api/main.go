package main

import (
	"fmt"
	"github.com/jsutton9/preflight/api/errors"
	"github.com/jsutton9/preflight/commands"
	"github.com/jsutton9/preflight/persistence"
	"github.com/jsutton9/preflight/security"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: preflight-api CONFIG_FILE")
		return
	}
	settings, err := persistence.GetServerSettings(os.Args[1])
	if err != nil {
		err.Prepend("api.main: error getting server settings: ")
		fmt.Println(err.Error())
		return
	}

	logger, err := settings.GetLogger()
	if err != nil {
		err.Prepend("api.main: error getting logger: ")
		fmt.Println(err.Error())
		return
	}
	defer logger.Close()
	persister, err := settings.GetPersister()
	if err != nil {
		err.Prepend("api.main: error getting persister: ")
		fmt.Println(err.Error())
		return
	}
	defer persister.Close()

	e_handleUsers := encloseHandler(handleUsers, settings, logger, persister)
	e_handleChecklists := encloseHandler(handleChecklists, settings, logger, persister)
	e_handleTokens := encloseHandler(handleTokens, settings, logger, persister)
	e_handleSettings := encloseHandler(handleSettings, settings, logger, persister)

	http.HandleFunc("/users", e_handleUsers)
	http.HandleFunc("/users/", e_handleUsers)
	http.HandleFunc("/checklists", e_handleChecklists)
	http.HandleFunc("/checklists/", e_handleChecklists)
	http.HandleFunc("/tokens", e_handleTokens)
	http.HandleFunc("/tokens/", e_handleTokens)
	http.HandleFunc("/settings", e_handleSettings)
	http.HandleFunc("/settings/", e_handleSettings)

	portString := ":" + strconv.Itoa(settings.Port)
	log.Fatal(http.ListenAndServeTLS(portString, settings.CertFile, settings.KeyFile, nil))
}

func handleUsers(w http.ResponseWriter, r *http.Request, settings *persistence.ServerSettings, logger *persistence.LoggerCloser, persister *persistence.Persister) {
	pathWords := getPathWords(r)

	_, err := validate(r, security.PermissionFlags{}, true, persister)
	if err != nil {
		err.Prepend("api.handleUsers: error validating request: ")
		logger.Println(err.Error())
		err.WriteResponse(w)
		return
	}

	if strings.EqualFold(r.Method, "POST") && len(pathWords) == 1 {
		body, err := readBody(r, 1000)
		if err != nil {
			err.Prepend("api.handleUsers: error reading body: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}
		id, err := commands.AddUser(body, persister)
		if err != nil {
			err.Prepend("api.handleUsers: error adding user: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}
		w.WriteHeader(201)
		w.Write([]byte(id))
	} else if strings.EqualFold(r.Method, "DELETE") && len(pathWords) == 2 {
		id := pathWords[1]
		err = commands.DeleteUser(id, persister)
		if err != nil {
			err.Prepend("api.handleUsers: error deleting user: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}
		w.WriteHeader(204)
	} else {
		w.WriteHeader(404)
	}
}

func handleChecklists(w http.ResponseWriter, r *http.Request, settings *persistence.ServerSettings, logger *persistence.LoggerCloser, persister *persistence.Persister) {
	pathWords := getPathWords(r)

	if strings.EqualFold(r.Method, "GET") && len(pathWords) == 1 {
		permissions := security.PermissionFlags{ChecklistRead: true}
		id, err := validate(r, permissions, false, persister)
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
		id, err := validate(r, permissions, false, persister)
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
		id, err := validate(r, permissions, false, persister)
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
		id, err := validate(r, permissions, false, persister)
		if err != nil {
			err.Prepend("api.handleChecklists: error validating token: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		checklistName := pathWords[1]
		err = commands.Invoke(id, checklistName, settings.TrelloAppKey, persister)
		if err != nil {
			err = err.Prepend("api.handleChecklists: error invoking checklist: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		w.WriteHeader(204)
	} else if strings.EqualFold(r.Method, "PUT") && len(pathWords) == 2 {
		permissions := security.PermissionFlags{ChecklistWrite: true}
		id, err := validate(r, permissions, false, persister)
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
		id, err := validate(r, permissions, false, persister)
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

func handleTokens(w http.ResponseWriter, r *http.Request, settings *persistence.ServerSettings, logger *persistence.LoggerCloser, persister *persistence.Persister) {
	pathWords := getPathWords(r)

	if strings.EqualFold(r.Method, "GET") && len(pathWords) == 1 {
		permissions := security.PermissionFlags{GeneralRead: true}
		id, err := validate(r, permissions, false, persister)
		if err != nil {
			err.Prepend("api.handleTokens: error validating token: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		tokensString, err := commands.GetTokens(id, persister)
		if err != nil {
			err.Prepend("api.handleTokens: error getting tokens: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		w.WriteHeader(200)
		w.Write([]byte(tokensString))
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
		permissions := security.PermissionFlags{GeneralWrite: true}
		id, err := validate(r, permissions, false, persister)
		if err != nil {
			err = err.Prepend("api.handleTokens: error validating token: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		tokenId := pathWords[1]
		err = commands.DeleteToken(id, tokenId, persister)
		if err != nil {
			err = err.Prepend("api.handleTokens: error deleting token: ")
			logger.Println(err.Error())
			err.WriteResponse(w)
			return
		}

		w.WriteHeader(204)
	} else {
		w.WriteHeader(404)
	}
}

func handleSettings(w http.ResponseWriter, r *http.Request, settings *persistence.ServerSettings, logger *persistence.LoggerCloser, persister *persistence.Persister) {
	pathWords := getPathWords(r)

	if strings.EqualFold(r.Method, "GET") && len(pathWords) == 1 {
		permissions := security.PermissionFlags{GeneralRead: true}
		id, err := validate(r, permissions, false, persister)
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
		id, err := validate(r, permissions, false, persister)
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

		if settingName == "todoist-token" {
			err = commands.SetTodoistToken(id, settingValue, persister)
		} else if settingName == "trello-token" {
			err = commands.SetTrelloToken(id, settingValue, persister)
		} else {
			err = commands.SetGeneralSetting(id, settingName, settingValue, persister)
		}

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

func validate(r *http.Request, permissions security.PermissionFlags, nodeOnly bool, persister *persistence.Persister) (string, *errors.PreflightError) {
	query := r.URL.Query()
	clientToken := query.Get("token")
	nodeSecret := query.Get("nodeSecret")
	userId := query.Get("user")
	if nodeSecret != "" {
		err := commands.ValidateNodeSecret(nodeSecret, persister)
		if err != nil {
			err.Prepend("api.validate: error validating node secret: ")
		}
		return userId, err
	} else if clientToken != "" && !nodeOnly {
		userId, err := commands.GetUserIdFromToken(clientToken, persister)
		if err != nil {
			return "", err.Prepend("api.validate: error getting id: ")
		}
		err = commands.ValidateToken(userId, clientToken, permissions, persister)
		if err != nil {
			err.Prepend("api.validate: error validating token: ")
		}
		return userId, err
	} else {
		return "", &errors.PreflightError{
			Status: 401,
			InternalMessage: "api.validate: no token",
			ExternalMessage: "A security token is required.",
		}
	}
}

func encloseHandler(f func(http.ResponseWriter, *http.Request, *persistence.ServerSettings, *persistence.LoggerCloser, *persistence.Persister), settings *persistence.ServerSettings, logger *persistence.LoggerCloser, persister *persistence.Persister) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		pCopy := persister.Copy()
		defer pCopy.Close()
		f(w, r, settings, logger, pCopy)
	}
}
