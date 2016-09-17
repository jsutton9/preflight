package checklist

import (
	"github.com/jsutton9/preflight/api/errors"
	"testing"
	"time"
)

func actionTest(test *testing.T, s Schedule, lastAdd time.Time, last time.Time,
		now time.Time, correctAction int) {
	action, _, err := s.Action(lastAdd, last, now)
	if err != nil {
		test.Error(err)
	} else if action != correctAction {
		test.Log("test failure: ")
		test.Logf("\tlast: %s\n", last.String())
		test.Logf("\tnow: %s\n", now.String())
		test.Logf("\texpected %d, got %d\n", correctAction, action)
		test.Fail()
	}
}

func negativeActionTest(test *testing.T, s Schedule, lastAdd time.Time, last time.Time,
		now time.Time, incorrectAction int) {
	action, _, err := s.Action(lastAdd, last, now)
	if err != nil {
		test.Error(err)
	} else if action == incorrectAction {
		test.Log("negative test failure: ")
		test.Logf("\tlast: %s\n", last.String())
		test.Logf("\tnow: %s\n", now.String())
		test.Logf("\taction: %d\n", action)
		test.Fail()
	}
}

func nextAddRemoveTest(test *testing.T, s Schedule, now time.Time, expected time.Time, remove bool) {
	var result time.Time
	var err *errors.PreflightError
	var name string
	if remove {
		result, err = s.NextRemove(now)
		name = "NextRemove"
	} else {
		result, err = s.NextAdd(now)
		name = "NextAdd"
	}
	if err != nil {
		test.Error(err)
	} else if ! result.Equal(expected) {
		test.Logf("  %s failure: ", name)
		test.Logf("    now: %s\n", now.String())
		test.Logf("    expected: %s\n", expected.String())
		test.Logf("    result: %s\n", result.String())
		test.Fail()
	}
}

