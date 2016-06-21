package commands

import (
	"encoding/json"
	"errors"
	"github.com/jsutton9/preflight/checklist"
	"github.com/jsutton9/preflight/clients/todoist"
	"github.com/jsutton9/preflight/clients/trello"
	"github.com/jsutton9/preflight/persistence"
	"github.com/jsutton9/preflight/security"
	"sort"
	"time"
)

type updateJob struct {
	Checklist checklist.Checklist
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

func AddUser(email string, password string, persister *persistence.Persister) (string, error) {
	user, err := persister.AddUser(email, password)
	if err != nil {
		return "", errors.New("commands.AddUser: error adding user \"" +
			email + "\":\n\t" + err.Error())
	}
	return user.GetId(), nil
}

func GetUserIdFromEmail(email string, persister *persistence.Persister) (string, error) {
	user, err := persister.GetUserByEmail(email)
	if err != nil {
		return "", errors.New("commands.GetUserIdFromEmail: error getting user: " +
			"\n\t" + err.Error())
	}

	return user.GetId(), nil
}

func GetUserIdFromToken(secret string, persister *persistence.Persister) (string, error) {
	user, err := persister.GetUserByToken(secret)
	if err != nil {
		return "", errors.New("commands.GetUserIdFromToken: error getting user: " +
			"\n\t" + err.Error())
	}

	return user.GetId(), nil
}

func AddToken(id, tokenReqString string, persister *persistence.Persister) (string, error) {
	request := tokenRequest{}
	err := json.Unmarshal([]byte(tokenReqString), &request)
	if err != nil {
		return "", errors.New("commands.AddToken: error unmarshalling request: " +
			"\n\t" + err.Error())
	}

	user, err := persister.GetUser(id)
	if err != nil {
		return "", errors.New("commands.AddToken: error getting user: \n\t" + err.Error())
	}

	token, err := user.Security.AddToken(request.Permissions, request.ExpiryHours, request.Description)
	if err != nil {
		return "", errors.New("commands.AddToken: error adding token: " +
			"\n\t" + err.Error())
	}

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return "", errors.New("commands.AddToken: error marshalling token: " +
			"\n\t" + err.Error())
	}

	err = persister.UpdateUser(user)
	if err != nil {
		return "", errors.New("commands.AddToken: error updating user in db: " +
			"\n\t" + err.Error())
	}

	return string(tokenBytes[:]), nil
}

func DeleteToken(id, tokenId string, persister *persistence.Persister) error {
	user, err := persister.GetUser(id)
	if err != nil {
		return errors.New("commands.DeleteToken: error getting user: " +
			"\n\t" + err.Error())
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
		return errors.New("commands.DeleteToken: token id \"" + tokenId + "\" not found")
	}

	err = persister.UpdateUser(user)
	if err != nil {
		return errors.New("commands.DeleteToken: error updating user in db: " +
			"\n\t" + err.Error())
	}

	return nil
}

func GetTokens(id string, persister *persistence.Persister) (string, error) {
	user, err := persister.GetUser(id)
	if err != nil {
		return "", errors.New("commands.GetTokens: error getting user: " +
			"\n\t" + err.Error())
	}

	for i, _ := range user.Security.Tokens {
		user.Security.Tokens[i].Secret = ""
	}
	tokensBytes, err := json.Marshal(user.Security.Tokens)
	if err != nil {
		return "", errors.New("commands.GetTokens: error marshalling tokens: " +
			"\n\t" + err.Error())
	}

	return string(tokensBytes), nil
}

func SetTodoistToken(id string, token string, persister *persistence.Persister) error {
	user, err := persister.GetUser(id)
	if err != nil {
		return errors.New("commands.SetTodoistToken: error getting user: " +
			"\n\t" + err.Error())
	}

	user.Security.Todoist.Token = token

	err = persister.UpdateUser(user)
	if err != nil {
		return errors.New("commands.SetTodoistToken: error updating user in db: " +
			"\n\t" + err.Error())
	}

	return nil
}

func SetTrelloToken(id string, token string, persister *persistence.Persister) error {
	user, err := persister.GetUser(id)
	if err != nil {
		return errors.New("commands.SetTrelloToken: error getting user: " +
			"\n\t" + err.Error())
	}

	user.Security.Trello.Token = token

	err = persister.UpdateUser(user)
	if err != nil {
		return errors.New("commands.SetTrelloToken: error updating user in db: " +
			"\n\t" + err.Error())
	}

	return nil
}

