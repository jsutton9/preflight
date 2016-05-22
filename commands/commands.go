package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jsutton9/preflight/clients/todoist"
	"github.com/jsutton9/preflight/clients/trello"
	"github.com/jsutton9/preflight/config"
	"github.com/jsutton9/preflight/persistence"
	"sort"
	"time"
)

type updateJob struct {
	Name string
	Template config.Template
	Action int
	Time time.Time
	Record persistence.UpdateRecord
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

type templateRequest struct {
	Name     string
	Template config.Template
}

func SetConfig(user string, path string) error {
	conf, err := config.New(path)
	if err != nil {
		return errors.New("commands.SetConfig: error creating config: \n\t" + err.Error())
	}

	persist, err := persistence.Load(user)
	if err != nil {
		return errors.New("commands.SetConfig: error loading persistence: \n\t" + err.Error())
	}

	persist.Config = *conf
	err = persist.Save()
	if err != nil {
		return errors.New("commands.SetConfig: error saving persistence: \n\t" + err.Error())
	}

	return nil
}

func Update(user string) error {
	persist, err := persistence.Load(user)
	if err != nil {
		return errors.New("commands.Update: error loading persistence: \n\t" + err.Error())
	}

	c := todoist.New(persist.Config.TodoistToken)
	tr := persist.Config.Trello
	trelloClient := trello.New(tr.Key, tr.Token, tr.BoardName)

	loc, err := time.LoadLocation(persist.Config.Timezone)
	if err != nil {
		return errors.New("commands.Update: error loading timezone: \n\t" + err.Error())
	}
	now := time.Now().In(loc)

	jobs := make(jobsByTime, 0)
	for name, template := range persist.Config.Templates {
		record, found := persist.UpdateHistory[name]
		if ! found {
			persist.UpdateHistory[name] = persistence.UpdateRecord{Ids:make([]int, 0)}
		}
		action, updateTime, err := template.Action(record.AddTime, record.Time, now)
		if err != nil {
			return errors.New("commands.Update: error determining action: \n\t" + err.Error())
		}
		if action != 0 {
			jobs = append(jobs, updateJob{
				Name: name,
				Template: template,
				Action: action,
				Time: updateTime,
				Record: record,
			})
		}
	}

	sort.Stable(jobs)
	for _, job := range jobs {
		if job.Action > 0 {
			job.Record.Ids, err = postTasks(c, trelloClient, job.Template)
			if err != nil {
				return errors.New("commands.Update: error posting tasks: \n\t" + err.Error())
			}
			job.Record.AddTime = now
		} else {
			for _, id := range job.Record.Ids {
				err := c.DeleteTask(id)
				if err != nil {
					return errors.New("commands.Update: error deleting tasks: \n\t" + err.Error())
				}
			}
			job.Record.Ids = make([]int, 0)
		}
		job.Record.Time = now
		persist.UpdateHistory[job.Name] = job.Record
		persist.Save()
	}

	return nil
}

func Invoke(user, name string) error {
	persist, err := persistence.Load(user)
	if err != nil {
		return errors.New("commands.Invoke: error loading persistence: \n\t" + err.Error())
	}
	template, found := persist.Config.Templates[name]
	if ! found {
		return errors.New("commands.Invoke: template \""+name+"\" not found")
	}

	todoistClient := todoist.New(persist.Config.TodoistToken)
	tr := persist.Config.Trello
	trelloClient := trello.New(tr.Key, tr.Token, tr.BoardName)
	record, found := persist.UpdateHistory[name]
	if ! found {
		persist.UpdateHistory[name] = persistence.UpdateRecord{Ids:make([]int, 0)}
	}

	loc, err := time.LoadLocation(persist.Config.Timezone)
	if err != nil {
		return errors.New("commands.Invoke: error loading timezone: \n\t" + err.Error())
	}
	now := time.Now().In(loc)

	record.Ids, err = postTasks(todoistClient, trelloClient, template)
	if err != nil {
		return errors.New("commands.Invoke: error posting tasks: \n\t" + err.Error())
	}
	record.Time = now
	persist.UpdateHistory[name] = record
	persist.Save()

	return nil
}

func AddTemplate(user, templateReqString string) error {
	request := templateRequest{}
	err := json.Unmarshal([]byte(templateReqString), &request)
	if err != nil {
		return errors.New("commands.AddTemplate: error parsing templateReqString: " +
			"\n\t" + err.Error())
	}

	persist, err := persistence.Load(user)
	if err != nil {
		return errors.New("commands.AddTemplate: error loading persistence: \n\t" + err.Error())
	}
	_, found := persist.Config.Templates[request.Name]
	if found {
		return errors.New("commands.AddTemplate: template \"" + request.Name + "\" already exists")
	}

	persist.Config.Templates[request.Name] = request.Template
	persist.Save()

	return nil
}

func DeleteTemplate(user, name string) error {
	persist, err := persistence.Load(user)
	if err != nil {
		return errors.New("commands.DeleteTemplate: error loading persistence: \n\t" + err.Error())
	}
	_, found := persist.Config.Templates[name]
	if ! found {
		return errors.New("commands.DeleteTemplate: template not found")
	}

	delete(persist.Config.Templates, name)
	persist.Save()

	return nil
}

func GetTemplateString(user, name string) (string, error) {
	persist, err := persistence.Load(user)
	if err != nil {
		return "", errors.New("commands.GetTemplateString: error loading persistence: " +
			"\n\t" + err.Error())
	}
	template, found := persist.Config.Templates[name]
	if ! found {
		return "", errors.New("commands.GetTemplateString: template \"" + name + "\" not found")
	}

	jsonBytes, err := json.Marshal(template)
	if err != nil {
		return "", errors.New("commands.GetTemplateString: error marshalling template: " +
			"\n\t" + err.Error())
	}

	return string(jsonBytes[:]), nil
}

func GetTemplatesString(user string) (string, error) {
	persist, err := persistence.Load(user)
	if err != nil {
		return "", errors.New("commands.GetTemplatesString: error loading persistence: " +
			"\n\t" + err.Error())
	}

	jsonBytes, err := json.Marshal(persist.Config.Templates)
	if err != nil {
		return "", errors.New("commands.GetTemplatesString: error marshalling templates: " +
			"\n\t" + err.Error())
	}

	return string(jsonBytes[:]), nil
}

func GetGlobalSettings(user string) (string, error) {
	persist, err := persistence.Load(user)
	if err != nil {
		return "", errors.New("commands.GetGlobalSettings: error loading persistence: " +
			"\n\t" + err.Error())
	}

	trelloString, err := json.Marshal(persist.Config.Trello)
	if err != nil {
		return "", errors.New("commands.GetGlobalSettings: error marshalling trello: " +
			"\n\t" + err.Error())
	}

	return fmt.Sprintf("{todoist_token: %s, timezone: %s, trello: %s}",
		persist.Config.TodoistToken, persist.Config.Timezone, trelloString), nil
}

func SetGlobalSetting(user, name, value string) error {
	persist, err := persistence.Load(user)
	if err != nil {
		return errors.New("commands.SetGlobalSetting: error loading persistence: " +
			"\n\t" + err.Error())
	}

	if name == "todoist_token" {
		persist.Config.TodoistToken = value
	} else if name == "timezone" {
		persist.Config.Timezone = value
	} else if name == "trello" {
		trello := config.Trello{}
		err = json.Unmarshal([]byte(value), &trello)
		if err != nil {
			return errors.New("commands.SetGlobalSetting: error unmarshalling trello: " +
				"\n\t" + err.Error())
		}
		persist.Config.Trello = trello
	} else {
		return errors.New("commands.SetGlobalSetting: setting \"" + name + "\" not recognized")
	}

	persist.Save()
	return nil
}

func postTasks(c todoist.Client, trl trello.Client, template config.Template) ([]int, error) {
	ids := make([]int, 0)

	if template.Tasks != nil {
		for _, task := range template.Tasks {
			id, err := c.PostTask(task)
			if err != nil {
				return ids, errors.New("commands.postTasks: error posting tasks: \n\t" + err.Error())
			}
			ids = append(ids, id)
		}
	}

	if template.Trello != nil && template.Trello.ListName != "" {
		p := template.Trello
		tasks, err := trl.Tasks(p.Key, p.Token, p.BoardName, p.ListName)
		if err != nil {
			return ids, errors.New("commands.postTasks: error getting tasks from trello: \n\t" + err.Error())
		}
		for _, task := range tasks {
			id, err := c.PostTask(task)
			if err != nil {
				return ids, errors.New("commands.postTasks: error posting tasks: \n\t" + err.Error())
			}
			ids = append(ids, id)
		}
	}

	return ids, nil
}
