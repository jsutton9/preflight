package config

import (
	"testing"
	"time"
)

func positiveActionTest(test *testing.T, t Template, last time.Time,
		now time.Time, correctAction Action) {
}

func negativeActionTest(test *testing.T, t Template, last time.Time,
		now time.Time, incorrectAction Action) {
}

func TestConfig(test *testing.T) {
	conf, err := New("test_config.json")
	if err != nil {
		test.Fatal(err)
	}

	daily := conf.Templates["daily"]
	dailyEnd := conf.Templates["daily_end"]
	weekdays := conf.Templates["weekdays"]
	weekdaysEnd := conf.Templates["weekdays_end"]

	format := "2006-01-02 15:04:05 MST"
	mondayMorning  := time.Parse(format, "2016-04-04 01:00:00 MST")
	mondayNoon     := time.Parse(format, "2016-04-04 12:00:00 MST")
	mondayEvening  := time.Parse(format, "2016-04-04 23:00:00 MST")
	tuesdayMorning := time.Parse(format, "2016-04-05 01:00:00 MST")
	tuesdayNoon    := time.Parse(format, "2016-04-05 12:00:00 MST")
	tuesdayEvening := time.Parse(format, "2016-04-05 23:00:00 MST")
	wednesdayMorning := time.Parse(format, "2016-04-06 01:00:00 MST")
	wednesdayEvening := time.Parse(format, "2016-04-06 23:00:00 MST")

	positiveActionTest(test, daily, mondayMorning, mondayMorning, 0)
	positiveActionTest(test, daily, mondayMorning, mondayNoon, 1)
	positiveActionTest(test, daily, mondayMorning, tuesdayMorning, 1)
	positiveActionTest(test, daily, mondayEvening, mondayEvening, 0)

	positiveActionTest(test, dailyEnd, mondayMorning, mondayMorning, 0)
	positiveActionTest(test, dailyEnd, mondayMorning, mondayNoon, 1)
	negativeActionTest(test, dailyEnd, mondayMorning, mondayEvening, 1)
	positiveActionTest(test, dailyEnd, mondayMorning, tuesdayNoon, 1)
	positiveActionTest(test, dailyEnd, mondayMorning, tuesdayEvening, -1)
	positiveActionTest(test, dailyEnd, mondayNoon, mondayNoon, 0)
	positiveActionTest(test, dailyEnd, mondayNoon, mondayEvening, -1)
	positiveActionTest(test, dailyEnd, mondayNoon, tuesdayMorning, -1)
	positiveActionTest(test, dailyEnd, mondayNoon, tuesdayEvening, -1)
	positiveActionTest(test, dailyEnd, mondayEvening, mondayEvening, 0)
	positiveActionTest(test, dailyEnd, mondayEvening, tuesdayNoon, 1)
	negativeActionTest(test, dailyEnd, mondayEvening, tuesdayEvening, 1)

	positiveActionTest(test, weekdays, mondayMorning, mondayMorning, 0)
	positiveActionTest(test, weekdays, mondayMorning, mondayNoon, 1)
	positiveActionTest(test, weekdays, mondayMorning, tuesdayMorning, 1)
	positiveActionTest(test, weekdays, mondayMorning, tuesdayEvening, 1)
	positiveActionTest(test, weekdays, mondayEvening, mondayEvening, 0)
	positiveActionTest(test, weekdays, mondayEvening, tuesdayMorning, 0)
	positiveActionTest(test, weekdays, mondayEvening, tuesdayEvening, 0)
	positiveActionTest(test, weekdays, tuesdayMorning, wednesdayMorning, 0)
	positiveActionTest(test, weekdays, tuesdayMorning, wednesdayEvening, 1)
	positiveActionTest(test, weekdays, tuesdayEvening, wednesdayEvening, 1)

	positiveActionTest(test, weekdaysEnd, mondayMorning, mondayMorning, 0)
