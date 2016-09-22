package persistence

import (
	"time"
	"github.com/jsutton9/preflight/checklist"
	"github.com/jsutton9/preflight/user"
)

type UpdateJob struct {
	Time *time.Time
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

func (q *Queue) insert(job *UpdateJob) {
	//TODO
}

func (q *Queue) remove(i int) {
	//TODO
}

func (q *Queue) Pop(now *time.Time) *UpdateJob {
	//TODO
}

func (q *Queue) AddChecklist(u *user.User, cl *checklist.Checklist, now *time.Time) {
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
}
