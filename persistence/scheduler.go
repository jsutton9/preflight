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

type checklistTracker struct {
	User *user.User
	Checklist *checklist.Checklist
	AddJob *UpdateJob
	RemoveJob *UpdateJob
}

type Queue struct {
	Jobs []*UpdateJob
	Size int
	trackersByChecklist map[string]*checklistTracker
	trackersByUser map[string][]*checklistTracker
}

func NewQueue() *Queue {
	return &Queue{
		Jobs: make([]*UpdateJob, 0),
		Size: 0,
		trackersByChecklist: make(map[string]*checklistTracker),
		trackersByUser: make(map[string][]*checklistTracker),
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
	job := q.Jobs[i]

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

func (q *Queue) Pop(now *time.Time) (*UpdateJob, *errors.PreflightError) {
	if q.Size == 0 || q.Jobs[0].Time.After(*now) {
		return nil, nil
	} else {
		job := q.Jobs[0]
		q.remove(0)
		if job.Remove == false && job.Checklist.IsScheduled {
			t, err := job.Checklist.Schedule.NextAdd(*now)
			if err != nil {
				return job, err.Prepend("persistence.Queue.Pop: " +
					"error scheduling next add: ")
			}
			q.insert(&UpdateJob{
				Time: t,
				Checklist: job.Checklist,
				User: job.User,
			})
		}
		return job, nil
	}
}

func (q *Queue) SetChecklist(u *user.User, cl *checklist.Checklist, now *time.Time) *errors.PreflightError {
	tracker, found := q.trackersByChecklist[cl.Id]
	var prev *checklist.Checklist
	if found {
		prev = tracker.Checklist
		tracker.Checklist = cl
		if ! cl.IsScheduled && tracker.AddJob != nil {
			q.remove(tracker.AddJob.Index)
			tracker.AddJob = nil
		}
	} else {
		tracker = &checklistTracker{
			User: u,
			Checklist: cl,
		}
		q.trackersByChecklist[cl.Id] = tracker

		trackers, userFound := q.trackersByUser[u.GetId()]
		if ! userFound {
			q.trackersByUser[u.GetId()] = make([]*checklistTracker, 0, 1)
		}
		q.trackersByUser[u.GetId()] = append(trackers, tracker)
	}

	if cl.IsScheduled && ! (found && cl.Schedule.Equals(prev.Schedule)) {
		if tracker.AddJob != nil {
			q.remove(tracker.AddJob.Index)
		}

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
		tracker.AddJob = job
	}

	return nil
}

func (q *Queue) RemoveChecklist(id string) {
	tracker, found := q.trackersByChecklist[id]
	if found {
		trackers, _ := q.trackersByUser[tracker.User.GetId()]
		for i, tr := range trackers {
			if tr == tracker {
				trackers[i] = trackers[len(trackers)-1]
				q.trackersByUser[tracker.User.GetId()] = trackers[:len(trackers)-1]
				break
			}
		}

		if tracker.AddJob != nil {
			q.remove(tracker.AddJob.Index)
		}
		delete(q.trackersByChecklist, id)
	}
}

func (q *Queue) SetUser(u *user.User, now *time.Time) {
	for _, cl := range u.Checklists {
		tracker, found := q.trackersByChecklist[cl.Id]
		if found && tracker != nil {
			tracker.User = u
			if tracker.AddJob != nil {
				tracker.AddJob.User = u
			}
			if tracker.RemoveJob != nil {
				tracker.RemoveJob.User = u
			}
		}
	}
}

func (q *Queue) RemoveUser(id string) {
	delete(q.trackersByUser, id)
}

/*
func Schedule(updateChannel chan *user.UserDelta) {
	//TODO
}

func DoAdd(job *UpdateJob, removeChannel chan UpdateChannel) {
	//TODO
}
*/
