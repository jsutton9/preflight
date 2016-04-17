package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"time"
)

type Config struct {
	ApiToken string               `json:"api_token"`
	Timezone string               `json:"timezone",omitempty`
	Templates map[string]Template `json:"templates"`
}

type Template struct {
	Tasks []string    `json:"tasks"`
	Schedule *schedule `json:"schedule,omitempty"`
}

type schedule struct {
	Interval int     `json:"interval,omitempty"`
	Days []string    `json:"days,omitempty"`
	Start string     `json:"start"`
	End string       `json:"end,omitempty"`
}

func New(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := Config{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return &config, err
	}

	return &config, nil
}

func parseWeekday(s string) (time.Weekday, error) {
	switch s {
	case "Sunday", "sunday", "Sun", "sun":
		return time.Sunday, nil
	case "Monday", "monday", "Mon", "mon":
		return time.Monday, nil
	case "Tuesday", "tuesday", "Tues", "tues":
		return time.Tuesday, nil
	case "Wednesday", "wednesday", "Wed", "wed":
		return time.Wednesday, nil
	case "Thursday", "thursday", "Thurs", "thurs":
		return time.Thursday, nil
	case "Friday", "friday", "Fri", "fri":
		return time.Friday, nil
	case "Saturday", "saturday", "Sat", "sat":
		return time.Saturday, nil
	default:
		return time.Sunday, errors.New(
			"Unable to parse day of the week \""+s+"\"")
	}
}

func (s *schedule) Action(lastAdd time.Time, lastUpdate time.Time, now time.Time) (int, error) {
	if s == nil {
		return 0, nil
	}

	var scheduledToday bool
	lastScheduledDelta := 7
	if s.Days != nil {
		scheduledToday = false
		currentWeekday := now.Weekday()
		for _, weekdayString := range s.Days {
			weekday, err := parseWeekday(weekdayString)
			if err != nil {
				return 0, err
			}
			weekdayDelta := int(currentWeekday-weekday)
			if weekdayDelta < 0 {
				weekdayDelta += 7
			}
			if weekdayDelta == 0 {
				scheduledToday = true
			} else if weekdayDelta < lastScheduledDelta {
				lastScheduledDelta = weekdayDelta
			}
		}
	} else {
		scheduledToday = true
		lastScheduledDelta = 1
	}

	y, m, d := now.Date()
	location := now.Location()
	if ! scheduledToday {
		d -= lastScheduledDelta
	}

	startTime, err := time.ParseInLocation("15:04", s.Start, location)
	if err != nil {
		return 0, err
	}
	lastStart := time.Date(y, m, d, startTime.Hour(), startTime.Minute(), 0, 0, location)
	if scheduledToday && lastStart.After(now) {
		lastStart = lastStart.AddDate(0, 0, -lastScheduledDelta)
	}

	var lastEnd time.Time
	if s.End != "" {
		endTime, err := time.ParseInLocation("15:04", s.End, location)
		if err != nil {
			return 0, err
		}
		lastEnd = time.Date(y, m, d, endTime.Hour(), endTime.Minute(), 0, 0, location)
		if scheduledToday && lastEnd.After(now) {
			lastEnd = lastEnd.AddDate(0, 0, -lastScheduledDelta)
		}
	}

	y, m, d = lastAdd.Date()
	d += s.Interval
	intervalMin := time.Date(y, m, d, 0, 0, 0, 0, location)

	if lastEnd.After(lastStart) && lastUpdate.Before(lastEnd) {
		return -1, nil
	} else if lastStart.After(lastEnd) && lastUpdate.Before(lastStart) && now.After(intervalMin) {
		return 1, nil
	} else {
		return 0, nil
	}
}

/*
 * returns 1 for add, -1 for delete, 0 for no action
 */
func (t Template) Action(lastAdd time.Time, lastUpdate time.Time, now time.Time) (int, error) {
	return t.Schedule.Action(lastAdd, lastUpdate, now)
}
