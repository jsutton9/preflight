package config

import (
	"encoding/json"
	"errors"
)

type Config struct {
	ApiToken string               `json:"api_token"`
	Templates map[string]Template `json:"templates"`
}

type Template struct {
	Tasks []string    `json:"tasks"`
	Schedule schedule `json:"schedule,omitempty"`
}

type schedule struct {
	Interval string  `json:"interval,omitempty"`
	Days []string    `json:"days,omitempty"`
	Start string     `json:"start_time"`
	End string       `json:"end_time,omitempty"`
}

func New(filename string) Config, error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := Config{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	/*for name, template := range config.Templates {
		if err = template.Schedule.Validate(name) {
			return config, err
		}
	}*/

	return config, nil
}

func (t Template) Action(lastUpdate time.Time) string {
	s := t.Schedule
	if s == nil {
		return nil
	}
	//TODO
}

/*func (s Schedule) Validate(name string) error {
	if s == nil {
		return nil
	}

	if s.Start == nil {
		return errors.New("Invalid config for template \""+name+"\": "+
		"schedule must have start_time")
	} else {
		_, err = time.Parse("15:04 MST", s.Start)
		if err != nil {
			return errors.New("Invalid config for template \""+name+"\": "+
			"start_time must be formatted like \"15:04 MST\"")
		}
	}

	if s.End != nil {
		_, err = time.Parse("15:04 MST", s.End)
		if err != nil {
			return errors.New("Invalid config for template \""+name+"\": "+
			"end_time must be formatted like \"15:04 MST\"")
		}
	}

	if s.

	return nil
}*/ //TODO: integrate validation into Template.Action
