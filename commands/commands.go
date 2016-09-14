package commands

import (
	"encoding/json"
	"github.com/jsutton9/preflight/api/errors"
	"github.com/jsutton9/preflight/checklist"
	"github.com/jsutton9/preflight/clients/todoist"
	"github.com/jsutton9/preflight/clients/trello"
	"github.com/jsutton9/preflight/persistence"
	"github.com/jsutton9/preflight/security"
	"sort"
	"time"
)

type updateJob struct {
	Checklist *checklist.Checklist
	Action int
	Time time.Time
}

type jobsByTime []updateJob

func (l jobsByTime) Len() int {
	return len(l)
}
func (l jobsByTime) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l jobsByTime) Less(i, j int) bool {
	return l[i].Time.Before(l[j].Time)
}

type checklistRequest struct {
	Name      string              `json:"name"`
	Checklist checklist.Checklist `json:"checklist"`
}

type tokenRequest struct {
	Permissions security.PermissionFlags `json:"permissions"`
	ExpiryHours int                      `json:"expiryHours"`
	Description string                   `json:"description"`
}

type userRequest struct {
	Email string    `json:"email"`
	Password string `json:"password"`
}

func AddUser(userReqString string, persister *persistence.Persister) (string, *errors.PreflightError) {
	request := userRequest{}
	err := json.Unmarshal([]byte(userReqString), &request)
	if err != nil {
		return "", &errors.PreflightError{
			Status: 400,
			InternalMessage: "commands.AddUser: error parsing userReqString: " +
				"\n\t" + err.Error(),
			ExternalMessage: "Request body is invalid.",
		}
	}

	user, pErr := persister.AddUser(request.Email, request.Password)
	if pErr != nil {
		return "", pErr.Prepend("commands.AddUser: error adding user \"" +
			request.Email + "\":")
	}
	return user.GetId(), nil
}

func DeleteUser(id string, persister *persistence.Persister) *errors.PreflightError {
	user, err := persister.GetUser(id)
	if err != nil {
		return err.Prepend("commands.DeleteUser: error getting user: ")
	}

	err = persister.DeleteUser(user)
	if err != nil {
		return err.Prepend("commands.DeleteUser: error deleting user: ")
	}

	return nil
}

func GetUserIdFromEmail(email string, persister *persistence.Persister) (string, *errors.PreflightError) {
	user, err := persister.GetUserByEmail(email)
	if err != nil {
		return "", err.Prepend("commands.GetUserIdFromEmail: error getting user: ")
	}

	return user.GetId(), nil
}

func GetUserIdFromToken(secret string, persister *persistence.Persister) (string, *errors.PreflightError) {
	user, err := persister.GetUserByToken(secret)
	if err != nil {
		return "", err.Prepend("commands.GetUserIdFromToken: error getting user: ")
	}

	return user.GetId(), nil
}

func ChangePassword(id string, newPassword string, persister *persistence.Persister) *errors.PreflightError {
	user, err := persister.GetUser(id)
	if err != nil {
		return err.Prepend("commands.ChangePassword: error getting user: ")
	}

	err = user.Security.SetPassword(newPassword)
	if err != nil {
		return err.Prepend("commands.ChangePassword: error setting password: ")
	}

	err = persister.UpdateUser(user)
	if err != nil {
		return err.Prepend("commands.ChangePassword: error updating user in db: ")
	}

	return nil
}

func ValidatePassword(id string, password string, persister *persistence.Persister) *errors.PreflightError {
	user, err := persister.GetUser(id)
	if err != nil {
		return err.Prepend("commands.ValidatePassword: error getting user: ")
	}

	err = user.Security.ValidatePassword(password)
	if err != nil {
		return err.Prepend("commands.ValidatePassword: error validating password: ")
	}

	return nil
}

func ValidateNodeSecret(secret string, persister *persistence.Persister) *errors.PreflightError {
	valid, err := persister.ValidateNodeSecret(secret)
	if err != nil {
		err.Prepend("commands.ValidateNodeSecret: error validating node secret: ")
		return err
	} else if ! valid {
		return &errors.PreflightError{
			Status: 401,
			InternalMessage: "commands.ValidateNodeSecret: Node secret invalid",
			ExternalMessage: "Node secret invalid",
		}
	}

	return nil
}

func ValidateToken(id string, secret string, permissions security.PermissionFlags, persister *persistence.Persister) *errors.PreflightError {
	user, err := persister.GetUser(id)
	if err != nil {
		return err.Prepend("commands.ValidateToken: error getting user: ")
	}

	err = user.Security.ValidateToken(secret, permissions)
	if err != nil {
		return err.Prepend("commands.ValidateToken: error validating token: ")
	}

	return nil
}

