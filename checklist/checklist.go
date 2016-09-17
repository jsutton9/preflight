package checklist

import (
	"github.com/jsutton9/preflight/api/errors"
	"github.com/jsutton9/preflight/clients/trello"
	"os/exec"
	"time"
)

type Checklist struct {
	Id string              `json:"id"`
	TasksSource string     `json:"tasksSource"`
	TasksTarget string     `json:"tasksTarget"`
	IsScheduled bool       `json:"isScheduled"`
	Tasks []string         `json:"tasks,omitempty"`
	Trello *trello.ListKey `json:"trello,omitempy"`
	Schedule *Schedule     `json:"schedule,omitempty"`
}

type Schedule struct {
	Interval int     `json:"interval,omitempty"`
	Days []string    `json:"days,omitempty"`
	Start string     `json:"start"`
	End string       `json:"end,omitempty"`
}

func parseWeekday(s string) (time.Weekday, *errors.PreflightError) {
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
		return time.Sunday, &errors.PreflightError{
			Status: 422,
			InternalMessage: "checklist.parseWeekday: unable to parse \"" + s + "\"",
			ExternalMessage: "day of week \"" + s + "\" not understood",
		}
	}
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

func (s *Schedule) NextAdd(now time.Time) (time.Time, *errors.PreflightError) {
	var scheduledToday bool
	daysDelta := 1
	if s.Interval > 0 {
		daysDelta = s.Interval
	}
	if s.Days != nil && len(s.Days) > 0 {
		minWeekday := (int(now.Weekday()) + daysDelta) % 7
		scheduledToday = false
		bestMinusMin := 8
		for _, weekdayString := range s.Days {
			weekday, err := parseWeekday(weekdayString)
			if err != nil {
				return now, err.Prepend("checklist.Schedule.NextAdd: error parsing weekday: ")
			}
			weekdayDelta := (int(weekday) - minWeekday) % 7
			if weekdayDelta < 0 {
				weekdayDelta += 7
			}
			if weekdayDelta < bestMinusMin {
				bestMinusMin = weekdayDelta
			}
			if weekday == now.Weekday() {
				scheduledToday = true
			}
		}
		daysDelta += bestMinusMin
	} else {
		scheduledToday = true
	}

	y, m, d := now.Date()
	location := now.Location()
	startTime, err := time.ParseInLocation("15:04", s.Start, location)
	if err != nil {
		return now, &errors.PreflightError{
			Status: 422,
			InternalMessage: "checklist.Schedule.NextAdd: error parsing start time " +
				"\"" + s.Start + "\": \n\t" + err.Error(),
			ExternalMessage: "Unable to parse start time \"" + s.Start + "\"; should be like \"15:04\"",
		}
	}
	nextAdd := time.Date(y, m, d+daysDelta, startTime.Hour(), startTime.Minute(), 0, 0, location)
	if scheduledToday && s.Interval == 0 {
		addToday := time.Date(y, m, d, startTime.Hour(), startTime.Minute(), 0, 0, location)
		if addToday.After(now) {
			nextAdd = addToday
		}
	}

	return nextAdd, nil
}

func (s *Schedule) NextRemove(now time.Time) (time.Time, *errors.PreflightError) {
	y, m, d := now.Date()
	location := now.Location()

	endTime, err := time.ParseInLocation("15:04", s.End, location)
	if err != nil {
		return now, &errors.PreflightError{
			Status: 422,
			InternalMessage: "checklist.Schedule.NextRemove: error parsing end time " +
				"\"" + s.End + "\": \n\t" + err.Error(),
			ExternalMessage: "Unable to parse end time \"" + s.End + "\"; should be like \"15:04\"",
		}
	}

	removeTime := time.Date(y, m, d, endTime.Hour(), endTime.Minute(), 0, 0, location)
	if ! removeTime.After(now) {
		removeTime = removeTime.AddDate(0, 0, 1)
	}

	return removeTime, nil
}
