package commands

import (
	"errors"
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

func SetConfig(path string) error {
	conf, err := config.New(path)
	if err != nil {
		return errors.New("commands.SetConfig: error creating config: \n\t" + err.Error())
	}

	persist, err := persistence.Load("main")
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

func Update() error {
	persist, err := persistence.Load("main")
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

func Invoke(name string) error {
	persist, err := persistence.Load("main")
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