func AddToken(id, tokenReqString string, persister *persistence.Persister) (string, *errors.PreflightError) {
	request := tokenRequest{}
	err := json.Unmarshal([]byte(tokenReqString), &request)
	if err != nil {
		return "", &errors.PreflightError{
			Status: 400,
			InternalMessage: "commands.AddToken: error unmarshalling request: " +
				"\n\t" + err.Error(),
			ExternalMessage: "Request body is invalid.",
		}
	}

	user, pErr := persister.GetUser(id)
	if pErr != nil {
		return "", pErr.Prepend("commands.AddToken: error getting user: ")
	}

	token, pErr := user.Security.AddToken(request.Permissions, request.ExpiryHours, request.Description)
	if pErr != nil {
		return "", pErr.Prepend("commands.AddToken: error adding token: ")
	}

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return "", &errors.PreflightError{
			Status: 500,
			InternalMessage: "commands.AddToken: error marshalling token: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error creating the token.",
		}
	}

	pErr = persister.UpdateUser(user)
	if pErr != nil {
		return "", pErr.Prepend("commands.AddToken: error updating user in db: ")
	}

	return string(tokenBytes[:]), nil
}

func DeleteToken(id, tokenId string, persister *persistence.Persister) *errors.PreflightError {
	user, err := persister.GetUser(id)
	if err != nil {
		return err.Prepend("commands.DeleteToken: error getting user: ")
	}

	tokens := user.Security.Tokens
	found := false
	for i, token := range tokens {
		if token.Id == tokenId {
			found = true
			tokens[i] = tokens[len(tokens)-1]
			user.Security.Tokens = tokens[:len(tokens)-1]
			break
		}
	}

	if ! found {
		return &errors.PreflightError{
			Status: 404,
			InternalMessage: "commands.DeleteToken: token id \"" + tokenId + "\" not found",
			ExternalMessage: "Token not found",
		}
	}

	err = persister.UpdateUser(user)
	if err != nil {
		return err.Prepend("commands.DeleteToken: error updating user in db: ")
	}

	return nil
}

func GetTokens(id string, persister *persistence.Persister) (string, *errors.PreflightError) {
	user, pErr := persister.GetUser(id)
	if pErr != nil {
		return "", pErr.Prepend("commands.GetTokens: error getting user: ")
	}

	for i, _ := range user.Security.Tokens {
		user.Security.Tokens[i].Secret = ""
	}
	tokensBytes, err := json.Marshal(user.Security.Tokens)
	if err != nil {
		return "", &errors.PreflightError{
			Status: 500,
			InternalMessage: "commands.GetTokens: error marshalling tokens: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error getting the tokens.",
		}
	}

	return string(tokensBytes), nil
}

func SetTodoistToken(id string, token string, persister *persistence.Persister) *errors.PreflightError {
	user, err := persister.GetUser(id)
	if err != nil {
		return err.Prepend("commands.SetTodoistToken: error getting user: ")
	}

	user.Security.Todoist.Token = token

	err = persister.UpdateUser(user)
	if err != nil {
		return err.Prepend("commands.SetTodoistToken: error updating user in db: ")
	}

	return nil
}

func SetTrelloToken(id string, token string, persister *persistence.Persister) *errors.PreflightError {
	user, err := persister.GetUser(id)
	if err != nil {
		return err.Prepend("commands.SetTrelloToken: error getting user: ")
	}

	user.Security.Trello.Token = token

	err = persister.UpdateUser(user)
	if err != nil {
		return err.Prepend("commands.SetTrelloToken: error updating user in db: ")
	}

	return nil
}

func Update(id string, trelloKey string, persister *persistence.Persister) *errors.PreflightError {
	user, pErr := persister.GetUser(id)
	if pErr != nil {
		return pErr.Prepend("commands.Update: error getting user: ")
	}

	td := todoist.New(user.Security.Todoist)
	trelloClient := trello.New(user.Security.Trello, trelloKey, user.Settings.TrelloBoard)

	loc, err := time.LoadLocation(user.Settings.Timezone)
	if err != nil {
		return &errors.PreflightError{
			Status: 424,
			InternalMessage: "commands.Update: error loading timezone: " +
				"\n\t" + err.Error(),
			ExternalMessage: "Could not find your timezone in the IANA database.",
		}
	}
	now := time.Now().In(loc)

	jobs := make(jobsByTime, 0)
	for _, cl := range user.Checklists {
		if cl.Record == nil {
			cl.Record = &checklist.UpdateRecord{Ids:make([]int,0)}
		}
		action, updateTime, pErr := cl.Action(cl.Record.AddTime, cl.Record.Time, now)
		if pErr != nil {
			return pErr.Prepend("commands.Update: error determining action: ")
		}
		if action != 0 {
			jobs = append(jobs, updateJob{
				Checklist: cl,
				Action: action,
				Time: updateTime,
			})
		}
	}

	sort.Stable(jobs)
	for _, job := range jobs {
		if job.Checklist.Record == nil {
			job.Checklist.Record = new(checklist.UpdateRecord)
		}
		if job.Action > 0 {
			job.Checklist.Record.Ids, pErr = postTasks(td, trelloClient, *job.Checklist)
			if pErr != nil {
				return pErr.Prepend("commands.Update: error posting tasks: ")
			}
			job.Checklist.Record.AddTime = now
		} else {
			for _, id := range job.Checklist.Record.Ids {
				pErr = td.DeleteTask(id)
				if pErr != nil {
					return pErr.Prepend("commands.Update: error deleting tasks: ")
				}
			}
			job.Checklist.Record.Ids = make([]int, 0)
		}
		job.Checklist.Record.Time = now
	}

	pErr = persister.UpdateUser(user)
	if pErr != nil {
		return pErr.Prepend("commands.Update: error updating user in db: ")
	}

	return nil
}

