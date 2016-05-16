package main

import (
	"errors"
	"fmt"
	"github.com/jsutton9/preflight/clients/todoist"
	"github.com/jsutton9/preflight/clients/trello"
	"github.com/jsutton9/preflight/config"
	"github.com/jsutton9/preflight/persistence"
	"os"
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

func main() {
	usage := "Usage: preflight (TEMPLATE_NAME | update | config CONFIG_FILE)"
	if len(os.Args) < 2 {
		fmt.Println(usage)
	} else if os.Args[1] == "update" {
		if len(os.Args) == 2 {
			err := update()
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println(usage)
		}
	} else if os.Args[1] == "config" {
		if len(os.Args) == 3 {
			err := setConfig(os.Args[2])
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println(usage)
		}
	} else {
		if len(os.Args) == 2 {
			err := invoke(os.Args[1])
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println(usage)
		}
	}
}

func setConfig(path string) error {
	conf, err := config.New(path)
	if err != nil {
		return err
	}

	persist, err := persistence.Load("main")
	if err != nil {
		return err
	}

	persist.Config = *conf
	err = persist.Save()
	if err != nil {
		return err
	}

	return nil
}

func update() error {
	persist, err := persistence.Load("main")
	if err != nil {
		return err
	}

	c := todoist.New(persist.Config.ApiToken)
	tr := persist.Config.Trello
	trelloClient := trello.New(tr.Key, tr.Token, tr.BoardName)

	loc, err := time.LoadLocation(persist.Config.Timezone)
	if err != nil {
		return err
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
			return err
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
				return err
			}
			job.Record.AddTime = now
		} else {
			for _, id := range job.Record.Ids {
				err := c.DeleteTask(id)
				if err != nil {
					return err
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

func invoke(name string) error {
	persist, err := persistence.Load("main")
	if err != nil {
		return err
	}
	template, found := persist.Config.Templates[name]
	if ! found {
		return errors.New("Template \""+name+"\" not found")
	}

	todoistClient := todoist.New(persist.Config.ApiToken)
	tr := persist.Config.Trello
	trelloClient := trello.New(tr.Key, tr.Token, tr.BoardName)
	record, found := persist.UpdateHistory[name]
	if ! found {
		persist.UpdateHistory[name] = persistence.UpdateRecord{Ids:make([]int, 0)}
	}

	loc, err := time.LoadLocation(persist.Config.Timezone)
	if err != nil {
		return err
	}
	now := time.Now().In(loc)

	record.Ids, err = postTasks(todoistClient, trelloClient, template)
	if err != nil {
		return err
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
				return ids, err
			}
			ids = append(ids, id)
		}
	}

	if template.Trello != nil && template.Trello.ListName != "" {
		p := template.Trello
		tasks, err := trl.Tasks(p.Key, p.Token, p.BoardName, p.ListName)
		if err != nil {
			return ids, err
		}
		for _, task := range tasks {
			id, err := c.PostTask(task)
			if err != nil {
				return ids, err
			}
			ids = append(ids, id)
		}
	}

	return ids, nil
}
