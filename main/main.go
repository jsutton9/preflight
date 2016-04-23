package main

import (
	"errors"
	"fmt"
	"github.com/jsutton9/todoistist/clients/todoist"
	"github.com/jsutton9/todoistist/config"
	"github.com/jsutton9/todoistist/persistence"
	"os"
	"time"
)

func main() {
	usage := "Usage: todoistist (TEMPLATE_NAME | update | config CONFIG_FILE)"
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

	loc, err := time.LoadLocation(persist.Config.Timezone)
	if err != nil {
		return err
	}
	now := time.Now().In(loc)

	for name, template := range persist.Config.Templates {
		record, found := persist.UpdateHistory[name]
		if ! found {
			persist.UpdateHistory[name] = persistence.UpdateRecord{Ids:make([]int, 0)}
		}
		action, err := template.Action(record.AddTime, record.Time, now)
		if err != nil {
			return err
		}
		if action > 0 {
			record.Ids = make([]int, 0, 20)
			for _, task := range template.Tasks {
				id, err := c.PostTask(task)
				if err != nil {
					return err
				}
				record.Ids = append(record.Ids, id)
			}
			record.AddTime = now
		} else if action < 0 {
			for _, id := range record.Ids {
				err := c.DeleteTask(id)
				if err != nil {
					return err
				}
			}
			record.Ids = make([]int, 0)
		}
		record.Time = now
		persist.UpdateHistory[name] = record
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

	c := todoist.New(persist.Config.ApiToken)
	for _, task := range template.Tasks {
		_, err := c.PostTask(task)
		if err != nil {
			return err
		}
	}

	return nil
}
