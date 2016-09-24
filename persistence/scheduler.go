package persistence

import (
	"time"
	"github.com/jsutton9/preflight/checklist"
	"github.com/jsutton9/preflight/user"
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
	byChecklist map[string][]*UpdateJob
	byUser map[string][]*UpdateJob
}

func NewQueue() *Queue {
	return &Queue{
		Jobs: make([]*UpdateJob, 0),
		Size: 0,
		byChecklist: make(map[string][]*UpdateJob),
		byUser: make(map[string][]*UpdateJob),
	}
}

func (q *Queue) insert(job *UpdateJob) {
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
	q.Size--
	q.Jobs[i] = q.Jobs[q.Size]
	q.Jobs = q.Jobs[:q.Size]

	if i == q.Size {
		return
	}
	for {
		job := q.Jobs[i]
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

/*func (q *Queue) AddChecklist(u *user.User, cl *checklist.Checklist, now *time.Time) {
	//TODO
}

func (q *Queue) AddUser(u *user.User, now *time.Time) {
	//TODO
}

func (q *Queue) RemoveChecklist(id string) {
	//TODO
}

func (q *Queue) RemoveUser(id string) {
	//TODO
}

func Schedule(updateChannel chan *user.UserDelta) {
	//TODO
}*/
