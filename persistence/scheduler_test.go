package persistence

import (
	"github.com/jsutton9/preflight/checklist"
	"github.com/jsutton9/preflight/user"
	"math/rand"
	"testing"
	"time"
)

func TestQueueing(t *testing.T) {
	q := NewQueue()
	now := time.Now()
	before := now.AddDate(-1, 0, 0)
	after := now.AddDate(1, 0, 0)
	cl := &checklist.Checklist{}

	t.Log("starting first insertion")
	for i:=0; i<60; i++ {
		offset := time.Duration(rand.Intn(100))*time.Hour
		t := now.Add(offset)
		q.insert(&UpdateJob{Time: t, Checklist: cl})
	}

	resultBefore, err := q.Pop(&before)
	if err != nil {
		t.Fatal(err)
	}
	if resultBefore != nil {
		t.Log("fail on popping before time: expected nil, got job")
		t.Fail()
	}

	t.Log("starting first popping")
	var prev *UpdateJob
	for i:=0; i<30; i++ {
		job, err := q.Pop(&after)
		if err != nil {
			t.Fatal(err)
		}
		if job == nil {
			t.Log("fail on popping: expected job, got nil")
			t.Fail()
		} else if prev!=nil && job.Time.Before(prev.Time) {
			t.Log("fail on queueing: jobs popped out of order")
			t.Fail()
		}
		prev = job
	}

	t.Log("starting second insertion")
	for i:=0; i<30; i++ {
		offset := time.Duration(rand.Intn(100))*time.Hour
		t := now.Add(offset)
		q.insert(&UpdateJob{Time: t, Checklist: cl})
	}

	t.Log("starting second popping")
	prev = nil
	for i:=0; i<60; i++ {
		job, err := q.Pop(&after)
		if err != nil {
			t.Fatal(err)
		}
		if job == nil {
			t.Log("fail on popping: expected job, got nil")
			t.Fail()
		} else if prev!=nil && job.Time.Before(prev.Time) {
			t.Log("fail on queueing: jobs popped out of order")
			t.Fail()
		}
		prev = job
	}

	jobAfter, err := q.Pop(&after)
	if err != nil {
		t.Fatal(err)
	}
	if jobAfter != nil {
		t.Log("fail on popping after emptying: expected nil, got job")
		t.Fail()
	}
}

func TestChecklistChange(t *testing.T) {
	u, pErr := user.New("a@b.c", "pass")
	if pErr != nil {
		t.Fatal(pErr)
	}

	clBefore := &checklist.Checklist{
		Name: "before",
		IsScheduled: true,
		Schedule: &checklist.Schedule{
			Days: []string{"mon"},
			Start: "12:00",
		},
	}
	pErr = clBefore.GenId()
	if pErr != nil {
		t.Fatal(pErr)
	}
	clAfter := &checklist.Checklist{
		Id: clBefore.Id,
		Name: "after",
		IsScheduled: true,
		Schedule: &checklist.Schedule{
			Days: []string{"tues"},
			Start: "12:00",
		},
	}

	format := "2006-01-02 15:04:05"
	location, err := time.LoadLocation("America/Denver")
	if err != nil {
		t.Fatal(err)
	}
	mondayMorning, err := time.ParseInLocation(format, "2016-09-26 01:00:00", location)
	if err != nil {
		t.Fatal(err)
	}
	mondayNight, err := time.ParseInLocation(format, "2016-09-26 23:00:00", location)
	if err != nil {
		t.Fatal(err)
	}
	tuesdayNight := mondayNight.AddDate(0, 0, 1)
	nextMondayNight := mondayNight.AddDate(0, 0, 7)

	q := NewQueue()
	q.SetChecklist(u, clBefore, &mondayMorning)
	popBefore, pErr := q.Pop(&mondayNight)
	if pErr != nil {
		t.Fatal(pErr)
	}
	q.SetChecklist(u, clAfter, &mondayNight)
	popAfter, pErr := q.Pop(&tuesdayNight)
	if pErr != nil {
		t.Fatal(pErr)
	}
	popNextMonday, pErr := q.Pop(&nextMondayNight)
	if pErr != nil {
		t.Fatal(pErr)
	}

	if popBefore == nil {
		t.Log("failed on pop before change: expected \"before\", got nil")
		t.Fail()
	}
	if popAfter == nil {
		t.Log("failed on pop after change: expected \"after\", got nil")
		t.Fail()
	} else if popAfter.Checklist.Name != "after" {
		t.Logf("wrong checklist popped after change: expected \"after\", got \"%s\"\n",
			popAfter.Checklist.Name)
		t.Fail()
	}
	if popNextMonday != nil {
		t.Logf("failed on pop following week: expected nil, got \"%s\"\n",
			popNextMonday.Checklist.Name)
		t.Fail()
	}
}

func TestUserChange(t *testing.T) {
	emailBefore := "before@preflight.com"
	emailAfter := "after@preflight.com"
	userBefore, pErr := user.New(emailBefore, "pass")
	if pErr != nil {
		t.Fatal(pErr)
	}
	userAfter := &(*userBefore)
	userAfter.Email = emailAfter

	cl := &checklist.Checklist{
		Name: "foo",
		IsScheduled: true,
		Schedule: &checklist.Schedule{
			Start: "9:00",
		},
	}
	userBefore.Checklists = append(userBefore.Checklists, cl)
	userAfter.Checklists = append(userAfter.Checklists, cl)

	format := "2006-01-02 15:04:05"
	location, err := time.LoadLocation("America/Denver")
	if err != nil {
		t.Fatal(err)
	}
	monday, err := time.ParseInLocation(format, "2016-09-26 12:00:00", location)
	if err != nil {
		t.Fatal(err)
	}
	tuesday := monday.AddDate(0, 0, 1)
	wednesday := monday.AddDate(0, 0, 2)

	q := NewQueue()
	q.SetUser(userBefore, &monday)
	q.SetChecklist(userBefore, cl, &monday)
	jobBefore, pErr := q.Pop(&tuesday)
	if pErr != nil {
		t.Fatal(pErr)
	}
	q.SetUser(userAfter, &tuesday)
	jobAfter, pErr := q.Pop(&wednesday)
	if pErr != nil {
		t.Fatal(pErr)
	}

	if jobBefore == nil {
		t.Log("failure on pop before change: got nil, expected \"before\"")
		t.Fail()
	} else if jobBefore.User == nil {
		t.Log("failure on pop before change: got job with nil user, expected \"before\"")
		t.Fail()
	}
	if jobAfter == nil {
		t.Log("failure on pop after change: got nil, expected \"after\"")
		t.Fail()
	} else if jobAfter.User == nil {
		t.Log("failure on pop after change: got job with nil user, expected \"after\"")
		t.Fail()
	} else if jobAfter.User.Email != emailAfter {
		t.Logf("failure on pop after change: got job with wrong user, expected \"after\", " +
			"got \"%s\"\n", jobAfter.User.Email)
		t.Fail()
	}
}
