package config

import (
	"encoding/json"
	"errors"
	"time"
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
	Interval int     `json:"interval,omitempty"`
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

	return config, nil
}

func (t Template) Action(lastUpdate time.Time) string, error {
	return t.Schedule.Action(lastUpdate)
}

func (s Schedule) targetState(t time.Time) bool, error {
	var scheduledDay bool
	if s.Days != nil {
		currentWeekday := t.Weekday().String()
		scheduledDay = false
		for _, weekday := range s.Days {
			if weekday == currentWeekday {
				scheduledDay = true
				break
			}
		}
	} else {
		scheduledDay = true
	}

	if ! scheduledDay {
		if s.End == nil {
			return true, nil
		} else {
			return false, nil
		}
	}

	y, m, d = t.getDate()

	start, err := time.Parse("15:04 MST", s.Start)
	if err != nil {
		return nil, err
	}

	var end time.Time
	if s.End != nil {
		end, err = time.Parse("15:04 MST", s.End)
		if err != nil {
			return nil, err
		}
	} else {
		end = time.Date(3000, 1, 1, 0, 0, 0, 0, t.Location())
	}

	if t.After(start) && t.Before(end) {
		return true, nil
	} else {
		return false, nil
	}
}


func (s Schedule) Action(lastUpdate time.Time, now time.Time) int, error {
	if s == nil {
		return nil, nil
	}

	targetNow, err := s.targetState(now)
	if err != nil {
		return nil, err
	}
	targetLast, err := s.targetState(lastUpdate)
	if err != nil {
		return nil, err
	}

	if (! targetLast) && targetNow {
		return 1, nil
	} else if targetLast && (! targetNow) {
		return -1, nil
	} else {
		return 0, nil
	}
}
