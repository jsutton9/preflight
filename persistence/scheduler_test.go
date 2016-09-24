package persistence

import (
	"math/rand"
	"testing"
	"time"
)

func TestQueueing(t *testing.T) {
	q := NewQueue()
	now := time.Now()
	before := now.AddDate(-1, 0, 0)
	after := now.AddDate(1, 0, 0)

	t.Log("starting first insertion")
	for i:=0; i<60; i++ {
		offset := time.Duration(rand.Intn(100))*time.Hour
		t := now.Add(offset)
		q.insert(&UpdateJob{Time: t})
	}

	resultBefore := q.Pop(&before)
	if resultBefore != nil {
		t.Log("fail on popping before time: expected nil, got job")
		t.Fail()
	}

	t.Log("starting first popping")
	var prev *UpdateJob
	for i:=0; i<30; i++ {
		job := q.Pop(&after)
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
		q.insert(&UpdateJob{Time: t})
	}

	t.Log("starting second popping")
	prev = nil
	for i:=0; i<60; i++ {
		job := q.Pop(&after)
		if job == nil {
			t.Log("fail on popping: expected job, got nil")
			t.Fail()
		} else if prev!=nil && job.Time.Before(prev.Time) {
			t.Log("fail on queueing: jobs popped out of order")
			t.Fail()
		}
		prev = job
	}

	if q.Pop(&after) != nil {
		t.Log("fail on popping after emptying: expected nil, got job")
		t.Fail()
	}
}
