package checklist

import (
	"github.com/jsutton9/preflight/api/errors"
	"github.com/jsutton9/preflight/clients/todoist"
	"github.com/jsutton9/preflight/clients/trello"
	"os/exec"
)

type Checklist struct {
	Id string              `json:"id"`
	Name string            `json:"name"`
	TasksSource string     `json:"tasksSource"`
	TasksTarget string     `json:"tasksTarget"`
	IsScheduled bool       `json:"isScheduled"`
	Tasks []string         `json:"tasks,omitempty"`
	Trello *trello.ListKey `json:"trello,omitempy"`
	Schedule *Schedule     `json:"schedule,omitempty"`
}

func (c *Checklist) GenId() *errors.PreflightError {
	idBytes, err := exec.Command("uuidgen").Output()
	if err != nil {
		return &errors.PreflightError{
			Status: 500,
			InternalMessage: "checklist.Checklist.SetId: error generating uuid: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error creating the checklist.",
		}
	}
	c.Id = string(idBytes[:len(idBytes)-1])
	return nil
}

func (c *Checklist) Equals(b *Checklist) bool {
	if c==nil && b==nil {
		return true
	} else if c==nil || b==nil {
		return false
	}

	if c.Id!=b.Id || c.Name!=b.Name || c.TasksSource!=b.TasksSource || c.IsScheduled!=b.IsScheduled {
		return false
	} else if ! c.Trello.Equals(b.Trello) {
		return false
	} else if ! c.Schedule.Equals(b.Schedule) {
		return false
	} else if len(c.Tasks) != len(b.Tasks) {
		return false
	} else {
		for i, task := range c.Tasks {
			if task != b.Tasks[i] {
				return false
			}
		}
	}

	return true
}

func (c *Checklist) PostTasks(tdst todoist.Client, trl trello.Client) ([]int, *errors.PreflightError) {
	ids := make([]int, 0)

	if c.TasksSource == "preflight" {
		for _, task := range c.Tasks {
			id, pErr := tdst.PostTask(task)
			if pErr != nil {
				return ids, pErr.Prepend("Checklist.PostTasks: error posting task:")
			}
			ids = append(ids, id)
		}
	} else if c.TasksSource == "trello" {
		tasks, pErr := trl.Tasks(c.Trello)
		if pErr != nil {
			return ids, pErr.Prepend("Checklist.PostTasks: error getting tasks from trello:")
		}
		for _, task := range tasks {
			id, pErr := tdst.PostTask(task)
			if pErr != nil {
				return ids, pErr.Prepend("Checklist.PostTasks: error posting task:")
			}
			ids = append(ids, id)
		}
	}

	return ids, nil
}

func (c *Checklist) DeleteTasks(tdst todoist.Client, ids []int) *errors.PreflightError {
	for _, id := range ids {
		err := tdst.DeleteTask(id)
		if err != nil {
			return err.Prepend("Checklist.DeleteTasks: error deleting task: ")
		}
	}
	return nil
}
