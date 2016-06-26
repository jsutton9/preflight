package main

import (
	"fmt"
	"github.com/jsutton9/preflight/commands"
	"github.com/jsutton9/preflight/persistence"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	usage := "Usage:\n"
	usage += "\tpreflight add-user EMAIL PASSWORD\n"
	usage += "\tpreflight update EMAIL TRELLO_KEY\n"
	usage += "\tpreflight invoke EMAIL CHECKLIST_NAME TRELLO_KEY\n"
	usage += "\tpreflight get-checklists EMAIL\n"
	usage += "\tpreflight add-checklist EMAIL CHECKLIST_NAME CHECKLIST_FILE\n"
	usage += "\tpreflight update-checklist EMAIL CHECKLIST_NAME CHECKLIST_FILE\n"
	usage += "\tpreflight delete-checklist EMAIL CHECKLIST_NAME\n"
	usage += "\tpreflight set-todoist-token EMAIL TOKEN\n"
	usage += "\tpreflight set-trello-token EMAIL TOKEN\n"
	usage += "\tpreflight get-general-settings EMAIL\n"
	usage += "\tpreflight set-general-setting EMAIL SETTING VALUE\n"

	logger := log.New(os.Stderr, "", log.Ldate | log.Ltime)
	if len(os.Args) < 2 {
		logger.Println(usage)
	} else if os.Args[1] == "add-user" {
		if len(os.Args) != 4 {
			logger.Println(usage)
			return
		}
		persister, err := persistence.New("localhost", "users")
		if err != nil {
			logger.Println("main: error getting persister: " +
				"\n\t" + err.Error())
			return
		}
		email := os.Args[2]
		password := os.Args[3]
		userReq := fmt.Sprintf("{\"email\":\"%s\",\"password\":\"%s\"}", email, password)
		id, err := commands.AddUser(userReq, persister)
		if err != nil {
			logger.Println("main: error adding user: " +
				"\n\t" + err.Error())
			return
		}
		fmt.Println("id: " + id)
	} else if os.Args[1] == "update" {
		if len(os.Args) != 4 {
			logger.Println(usage)
			return
		}
		email := os.Args[2]
		trelloKey := os.Args[3]
		persister, err := persistence.New("localhost", "users")
		if err != nil {
			logger.Println("main: error getting persister: " +
				"\n\t" + err.Error())
			return
		}
		id, err := commands.GetUserIdFromEmail(email, persister)
		if err != nil {
			logger.Println("main: error getting user id: " +
				"\n\t" + err.Error())
			return
		}
		err = commands.Update(id, trelloKey, persister)
		if err != nil {
			logger.Println("main: error updating: " +
				"\n\t" + err.Error())
			return
		}
	} else if os.Args[1] == "invoke" {
		if len(os.Args) != 5 {
			logger.Println(usage)
			return
		}
		email := os.Args[2]
		name := os.Args[3]
		trelloKey := os.Args[4]
		persister, err := persistence.New("localhost", "users")
		if err != nil {
			logger.Println("main: error getting persister: " +
				"\n\t" + err.Error())
			return
		}
		id, err := commands.GetUserIdFromEmail(email, persister)
		if err != nil {
			logger.Println("main: error getting user id: " +
				"\n\t" + err.Error())
			return
		}
		err = commands.Invoke(id, name, trelloKey, persister)
		if err != nil {
			logger.Println("main: error invoking \"" + name + "\": " +
				"\n\t" + err.Error())
			return
		}
	} else if os.Args[1] == "get-checklists" {
		if len(os.Args) != 3 {
			logger.Println(usage)
			return
		}
		email := os.Args[2]
		persister, err := persistence.New("localhost", "users")
		if err != nil {
			logger.Println("main: error getting persister: " +
				"\n\t" + err.Error())
			return
		}
		id, err := commands.GetUserIdFromEmail(email, persister)
		if err != nil {
			logger.Println("main: error getting user id: " +
				"\n\t" + err.Error())
			return
		}
		checklistsString, err := commands.GetChecklistsString(id, persister)
		if err != nil {
			logger.Println("main: error getting checklists: " +
				"\n\t" + err.Error())
			return
		}
		fmt.Println(checklistsString)
	} else if os.Args[1] == "add-checklist" {
		if len(os.Args) != 5 {
			logger.Println(usage)
			return
		}
		email := os.Args[2]
		name := os.Args[3]
		filename := os.Args[4]
		persister, err := persistence.New("localhost", "users")
		if err != nil {
			logger.Println("main: error getting persister: " +
				"\n\t" + err.Error())
			return
		}
		id, err := commands.GetUserIdFromEmail(email, persister)
		if err != nil {
			logger.Println("main: error getting user id: " +
				"\n\t" + err.Error())
			return
		}
		checklistBytes, err := ioutil.ReadFile(filename)
		if err != nil {
			logger.Println("main: error reading file \"" + filename + "\": " +
				"\n\t" + err.Error())
			return
		}
		checklistReq := fmt.Sprintf("{\"name\":\"%s\",\"checklist\":%s}", name, string(checklistBytes))
		err = commands.AddChecklist(id, checklistReq, persister)
		if err != nil {
			logger.Println("main: error adding checklist: " +
				"\n\t" + err.Error())
			return
		}
	} else if os.Args[1] == "update-checklist" {
		if len(os.Args) != 5 {
			logger.Println(usage)
			return
		}
		email := os.Args[2]
		name := os.Args[3]
		filename := os.Args[4]
		persister, err := persistence.New("localhost", "users")
		if err != nil {
			logger.Println("main: error getting persister: " +
				"\n\t" + err.Error())
			return
		}
		id, err := commands.GetUserIdFromEmail(email, persister)
		if err != nil {
			logger.Println("main: error getting user id: " +
				"\n\t" + err.Error())
			return
		}
		checklistBytes, err := ioutil.ReadFile(filename)
		if err != nil {
			logger.Println("main: error reading file \"" + filename + "\": " +
				"\n\t" + err.Error())
			return
		}
		err = commands.UpdateChecklist(id, name, string(checklistBytes), persister)
		if err != nil {
			logger.Println("main: error updating checklist: " +
				"\n\t" + err.Error())
			return
		}
	} else if os.Args[1] == "delete-checklist" {
		if len(os.Args) != 4 {
			logger.Println(usage)
			return
		}
		email := os.Args[2]
		name := os.Args[3]
		persister, err := persistence.New("localhost", "users")
		if err != nil {
			logger.Println("main: error getting persister: " +
				"\n\t" + err.Error())
			return
		}
		id, err := commands.GetUserIdFromEmail(email, persister)
		if err != nil {
			logger.Println("main: error getting user id: " +
				"\n\t" + err.Error())
			return
		}
		err = commands.DeleteChecklist(id, name, persister)
		if err != nil {
			logger.Println("main: error deleting checklist: " +
				"\n\t" + err.Error())
			return
		}
	} else if os.Args[1] == "set-todoist-token" {
		if len(os.Args) != 4 {
			logger.Println(usage)
			return
		}
		email := os.Args[2]
		token := os.Args[3]
		persister, err := persistence.New("localhost", "users")
		if err != nil {
			logger.Println("main: error getting persister: " +
				"\n\t" + err.Error())
			return
		}
		id, err := commands.GetUserIdFromEmail(email, persister)
		if err != nil {
			logger.Println("main: error getting user id: " +
				"\n\t" + err.Error())
			return
		}
		err = commands.SetTodoistToken(id, token, persister)
		if err != nil {
			logger.Println("main: error setting todoist token: " +
				"\n\t" + err.Error())
			return
		}
	} else if os.Args[1] == "set-trello-token" {
		if len(os.Args) != 4 {
			logger.Println(usage)
			return
		}
		email := os.Args[2]
		token := os.Args[3]
		persister, err := persistence.New("localhost", "users")
		if err != nil {
			logger.Println("main: error getting persister: " +
				"\n\t" + err.Error())
			return
		}
		id, err := commands.GetUserIdFromEmail(email, persister)
		if err != nil {
			logger.Println("main: error getting user id: " +
				"\n\t" + err.Error())
			return
		}
		err = commands.SetTrelloToken(id, token, persister)
		if err != nil {
			logger.Println("main: error setting trello token: " +
				"\n\t" + err.Error())
			return
		}
	} else if os.Args[1] == "get-general-settings" {
		if len(os.Args) != 3 {
			logger.Println(usage)
			return
		}
		email := os.Args[2]
		persister, err := persistence.New("localhost", "users")
		if err != nil {
			logger.Println("main: error getting persister: " +
				"\n\t" + err.Error())
			return
		}
		id, err := commands.GetUserIdFromEmail(email, persister)
		if err != nil {
			logger.Println("main: error getting user id: " +
				"\n\t" + err.Error())
			return
		}
		settings, err := commands.GetGeneralSettings(id, persister)
		if err != nil {
			logger.Println("main: error getting settings: " +
				"\n\t" + err.Error())
			return
		}
		fmt.Println(settings)
	} else if os.Args[1] == "set-general-setting" {
		if len(os.Args) != 5 {
			logger.Println(usage)
			return
		}
		email := os.Args[2]
		setting := os.Args[3]
		value := os.Args[4]
		persister, err := persistence.New("localhost", "users")
		if err != nil {
			logger.Println("main: error getting persister: " +
				"\n\t" + err.Error())
			return
		}
		id, err := commands.GetUserIdFromEmail(email, persister)
		if err != nil {
			logger.Println("main: error getting user id: " +
				"\n\t" + err.Error())
			return
		}
		err = commands.SetGeneralSetting(id, setting, value, persister)
		if err != nil {
			logger.Println("main: error setting \"" + setting + "\": " +
				"\n\t" + err.Error())
			return
		}
	} else {
		logger.Println(usage)
	}
}