func Invoke(id string, name string, trelloKey string, persister *persistence.Persister) *errors.PreflightError {
	user, pErr := persister.GetUser(id)
	if pErr != nil {
		return pErr.Prepend("commands.Invoke: error getting user: ")
	}
	cl, found := user.Checklists[name]
	if ! found {
		return &errors.PreflightError{
			Status: 404,
			InternalMessage: "commands.Invoke: checklist \""+name+"\" not found",
			ExternalMessage: "Checklist \""+name+"\" not found",
		}
	}

	todoistClient := todoist.New(user.Security.Todoist)
	trelloClient := trello.New(user.Security.Trello, trelloKey, user.Settings.TrelloBoard)
	if cl.Record == nil {
		cl.Record = &checklist.UpdateRecord{Ids:make([]int, 0)}
	}

	// TODO: move time and ids updates to postTasks?
	loc, err := time.LoadLocation(user.Settings.Timezone)
	if err != nil {
		return &errors.PreflightError{
			Status: 424,
			InternalMessage: "commands.Invoke: error loading timezone: " +
				"\n\t" + err.Error(),
			ExternalMessage: "Could not find your timezone in the IANA database.",
		}
	}
	now := time.Now().In(loc)

	cl.Record.Ids, pErr = postTasks(todoistClient, trelloClient, *cl)
	if pErr != nil {
		return pErr.Prepend("commands.Invoke: error posting tasks: ")
	}
	cl.Record.Time = now
	pErr = persister.UpdateUser(user)
	if pErr != nil {
		return pErr.Prepend("commands.Invoke: error updating user in db: ")
	}

	return nil
}

func AddChecklist(id, checklistReqString string, persister *persistence.Persister) (string, *errors.PreflightError) {
	request := checklistRequest{}
	err := json.Unmarshal([]byte(checklistReqString), &request)
	if err != nil {
		return "", &errors.PreflightError{
			Status: 400,
			InternalMessage: "commands.AddUser: error parsing checklistReqString: " +
				"\n\t" + err.Error(),
			ExternalMessage: "Request body is invalid.",
		}
	}

	user, pErr := persister.GetUser(id)
	if pErr != nil {
		return "", pErr.Prepend("commands.AddChecklist: error getting user: ")
	}
	_, found := user.Checklists[request.Name]
	if found {
		return "", &errors.PreflightError{
			Status: 409,
			InternalMessage: "commands.AddChecklist: checklist \"" +
				request.Name + "\" already exists",
			ExternalMessage: "Checklist \"" + request.Name + "\" already exists.",
		}
	}

	pErr = request.Checklist.GenId()
	if pErr != nil {
		return "", pErr.Prepend("commands.AddChecklist: error generating ID: ")
	}
	user.Checklists[request.Name] = &request.Checklist
	pErr = persister.UpdateUser(user)
	if pErr != nil {
		return "", pErr.Prepend("commands.AddChecklist: error updating user in db: ")
	}

	return request.Name, nil
}

func UpdateChecklist(id, name, checklistString string, persister *persistence.Persister) *errors.PreflightError {
	cl := checklist.Checklist{}
	err := json.Unmarshal([]byte(checklistString), &cl)
	if err != nil {
		return &errors.PreflightError{
			Status: 400,
			InternalMessage: "commands.UpdateChecklist: error parsing templateString:" +
				"\n\t" + err.Error(),
			ExternalMessage: "Request body is invalid.",
		}
	}

	user, pErr := persister.GetUser(id)
	if pErr != nil {
		return pErr.Prepend("commands.UpdateChecklist: error getting user: ")
	}
	clOld, found := user.Checklists[name]
	if ! found {
		return &errors.PreflightError{
			Status: 404,
			InternalMessage: "command.UpdateChecklist: checklist \""+name+"\" not found",
			ExternalMessage: "Checklist \""+name+"\" not found.",
		}
	}

	cl.Record = clOld.Record

	user.Checklists[name] = &cl
	pErr = persister.UpdateUser(user)
	if pErr != nil {
		return pErr.Prepend("commands.UpdateChecklist: error updating user in db: ")
	}

	return nil
}

