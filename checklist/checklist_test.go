package checklist

import (
	"github.com/jsutton9/preflight/api/errors"
	"testing"
	"time"
)

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
