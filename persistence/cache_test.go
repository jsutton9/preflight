package persistence

import (
	"github.com/jsutton9/preflight/checklist"
	"github.com/jsutton9/preflight/security"
	"github.com/jsutton9/preflight/user"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	p, err := New("localhost", "preflight-test")
	if err != nil {
		t.Fatal(err)
	}
	emailWriteBefore := "before@preflight.com"
	emailWriteAfter := "after@preflight.com"
	userWrite, err := p.GetUserByEmail(emailWriteBefore)
	if err != nil {
		if err.Status == 404 {
			userWrite, err = p.AddUser(emailWriteBefore, "pass")
			if err != nil {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	}
	readChannel := make(chan *ReadRequest)
	writeChannel := make(chan *WriteRequest)
	responseChannel := make(chan *user.User)
	updateChannel := make(chan *user.UserDelta)
	go UserCache(readChannel, writeChannel, updateChannel)

	writeChannel <- &WriteRequest{User: userWrite}
	readChannel <- &ReadRequest{Id: userWrite.GetId(), ResponseChannel: responseChannel}
	userRead := <-responseChannel
	emailReadBefore := userRead.Email

	userWrite.Email = emailWriteAfter
	writeChannel <- &WriteRequest{User: userWrite}
	readChannel <- &ReadRequest{Id: userWrite.GetId(), ResponseChannel: responseChannel}
	userRead = <-responseChannel
	emailReadAfter := userRead.Email

	writeChannel <- &WriteRequest{User: userWrite, Remove: true}
	readChannel <- &ReadRequest{Id: userWrite.GetId(), ResponseChannel: responseChannel}
	userRead = <-responseChannel

	if emailReadBefore != emailWriteBefore {
		t.Logf("email wrong after initial write: expected \"%s\", got \"%s\"",
			emailWriteBefore, emailReadBefore)
		t.Fail()
	}
	if emailReadAfter != emailWriteAfter {
		t.Logf("email wrong after rewrite: expected \"%s\", got \"%s\"",
			emailWriteAfter, emailReadAfter)
		t.Fail()
	}
	if userRead != nil {
		t.Log("user not removed: expected nil, got non-nil")
		t.Fail()
	}
}

func TestCacheByToken(t *testing.T) {
	p, err := New("localhost", "preflight-test")
	if err != nil {
		t.Fatal(err)
	}
	email := "before@preflight.com"
	userWriteBefore, err := p.GetUserByEmail(email)
	if err != nil {
		if err.Status == 404 {
			userWriteBefore, err = p.AddUser(email, "pass")
			if err != nil {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	}

	userWriteBefore.Security, err = security.New("pass")
	if err != nil {
		t.Fatal(err)
	}
	token, err := userWriteBefore.Security.AddToken(security.PermissionFlags{}, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	secretBefore := token.Secret
	userWriteAfter := &user.User{Id: userWriteBefore.Id, Email: userWriteBefore.Email}
	userWriteAfter.Security, err = security.New("pass")
	if err != nil {
		t.Fatal(err)
	}
	token, err = userWriteAfter.Security.AddToken(security.PermissionFlags{}, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	secretAfter := token.Secret

	readChannel := make(chan *ReadRequest)
	writeChannel := make(chan *WriteRequest)
	responseChannel := make(chan *user.User)
	updateChannel := make(chan *user.UserDelta)
	go UserCache(readChannel, writeChannel, updateChannel)

	writeChannel <- &WriteRequest{User: userWriteBefore}
	readChannel <- &ReadRequest{TokenSecret: secretBefore, ResponseChannel: responseChannel}
	userReadBefore := <-responseChannel
	writeChannel <- &WriteRequest{User: userWriteAfter}
	readChannel <- &ReadRequest{TokenSecret: secretBefore, ResponseChannel: responseChannel}
	userReadOldToken := <-responseChannel
	readChannel <- &ReadRequest{TokenSecret: secretAfter, ResponseChannel: responseChannel}
	userReadNewToken := <-responseChannel

	if userReadBefore == nil {
		t.Log("initial get by token failed: got nil")
		t.Fail()
	}
	if userReadOldToken != nil {
		t.Log("get by old token failed: expected nil, got non-nil")
		t.Fail()
	}
	if userReadNewToken == nil {
		t.Log("get by updated token failed: got nil")
		t.Fail()
	}
}

func checkDelta(t *testing.T, delta *user.UserDelta, timeout bool,
		expectEmail string, expectAdd int, expectRemove int, expectUpdate int) {
	if timeout {
		t.Log("update timed out")
		t.Fail()
		return
	}
	if delta.User.Email != expectEmail {
		t.Logf("delta email wrong: expected %s, got %s\n", expectEmail, delta.User.Email)
		t.Fail()
	}
	if len(delta.Added) != expectAdd || len(delta.Removed) != expectRemove ||
			len(delta.Updated) != expectUpdate {
		t.Log( "delta checklists wrong: ")
		t.Logf("  expected %d added, %d removed, %d updated",
			expectAdd, expectRemove, expectUpdate)
		t.Logf("  got      %d added, %d removed, %d updated",
			len(delta.Added), len(delta.Removed), len(delta.Updated))
		t.Fail()
	}
}

func TestCacheUpdates(t *testing.T) {
	emailBefore := "before@preflight.com"
	emailAfter := "after@preflight.com"
	userBefore, err := user.New(emailBefore, "pass")
	if err != nil {
		t.Fatal(err)
	}
	userBefore.Checklists = append(userBefore.Checklists, &checklist.Checklist{Id:"a"})
	userBefore.Checklists = append(userBefore.Checklists, &checklist.Checklist{Id:"b"})
	userBefore.Checklists = append(userBefore.Checklists, &checklist.Checklist{Id:"c"})
	userAfter, err := user.New(emailAfter, "pass")
	if err != nil {
		t.Fatal(err)
	}
	userAfter.Id = userBefore.Id
	userAfter.Checklists = append(userAfter.Checklists, &checklist.Checklist{Id:"a"})
	userAfter.Checklists = append(userAfter.Checklists, &checklist.Checklist{Id:"b", Name:"not-empty"})
	userAfter.Checklists = append(userAfter.Checklists, &checklist.Checklist{Id:"d"})

	readChannel := make(chan *ReadRequest)
	writeChannel := make(chan *WriteRequest)
	updateChannel := make(chan *user.UserDelta)
	go UserCache(readChannel, writeChannel, updateChannel)

	writeChannel <- &WriteRequest{User: userBefore}
	initialDelta := &user.UserDelta{}
	initialTimeout := false
	select {
	case initialDelta = <-updateChannel:
	case <-time.After(time.Millisecond*1000):
		initialTimeout = true
	}
	writeChannel <- &WriteRequest{User: userBefore}
	unchangedDelta := &user.UserDelta{}
	unchangedTimeout := false
	select {
	case unchangedDelta = <-updateChannel:
	case <-time.After(time.Millisecond*1000):
		unchangedTimeout = true
	}
	writeChannel <- &WriteRequest{User: userAfter}
	changedDelta := &user.UserDelta{}
	changedTimeout := false
	select {
	case changedDelta = <-updateChannel:
	case <-time.After(time.Millisecond*1000):
		changedTimeout = true
	}
	writeChannel <- &WriteRequest{User: userAfter, Remove: true}
	removedDelta := &user.UserDelta{}
	removedTimeout := false
	select {
	case removedDelta = <-updateChannel:
	case <-time.After(time.Millisecond*1000):
		removedTimeout = true
	}

	t.Log("checking initial write")
	checkDelta(t, initialDelta, initialTimeout, emailBefore, 3, 0, 0)
	t.Log("checking unchanged write")
	checkDelta(t, unchangedDelta, unchangedTimeout, emailBefore, 0, 0, 0)
	t.Log("checking changed write")
	checkDelta(t, changedDelta, changedTimeout, emailAfter, 1, 1, 1)
	t.Log("checking removed write")
	checkDelta(t, removedDelta, removedTimeout, emailAfter, 0, 3, 0)
}
