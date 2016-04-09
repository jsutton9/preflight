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
	//Interval int     `json:"interval,omitempty"`
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

func (t Template) Action(lastUpdate time.Time, now time.Time) int, error {
	return t.Schedule.Action(lastUpdate, now)
}

/*func (s Schedule) targetState(t time.Time) bool, error {
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
}*/

func parseWeekday(s String) Weekday, error {
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
		return nil, errors.New(
			"Unable to parse day of the week \"%s\"", s)
	}
}

//TODO: every n days (schedule.Interval)
func (s Schedule) Action(lastUpdate time.Time, now time.Time) int, error {
	if s == nil {
		return 0, nil
	}

	var scheduledToday bool
	minWeekdayDelta := 7
	if s.Days != nil {
		scheduledToday = false
		currentWeekday := t.Weekday()
		for _, weekdayString := range s.Days {
			weekday, err := parseWeekday(weekdayString)
			if err != nil {
				return _, err
			}
			weekdayDelta := (currentWeekday-weekday)%7
			if weekdayDelta == 0 {
				scheduledToday = true
			} else if weekdayDelta < minWeekdayDelta {
				minWeekdayDelta = weekdayDelta
			}
		}
	} else {
		scheduledToday = true
		minWeekdayDelta = 1
	}

	y, m, d := now.getDate()

	lastStart := time.Parse("15:04 MST", s.Start)
	lastStart.AddDate(y, m, d)
	if lastStart.After(now) {
		lastStart.AddDate(0, 0, -minWeekdayDelta)
	}

	var lastEnd time.Time
	if s.End == nil {
		lastEnd = nil
	} else {
		lastEnd = time.Parse("15:04 MST", s.End)
		lastEnd.AddDate(y, m, d)
		if lastEnd.After(now) {
			lastEnd.addDate(0, 0, -minWeekdayDelta)
		}
	}

	if lastEnd != nil && lastEnd.After(lastStart) {
		if lastUpdate != nil && lastUpdate.Before(lastEnd) {
			return -1, nil
		} else {
			return 0, nil
		}
	} else {
		if lastUpdate == nil || lastUpdate.Before(lastStart) {
			return 1, nil
		} else {
			return 0, nil
		}
	}
}
