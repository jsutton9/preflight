package checklist

import (
	"github.com/jsutton9/preflight/api/errors"
	"github.com/jsutton9/preflight/clients/trello"
	"time"
)

type Checklist struct {
	TasksSource string     `json:"tasksSource"`
	TasksTarget string     `json:"tasksTarget"`
	IsScheduled bool       `json:"isScheduled"`
	Tasks []string         `json:"tasks,omitempty"`
	Trello *trello.ListKey `json:"trello,omitempy"`
	Schedule *Schedule     `json:"schedule,omitempty"`
	Record *UpdateRecord   `json:"updateRecord"`
}

type Schedule struct {
	Interval int     `json:"interval,omitempty"`
	Days []string    `json:"days,omitempty"`
	Start string     `json:"start"`
	End string       `json:"end,omitempty"`
}

type UpdateRecord struct {
	Ids []int           `json:"ids"`
	Time time.Time      `json:"time"`
	AddTime time.Time   `json:"addTime"`
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

func (s *Schedule) Action(lastAdd time.Time, lastUpdate time.Time, now time.Time) (int, time.Time, *errors.PreflightError) {
	if s == nil {
		return 0, lastUpdate, nil
	}

	var scheduledToday bool
	lastScheduledDelta := 7
	if s.Days != nil && len(s.Days) > 0 {
		scheduledToday = false
		currentWeekday := now.Weekday()
		for _, weekdayString := range s.Days {
			weekday, err := parseWeekday(weekdayString)
			if err != nil {
				return 0, lastUpdate, err.Prepend("checklist.Schedule.Action: error parsing weekday: ")
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
		return 0, lastUpdate, &errors.PreflightError{
			Status: 422,
			InternalMessage: "checklist.Schedule.Action: error parsing start time " +
				"\"" + s.Start + "\": \n\t" + err.Error(),
			ExternalMessage: "Unable to parse start time \"" + s.Start + "\"; should be like \"15:04\"",
		}
	}
	lastStart := time.Date(y, m, d, startTime.Hour(), startTime.Minute(), 0, 0, location)
	if scheduledToday && lastStart.After(now) {
		lastStart = lastStart.AddDate(0, 0, -lastScheduledDelta)
	}

	var lastEnd time.Time
	if s.End != "" {
		endTime, err := time.ParseInLocation("15:04", s.End, location)
		if err != nil {
			return 0, lastUpdate, &errors.PreflightError{
				Status: 422,
				InternalMessage: "checklist.Schedule.Action: error parsing end time " +
					"\"" + s.Start + "\": \n\t" + err.Error(),
				ExternalMessage: "Unable to parse end time \"" + s.End + "\"; should be like \"15:04\"",
			}
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
		return -1, lastEnd, nil
	} else if lastStart.After(lastEnd) && lastUpdate.Before(lastStart) && now.After(intervalMin) {
		return 1, lastStart, nil
	} else {
		return 0, lastUpdate, nil
	}
}

/*
 * returns 1 for add, -1 for delete, 0 for no action
 */
func (c Checklist) Action(lastAdd time.Time, lastUpdate time.Time, now time.Time) (int, time.Time, *errors.PreflightError) {
	action, updateTime, err := c.Schedule.Action(lastAdd, lastUpdate, now)
	if err != nil {
		err.Prepend("checklist.Checklist.Action: error: ")
	}
	return action, updateTime, err
}