func TestConfigScheduling(test *testing.T) {
	daily := Schedule{
		Start: "9:00",
	}
	dailyEnd := Schedule{
		Start: "9:00",
		End: "17:00",
	}
	weekdays := Schedule{
		Days: []string{"Monday", "Wednesday", "Friday"},
		Start: "9:00",
	}
	weekdaysEnd := Schedule{
		Days: []string{"Monday", "Wednesday", "Friday"},
		Start: "9:00",
		End: "17:00",
	}
	intervalWeekdays := Schedule{
		Days: []string{"Monday", "Wednesday", "Friday"},
		Interval: 3,
		Start: "9:00",
	}

	format := "2006-01-02 15:04:05"
	location, err := time.LoadLocation("America/Denver")
	if err != nil {
		test.Fatal(err)
	}
	never := time.Date(0, 0, 0, 0, 0, 0, 0, location)
	mondayMorning, err    := time.ParseInLocation(format, "2016-04-04 01:00:00", location)
	if err != nil {
		test.Fatal(err)
	}
	mondayNoon, err       := time.ParseInLocation(format, "2016-04-04 12:00:00", location)
	if err != nil {
		test.Fatal(err)
	}
	mondayEvening, err    := time.ParseInLocation(format, "2016-04-04 23:00:00", location)
	if err != nil {
		test.Fatal(err)
	}

	tuesdayMorning := mondayMorning.AddDate(0, 0, 1)
	tuesdayNoon := mondayNoon.AddDate(0, 0, 1)
	tuesdayEvening := mondayEvening.AddDate(0, 0, 1)

	wednesdayMorning := tuesdayMorning.AddDate(0, 0, 1)
	wednesdayNoon := tuesdayNoon.AddDate(0, 0, 1)
	wednesdayEvening := tuesdayEvening.AddDate(0, 0, 1)

	thursdayMorning := wednesdayMorning.AddDate(0, 0, 1)
	thursdayNoon := wednesdayNoon.AddDate(0, 0, 1)

	fridayMorning := thursdayMorning.AddDate(0, 0, 1)
	fridayNoon := thursdayNoon.AddDate(0, 0, 1)

	test.Log("testing daily")
	actionTest(test, daily, never, mondayMorning, mondayMorning, 0)
	actionTest(test, daily, never, mondayMorning, mondayNoon, 1)
	actionTest(test, daily, never, mondayMorning, tuesdayMorning, 1)
	actionTest(test, daily, never, mondayEvening, mondayEvening, 0)
	test.Log("")

	test.Log("testing dailyEnd")
	actionTest(test, dailyEnd, never, mondayMorning, mondayMorning, 0)
	actionTest(test, dailyEnd, never, mondayMorning, mondayNoon, 1)
	negativeActionTest(test, dailyEnd, never, mondayMorning, mondayEvening, 1)
	actionTest(test, dailyEnd, never, mondayMorning, tuesdayNoon, 1)
	actionTest(test, dailyEnd, never, mondayMorning, tuesdayEvening, -1)
	actionTest(test, dailyEnd, never, mondayNoon, mondayNoon, 0)
	actionTest(test, dailyEnd, never, mondayNoon, mondayEvening, -1)
	actionTest(test, dailyEnd, never, mondayNoon, tuesdayMorning, -1)
	actionTest(test, dailyEnd, never, mondayNoon, tuesdayEvening, -1)
	actionTest(test, dailyEnd, never, mondayEvening, mondayEvening, 0)
	actionTest(test, dailyEnd, never, mondayEvening, tuesdayNoon, 1)
	negativeActionTest(test, dailyEnd, never, mondayEvening, tuesdayEvening, 1)
	test.Log("")

	test.Log("testing weekdays")
	actionTest(test, weekdays, never, mondayMorning, mondayMorning, 0)
	actionTest(test, weekdays, never, mondayMorning, mondayNoon, 1)
	actionTest(test, weekdays, never, mondayMorning, tuesdayMorning, 1)
	actionTest(test, weekdays, never, mondayMorning, tuesdayEvening, 1)
	actionTest(test, weekdays, never, mondayEvening, mondayEvening, 0)
	actionTest(test, weekdays, never, mondayEvening, tuesdayMorning, 0)
	actionTest(test, weekdays, never, mondayEvening, tuesdayEvening, 0)
	actionTest(test, weekdays, never, tuesdayMorning, wednesdayMorning, 0)
	actionTest(test, weekdays, never, tuesdayMorning, wednesdayEvening, 1)
	actionTest(test, weekdays, never, tuesdayEvening, wednesdayEvening, 1)
	test.Log("")

	test.Log("testing weekdaysEnd")
	actionTest(test, weekdaysEnd, never, mondayMorning, mondayMorning, 0)
	actionTest(test, weekdaysEnd, never, mondayMorning, mondayNoon, 1)
	negativeActionTest(test, weekdaysEnd, never, mondayMorning, tuesdayMorning, 1)
	negativeActionTest(test, weekdaysEnd, never, mondayMorning, tuesdayEvening, 1)
	actionTest(test, weekdaysEnd, never, mondayNoon, mondayNoon, 0)
	actionTest(test, weekdaysEnd, never, mondayNoon, mondayNoon, 0)
	actionTest(test, weekdaysEnd, never, mondayNoon, mondayEvening, -1)
	actionTest(test, weekdaysEnd, never, mondayNoon, tuesdayNoon, -1)
	actionTest(test, weekdaysEnd, never, mondayEvening, mondayEvening, 0)
	actionTest(test, weekdaysEnd, never, mondayEvening, tuesdayNoon, 0)
	actionTest(test, weekdaysEnd, never, mondayEvening, wednesdayMorning, 0)
	actionTest(test, weekdaysEnd, never, mondayEvening, wednesdayNoon, 1)
	actionTest(test, weekdaysEnd, never, tuesdayNoon, tuesdayNoon, 0)
	actionTest(test, weekdaysEnd, never, tuesdayNoon, wednesdayMorning, 0)
	actionTest(test, weekdaysEnd, never, tuesdayNoon, wednesdayNoon, 1)
	negativeActionTest(test, weekdaysEnd, never, tuesdayNoon, wednesdayEvening, 1)
	test.Log("")

	test.Log("testing intervalWeekdays")
	actionTest(test, intervalWeekdays, mondayNoon, wednesdayMorning, wednesdayNoon, 0)
	actionTest(test, intervalWeekdays, mondayNoon, thursdayMorning, thursdayNoon, 0)
	actionTest(test, intervalWeekdays, mondayNoon, fridayMorning, fridayNoon, 1)
	test.Log("")
}

