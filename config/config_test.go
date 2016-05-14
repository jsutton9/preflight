package config

import (
	"testing"
	"time"
)

func actionTest(test *testing.T, t Template, lastAdd time.Time, last time.Time,
		now time.Time, correctAction int) {
	action, _, err := t.Action(lastAdd, last, now)
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

func negativeActionTest(test *testing.T, t Template, lastAdd time.Time, last time.Time,
		now time.Time, incorrectAction int) {
	action, _, err := t.Action(lastAdd, last, now)
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

func TestConfigScheduling(test *testing.T) {
	conf, err := New("test_config.json")
	if err != nil {
		test.Fatal(err)
	}

	daily := conf.Templates["daily"]
	dailyEnd := conf.Templates["daily_end"]
	weekdays := conf.Templates["weekdays"]
	weekdaysEnd := conf.Templates["weekdays_end"]
	intervalWeekdays := conf.Templates["interval_weekdays"]

	format := "2006-01-02 15:04:05"
	location, err := time.LoadLocation(conf.Timezone)
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
