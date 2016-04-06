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

func (s Schedule) Action(lastUpdate time.Time, now time.Time) string, error {
	//TODO: move some of this in to func (s Schedule) targetState(t time.Time) bool
	if s == nil {
		return nil, nil
	}

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
		end = time.Date(3000, 1, 1, 0, 0, 0, 0, now.Location())
	}

	var scheduledToday bool
	var scheduledLast bool
	if s.Days != nil {
		currentWeekday := now.Weekday().String()
		scheduledToday = false
		for _, weekday := range s.Days {
			if weekday == currentWeekday {
				scheduledToday = true
				break
			}
		}
		if lastUpdate == nil {
			scheduledLast = false
		} else {
			lastWeekday := lastUpdate.Weekday().String()
			if lastWeekday == currentWeekday {
				scheduledLast = scheduledToday
			} else {
				scheduledLast = false
				for _, weekday := range s.Days {
					if weekday == lastWeekday {
						scheduledLast = true
						break
					}
				}
			}
		}
	} else {
		scheduledToday = true
		scheduledLast = true
	}

	var y, m, d int
	if scheduledToday {
		y, m, d = now.Date()
	} else if scheduledLast {
		y, m, d = lastUpdate.Date()
	} else if lastUpdate != nil {
		y, m, d = lastUpdate.Date()
		d -= 1
	} else {
		y, m, d = now.Date()
		d -= 1
	}

	start = start.addDate(y, m, d)
	end = end.addDate(y, m, d)

	if lastUpdate.Before(start) {
		if now.After(start) && now.Before(end) {
			return "add", nil
		} else {
			return nil, nil
		}
	} else if lastUpdate.Before(end) {
		if now.After(end) {
			return "delete", nil
		} else {
			return nil, nil
		}
	} else {
		return nil, nil
	}
}