func Update(id string, trelloKey string, persister *persistence.Persister) error {
	user, err := persister.GetUser(id)
	if err != nil {
		return errors.New("commands.Update: error getting user: \n\t" + err.Error())
	}

	td := todoist.New(user.Security.Todoist)
	trelloClient := trello.New(user.Security.Trello, trelloKey, user.Settings.TrelloBoard)

	loc, err := time.LoadLocation(user.Settings.Timezone)
	if err != nil {
		return errors.New("commands.Update: error loading timezone: \n\t" + err.Error())
	}
	now := time.Now().In(loc)

	jobs := make(jobsByTime, 0)
	for _, cl := range user.Checklists {
		if cl.Record == nil {
			cl.Record = &checklist.UpdateRecord{Ids:make([]int,0)}
		}
		action, updateTime, err := cl.Action(cl.Record.AddTime, cl.Record.Time, now)
		if err != nil {
			return errors.New("commands.Update: error determining action: " +
				"\n\t" + err.Error())
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
		if job.Action > 0 {
			job.Checklist.Record.Ids, err = postTasks(td, trelloClient, job.Checklist)
			if err != nil {
				return errors.New("commands.Update: error posting tasks: \n\t" + err.Error())
			}
			job.Checklist.Record.AddTime = now
		} else {
			for _, id := range job.Checklist.Record.Ids {
				err = td.DeleteTask(id)
				if err != nil {
					return errors.New("commands.Update: error deleting tasks: \n\t" + err.Error())
				}
			}
			job.Checklist.Record.Ids = make([]int, 0)
		}
		job.Checklist.Record.Time = now
	}

	err = persister.UpdateUser(user)
	if err != nil {
		return errors.New("commands.Update: error updating user in db: " +
			"\n\t" + err.Error())
	}

	return nil
}

func Invoke(id string, name string, trelloKey string, persister *persistence.Persister) error {
	user, err := persister.GetUser(id)
	if err != nil {
		return errors.New("commands.Invoke: error getting user: \n\t" + err.Error())
	}
	cl, found := user.Checklists[name]
	if ! found {
		return errors.New("commands.Invoke: checklist \""+name+"\" not found")
	}

	todoistClient := todoist.New(user.Security.Todoist)
	trelloClient := trello.New(user.Security.Trello, trelloKey, user.Settings.TrelloBoard)
	if cl.Record == nil {
		cl.Record = &checklist.UpdateRecord{Ids:make([]int, 0)}
	}

	// TODO: move time and ids updates to postTasks?
	loc, err := time.LoadLocation(user.Settings.Timezone)
	if err != nil {
		return errors.New("commands.Invoke: error loading timezone: \n\t" + err.Error())
	}
	now := time.Now().In(loc)

	cl.Record.Ids, err = postTasks(todoistClient, trelloClient, cl)
	if err != nil {
		return errors.New("commands.Invoke: error posting tasks: \n\t" + err.Error())
	}
	cl.Record.Time = now
	err = persister.UpdateUser(user)
	if err != nil {
		return errors.New("commands.Invoke: error updating user in db: " +
			"\n\t" + err.Error())
	}

	return nil
}

func AddChecklist(id, checklistReqString string, persister *persistence.Persister) error {
	request := checklistRequest{}
	err := json.Unmarshal([]byte(checklistReqString), &request)
	if err != nil {
		return errors.New("commands.AddChecklist: error parsing checklistReqString: " +
			"\n\t" + err.Error())
	}

	user, err := persister.GetUser(id)
	if err != nil {
		return errors.New("commands.AddChecklist: error getting user: \n\t" + err.Error())
	}
	_, found := user.Checklists[request.Name]
	if found {
		return errors.New("commands.AddChecklist: checklist \"" +
				request.Name + "\" already exists")
	}

	user.Checklists[request.Name] = request.Checklist
	err = persister.UpdateUser(user)
	if err != nil {
		return errors.New("commands.AddChecklist: error updating user in db: " +
			"\n\t" + err.Error())
	}

	return nil
}

func UpdateChecklist(id, name, checklistString string, persister *persistence.Persister) error {
	cl := checklist.Checklist{}
	err := json.Unmarshal([]byte(checklistString), &cl)
	if err != nil {
		return errors.New("commands.UpdateChecklist: error parsing templateString:" +
			"\n\t" + err.Error())
	}

	user, err := persister.GetUser(id)
	if err != nil {
		return errors.New("commands.UpdateChecklist: error getting user: " +
			"\n\t" + err.Error())
	}
	_, found := user.Checklists[name]
	if ! found {
		return errors.New("command.UpdateChecklist: checklist \""+name+"\" not found")
	}

	user.Checklists[name] = cl
	err = persister.UpdateUser(user)
	if err != nil {
		return errors.New("commands.UpdateChecklist: error updating user in db: " +
			"\n\t" + err.Error())
	}

	return nil
}

func DeleteChecklist(id, name string, persister *persistence.Persister) error {
	user, err := persister.GetUser(id)
	if err != nil {
		return errors.New("commands.DeleteChecklist: error getting user: \n\t" + err.Error())
	}
	_, found := user.Checklists[name]
	if ! found {
		return errors.New("commands.DeleteChecklist: checklist \""+name+"\"not found")
	}

	delete(user.Checklists, name)

	err = persister.UpdateUser(user)
	if err != nil {
		return errors.New("commands.DeleteChecklist: error updating user in db: " +
			"\n\t" + err.Error())
	}

	return nil
}

func GetChecklistString(id, name string, persister *persistence.Persister) (string, error) {
	user, err := persister.GetUser(id)
	if err != nil {
		return "", errors.New("commands.GetChecklistString: error getting user: " +
			"\n\t" + err.Error())
	}
	cl, found := user.Checklists[name]
	if ! found {
		return "", errors.New("commands.GetChecklistString: checklist \"" + name + "\" not found")
	}

	jsonBytes, err := json.Marshal(cl)
	if err != nil {
		return "", errors.New("commands.GetChecklistString: error marshalling checklist: " +
			"\n\t" + err.Error())
	}

	return string(jsonBytes[:]), nil
}

func GetChecklistsString(id string, persister *persistence.Persister) (string, error) {
	user, err := persister.GetUser(id)
	if err != nil {
		return "", errors.New("commands.GetChecklistsString: error getting user: " +
			"\n\t" + err.Error())
	}

	jsonBytes, err := json.Marshal(user.Checklists)
	if err != nil {
		return "", errors.New("commands.GetChecklistsString: error marshalling checklists: " +
			"\n\t" + err.Error())
	}

	return string(jsonBytes[:]), nil
}

func GetGeneralSettings(id string, persister *persistence.Persister) (string, error) {
	user, err := persister.GetUser(id)
	if err != nil {
		return "", errors.New("commands.GetGeneralSettings: error gettingUser: " +
			"\n\t" + err.Error())
	}

	settingsBytes, err := json.Marshal(user.Settings)
	if err != nil {
		return "", errors.New("commands.GetGeneralSettings: error marshalling settings: " +
			"\n\t" + err.Error())
	}

	return string(settingsBytes), nil
}

func SetGeneralSetting(id, name, value string, persister *persistence.Persister) error {
	user, err := persister.GetUser(id)
	if err != nil {
		return errors.New("commands.SetGeneralSetting: error getting user: " +
			"\n\t" + err.Error())
	}

	if name == "timezone" {
		user.Settings.Timezone = value
	} else if name == "trelloBoard" {
		user.Settings.TrelloBoard = value
	} else {
		return errors.New("commands.SetGeneralSetting: setting \"" +
			name + "\" not recognized")
	}

	err = persister.UpdateUser(user)
	if err != nil {
		return errors.New("commands.SetGeneralSettings: error updating user in db: " +
			"\n\t" + err.Error())
	}
	return nil
}

func postTasks(c todoist.Client, trl trello.Client, checklist checklist.Checklist) ([]int, error) {
	ids := make([]int, 0)

	if checklist.TasksSource == "preflight" {
		for _, task := range checklist.Tasks {
			id, err := c.PostTask(task)
			if err != nil {
				return ids, errors.New("commands.postTasks: error posting tasks:" +
					"\n\t" + err.Error())
			}
			ids = append(ids, id)
		}
	} else if checklist.TasksSource == "trello" {
		tasks, err := trl.Tasks(checklist.Trello)
		if err != nil {
			return ids, errors.New("commands.postTasks: error getting tasks from trello:" +
				"\n\t" + err.Error())
		}
		for _, task := range tasks {
			id, err := c.PostTask(task)
			if err != nil {
				return ids, errors.New("commands.postTasks: error posting tasks:" +
					"\n\t" + err.Error())
			}
			ids = append(ids, id)
		}
	}

	return ids, nil
}
