package config

import (
	"testing"
	"time"
)

func actionTest(test *testing.T, t Template, last time.Time,
		now time.Time, correctAction int) {
	action, err := t.Action(last, now)
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

func negativeActionTest(test *testing.T, t Template, last time.Time,
		now time.Time, incorrectAction int) {
	action, err := t.Action(last, now)
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

	format := "2006-01-02 15:04:05"
	location, err := time.LoadLocation(conf.Timezone)
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
	tuesdayMorning, err   := time.ParseInLocation(format, "2016-04-05 01:00:00", location)
	if err != nil {
		test.Fatal(err)
	}
	tuesdayNoon, err      := time.ParseInLocation(format, "2016-04-05 12:00:00", location)
	if err != nil {
		test.Fatal(err)
	}
	tuesdayEvening, err   := time.ParseInLocation(format, "2016-04-05 23:00:00", location)
	if err != nil {
		test.Fatal(err)
	}
	wednesdayMorning, err := time.ParseInLocation(format, "2016-04-06 01:00:00", location)
	if err != nil {
		test.Fatal(err)
	}
	wednesdayNoon, err    := time.ParseInLocation(format, "2016-04-06 12:00:00", location)
	if err != nil {
		test.Fatal(err)
	}
	wednesdayEvening, err := time.ParseInLocation(format, "2016-04-06 23:00:00", location)
	if err != nil {
		test.Fatal(err)
	}

	test.Log("testing daily")
	actionTest(test, daily, mondayMorning, mondayMorning, 0)
	actionTest(test, daily, mondayMorning, mondayNoon, 1)
	actionTest(test, daily, mondayMorning, tuesdayMorning, 1)
	actionTest(test, daily, mondayEvening, mondayEvening, 0)
	test.Log("")

	test.Log("testing dailyEnd")
	actionTest(test, dailyEnd, mondayMorning, mondayMorning, 0)
	actionTest(test, dailyEnd, mondayMorning, mondayNoon, 1)
	negativeActionTest(test, dailyEnd, mondayMorning, mondayEvening, 1)
	actionTest(test, dailyEnd, mondayMorning, tuesdayNoon, 1)
	actionTest(test, dailyEnd, mondayMorning, tuesdayEvening, -1)
	actionTest(test, dailyEnd, mondayNoon, mondayNoon, 0)
	actionTest(test, dailyEnd, mondayNoon, mondayEvening, -1)
	actionTest(test, dailyEnd, mondayNoon, tuesdayMorning, -1)
	actionTest(test, dailyEnd, mondayNoon, tuesdayEvening, -1)
	actionTest(test, dailyEnd, mondayEvening, mondayEvening, 0)
	actionTest(test, dailyEnd, mondayEvening, tuesdayNoon, 1)
	negativeActionTest(test, dailyEnd, mondayEvening, tuesdayEvening, 1)
	test.Log("")

	test.Log("testing weekdays")
	actionTest(test, weekdays, mondayMorning, mondayMorning, 0)
	actionTest(test, weekdays, mondayMorning, mondayNoon, 1)
	actionTest(test, weekdays, mondayMorning, tuesdayMorning, 1)
	actionTest(test, weekdays, mondayMorning, tuesdayEvening, 1)
	actionTest(test, weekdays, mondayEvening, mondayEvening, 0)
	actionTest(test, weekdays, mondayEvening, tuesdayMorning, 0)
	actionTest(test, weekdays, mondayEvening, tuesdayEvening, 0)
	actionTest(test, weekdays, tuesdayMorning, wednesdayMorning, 0)
	actionTest(test, weekdays, tuesdayMorning, wednesdayEvening, 1)
	actionTest(test, weekdays, tuesdayEvening, wednesdayEvening, 1)
	test.Log("")

	test.Log("testing weekdaysEnd")
	actionTest(test, weekdaysEnd, mondayMorning, mondayMorning, 0)
	actionTest(test, weekdaysEnd, mondayMorning, mondayNoon, 1)
	negativeActionTest(test, weekdaysEnd, mondayMorning, tuesdayMorning, 1)
	negativeActionTest(test, weekdaysEnd, mondayMorning, tuesdayEvening, 1)
	actionTest(test, weekdaysEnd, mondayNoon, mondayNoon, 0)
	actionTest(test, weekdaysEnd, mondayNoon, mondayNoon, 0)
	actionTest(test, weekdaysEnd, mondayNoon, mondayEvening, -1)
	actionTest(test, weekdaysEnd, mondayNoon, tuesdayNoon, -1)
	actionTest(test, weekdaysEnd, mondayEvening, mondayEvening, 0)
	actionTest(test, weekdaysEnd, mondayEvening, tuesdayNoon, 0)
	actionTest(test, weekdaysEnd, mondayEvening, wednesdayMorning, 0)
	actionTest(test, weekdaysEnd, mondayEvening, wednesdayNoon, 1)
	actionTest(test, weekdaysEnd, tuesdayNoon, tuesdayNoon, 0)
	actionTest(test, weekdaysEnd, tuesdayNoon, wednesdayMorning, 0)
	actionTest(test, weekdaysEnd, tuesdayNoon, wednesdayNoon, 1)
	negativeActionTest(test, weekdaysEnd, tuesdayNoon, wednesdayEvening, 1)
	test.Log("")
}