func DeleteChecklist(id, name string, persister *persistence.Persister) *errors.PreflightError {
	user, pErr := persister.GetUser(id)
	if pErr != nil {
		return pErr.Prepend("commands.DeleteChecklist: error getting user: ")
	}
	_, found := user.Checklists[name]
	if ! found {
		return &errors.PreflightError{
			Status: 404,
			InternalMessage: "commands.DeleteChecklist: checklist \""+name+"\"not found",
			ExternalMessage: "Checklist \""+name+"\" not found.",
		}
	}

	delete(user.Checklists, name)

	pErr = persister.UpdateUser(user)
	if pErr != nil {
		return pErr.Prepend("commands.DeleteChecklist: error updating user in db: ")
	}

	return nil
}

func GetChecklistString(id, name string, persister *persistence.Persister) (string, *errors.PreflightError) {
	user, pErr := persister.GetUser(id)
	if pErr != nil {
		return "", pErr.Prepend("commands.GetChecklistString: error getting user: ")
	}
	cl, found := user.Checklists[name]
	if ! found {
		return "", &errors.PreflightError{
			Status: 404,
			InternalMessage: "commands.GetChecklistString: checklist \"" +
				name + "\" not found",
			ExternalMessage: "Checklist \""+name+"\" not found.",
		}
	}

	jsonBytes, err := json.Marshal(cl)
	if err != nil {
		return "", &errors.PreflightError{
			Status: 500,
			InternalMessage: "commands.GetChecklistString: error marshalling checklist: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error getting the checklist.",
		}
	}

	return string(jsonBytes[:]), nil
}

func GetChecklistsString(id string, persister *persistence.Persister) (string, *errors.PreflightError) {
	user, pErr := persister.GetUser(id)
	if pErr != nil {
		return "", pErr.Prepend("commands.GetChecklistsString: error getting user: ")
	}

	jsonBytes, err := json.Marshal(user.Checklists)
	if err != nil {
		return "", &errors.PreflightError{
			Status: 500,
			InternalMessage: "commands.GetChecklistsString: error marshalling checklists: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error getting the checklists.",
		}
	}

	return string(jsonBytes[:]), nil
}

func GetGeneralSettings(id string, persister *persistence.Persister) (string, *errors.PreflightError) {
	user, pErr := persister.GetUser(id)
	if pErr != nil {
		return "", pErr.Prepend("commands.GetGeneralSettings: error getting settings: ")
	}

	settingsBytes, err := json.Marshal(user.Settings)
	if err != nil {
		return "", &errors.PreflightError{
			Status: 500,
			InternalMessage: "commands.GetGeneralSettings: error marshalling settings: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error getting the settings.",
		}
	}

	return string(settingsBytes), nil
}

func SetGeneralSetting(id, name, value string, persister *persistence.Persister) *errors.PreflightError {
	user, pErr := persister.GetUser(id)
	if pErr != nil {
		return pErr.Prepend("commands.SetGeneralSetting: error getting user: ")
	}

	if name == "timezone" {
		user.Settings.Timezone = value
	} else if name == "trelloBoard" {
		user.Settings.TrelloBoard = value
	} else {
		return &errors.PreflightError{
			Status: 404,
			InternalMessage: "commands.SetGeneralSetting: setting \"" +
				name + "\" not recognized",
			ExternalMessage: "Setting \"" + name + "\" not recognized.",
		}
	}

	pErr = persister.UpdateUser(user)
	if pErr != nil {
		return pErr.Prepend("commands.SetGeneralSettings: error updating user in db: ")
	}
	return nil
}

func postTasks(c todoist.Client, trl trello.Client, checklist checklist.Checklist) ([]int, *errors.PreflightError) {
	ids := make([]int, 0)

	if checklist.TasksSource == "preflight" {
		for _, task := range checklist.Tasks {
			id, pErr := c.PostTask(task)
			if pErr != nil {
				return ids, pErr.Prepend("commands.postTasks: error posting tasks:")
			}
			ids = append(ids, id)
		}
	} else if checklist.TasksSource == "trello" {
		tasks, pErr := trl.Tasks(checklist.Trello)
		if pErr != nil {
			return ids, pErr.Prepend("commands.postTasks: error getting tasks from trello:")
		}
		for _, task := range tasks {
			id, pErr := c.PostTask(task)
			if pErr != nil {
				return ids, pErr.Prepend("commands.postTasks: error posting tasks:")
			}
			ids = append(ids, id)
		}
	}

	return ids, nil
}