func TestNextAdd(test *testing.T) {
	daily := Schedule{
		Start: "12:00",
	}
	weekdays := Schedule{
		Days: []string{"Monday", "Wednesday"},
		Start: "12:00",
	}
	intervalWeekdays := Schedule{
		Days: []string{"Monday", "Tuesday", "Thursday"},
		Interval: 3,
		Start: "12:00",
	}
	longInterval := Schedule{
		Interval: 100,
		Start: "12:00",
	}

	format := "2006-01-02 15:04:05"
	location, err := time.LoadLocation("America/Denver")
	if err != nil {
		test.Fatal(err)
	}
	mondayMorning, err    := time.ParseInLocation(format, "2016-04-04 01:00:00", location)
	if err != nil {
		test.Fatal(err)
	}
	mondayNoon, err       := time.ParseInLocation(format, "2016-04-04 12:00:00", location)
	if err != nil {
		test.Fatal(err)
	}
	mondayEvening, err    := time.ParseInLocation(format, "2016-04-04 23:00:00", location)
	if err != nil {
		test.Fatal(err)
	}

	tuesdayMorning := mondayMorning.AddDate(0, 0, 1)
	tuesdayNoon := mondayNoon.AddDate(0, 0, 1)
	wednesdayNoon := mondayNoon.AddDate(0, 0, 2)
	thursdayNoon := mondayNoon.AddDate(0, 0, 3)
	nextMondayNoon := mondayNoon.AddDate(0, 0, 7)
	hundredDaysLater := mondayNoon.AddDate(0, 0, 100)

	test.Log("testing daily")
	nextAddRemoveTest(test, daily, mondayMorning, mondayNoon, false)
	nextAddRemoveTest(test, daily, mondayEvening, tuesdayNoon, false)

	test.Log("testing weekdays")
	nextAddRemoveTest(test, weekdays, mondayMorning, mondayNoon, false)
	nextAddRemoveTest(test, weekdays, mondayEvening, wednesdayNoon, false)
	nextAddRemoveTest(test, weekdays, tuesdayMorning, wednesdayNoon, false)

	test.Log("testing interval and weekdays")
	nextAddRemoveTest(test, intervalWeekdays, mondayMorning, thursdayNoon, false)
	nextAddRemoveTest(test, intervalWeekdays, tuesdayMorning, nextMondayNoon, false)

	test.Log("testing long interval")
	nextAddRemoveTest(test, longInterval, mondayMorning, hundredDaysLater, false)
}

func TestNextRemove(test *testing.T) {
	endNoon := Schedule{
		End: "12:00",
	}

	format := "2006-01-02 15:04:05"
	location, err := time.LoadLocation("America/Denver")
	if err != nil {
		test.Fatal(err)
	}
	mondayMorning, err    := time.ParseInLocation(format, "2016-04-04 01:00:00", location)
	if err != nil {
		test.Fatal(err)
	}
	mondayNoon, err       := time.ParseInLocation(format, "2016-04-04 12:00:00", location)
	if err != nil {
		test.Fatal(err)
	}
	tuesdayNoon := mondayNoon.AddDate(0, 0, 1)

	test.Log("testing end at noon")
	nextAddRemoveTest(test, endNoon, mondayMorning, mondayNoon, true)
	nextAddRemoveTest(test, endNoon, mondayNoon, tuesdayNoon, true)
}

func TestId(test *testing.T) {
	c := Checklist{}
	c.GenId()
	if len(c.Id) == 0 {
		test.Log("id test failed: Id is empty after GenId()")
		test.Fail()
	}
	d := Checklist{}
	d.GenId()
	if c.Id == d.Id {
		test.Log("id test failed: Two checklists generated the same id: "+c.Id)
		test.Fail()
	}
}
