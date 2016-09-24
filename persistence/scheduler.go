package persistence

import (
	"github.com/jsutton9/preflight/checklist"
	"github.com/jsutton9/preflight/api/errors"
	"github.com/jsutton9/preflight/user"
	"time"
)

type UpdateJob struct {
	Time time.Time
	Checklist *checklist.Checklist
	User *user.User
	Remove bool
	TaskIds []string
	Index int
}

type Queue struct {
	Jobs []*UpdateJob
	Size int
	addsByChecklist map[string]*UpdateJob
	removesByChecklist map[string]*UpdateJob
}

func NewQueue() *Queue {
	return &Queue{
		Jobs: make([]*UpdateJob, 0),
		Size: 0,
		addsByChecklist: make(map[string]*UpdateJob),
		removesByChecklist: make(map[string]*UpdateJob),
	}
}

func (q *Queue) insert(job *UpdateJob) {
	if job.Remove {
		q.removesByChecklist[job.Checklist.Id] = job
	} else {
		q.addsByChecklist[job.Checklist.Id] = job
	}

	q.Jobs = append(q.Jobs, job)
	i := q.Size
	job.Index = i
	q.Size++

	i_parent := (i-1)/2
	parent := q.Jobs[i_parent]
	for i > 0 && job.Time.Before(parent.Time) {
		q.Jobs[i] = parent
		parent.Index = i
		q.Jobs[i_parent] = job
		job.Index = i_parent

		i = i_parent
		i_parent = (i-1)/2
		parent = q.Jobs[i_parent]
	}
}

func (q *Queue) remove(i int) {
	job := q.Jobs[i]
	if job.Remove {
		delete(q.removesByChecklist, job.Checklist.Id)
	} else {
		delete(q.addsByChecklist, job.Checklist.Id)
	}

	q.Size--
	q.Jobs[i] = q.Jobs[q.Size]
	q.Jobs = q.Jobs[:q.Size]

	if i == q.Size {
		return
	}
	for {
		job = q.Jobs[i]
		i_left := 2*i + 1
		i_right := 2*i + 2
		var leftChild, rightChild *UpdateJob
		leftChild = nil
		rightChild = nil
		if i_left < q.Size {
			leftChild = q.Jobs[i_left]
		}
		if i_right < q.Size {
			rightChild = q.Jobs[i_right]
		}

		var firstChild *UpdateJob
		if leftChild == nil {
			return
		} else if rightChild == nil || leftChild.Time.Before(rightChild.Time) {
			firstChild = leftChild
		} else {
			firstChild = rightChild
		}

		if firstChild.Time.Before(job.Time) {
			q.Jobs[firstChild.Index] = job
			q.Jobs[i] = firstChild
			job.Index = firstChild.Index
			firstChild.Index = i
			i = job.Index
		} else {
			return
		}
	}
}

func (q *Queue) Pop(now *time.Time) *UpdateJob {
	if q.Size == 0 || q.Jobs[0].Time.After(*now) {
		return nil
	} else {
		job := q.Jobs[0]
		q.remove(0)
		return job
	}
}

func (q *Queue) AddChecklist(u *user.User, cl *checklist.Checklist, now *time.Time) *errors.PreflightError {
	q.addsByChecklist[cl.Id] = nil
	q.removesByChecklist[cl.Id] = nil

	if cl.IsScheduled {
		t, err := cl.Schedule.NextAdd(*now)
		if err != nil {
			return err.Prepend("scheduler.Queue.AddChecklist: error getting add time: ")
		}
		job := &UpdateJob{
			Time: t,
			Checklist: cl,
			User: u,
		}
		q.insert(job)
	}

	return nil
}

func (q *Queue) AddUser(u *user.User, now *time.Time) {
	for _, cl := range u.Checklists {
		job, found := q.addsByChecklist[cl.Id]
		if found && job != nil {
			job.User = u
		}
		job, found = q.removesByChecklist[cl.Id]
		if found && job != nil {
			job.User = u
		}
	}
}

/*
func (q *Queue) RemoveChecklist(id string) {
	//TODO
}

func (q *Queue) RemoveUser(id string) {
	//TODO
}

func Schedule(updateChannel chan *user.UserDelta) {
	//TODO
}

func DoAdd(job *UpdateJob, removeChannel chan UpdateChannel) {
	//TODO
}
*/
