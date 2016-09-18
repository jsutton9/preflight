package persistence

import (
	"testing"
	"fmt"
	"github.com/jsutton9/preflight/checklist"
	"github.com/jsutton9/preflight/security"
	"github.com/jsutton9/preflight/user"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
)

func TestGetUser(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	email := fmt.Sprintf("testuser-%d@preflight.com", rand.Int())
	password := "password"

	p, err := New("localhost", "preflight-test")
	if err != nil {
		t.Fatal(err)
	}
	user, err := p.AddUser(email, password)
	if err != nil {
		t.Fatal(err)
	}
	id := user.GetId()
	user, err = p.GetUser(id)
	if err != nil {
		t.Fatal(err)
	}

	if user.Email != email {
		t.Logf("email wrong: expected \"%s\", got \"%s\"", email, user.Email)
		t.Fail()
	}
}

func TestGetUserByEmail(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	email := fmt.Sprintf("testuser-%d@preflight.com", rand.Int())
	password := "password"

	p, err := New("localhost", "preflight-test")
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.AddUser(email, password)
	if err != nil {
		t.Fatal(err)
	}
	user, err := p.GetUserByEmail(email)
	if err != nil {
		t.Fatal(err)
	}

	if user.Email != email {
		t.Logf("email wrong: expected \"%s\", got \"%s\"", email, user.Email)
		t.Fail()
	}
}

func TestGetUserByToken(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	email := fmt.Sprintf("testuser-%d@preflight.com", rand.Int())
	password := "password"

	p, pErr := New("localhost", "preflight-test")
	if pErr != nil {
		t.Fatal(pErr)
	}
	user, pErr := p.AddUser(email, password)
	if pErr != nil {
		t.Fatal(pErr)
	}

	permissions := security.PermissionFlags{ChecklistRead:true}
	token, err := user.Security.AddToken(permissions, 24, "persistence test")
	if err != nil {
		t.Fatal(err)
	}
	pErr = p.UpdateUser(user)
	if pErr != nil {
		t.Fatal(pErr)
	}

	user, pErr = p.GetUserByToken(token.Secret)
	if pErr != nil {
		t.Fatal(pErr)
	}

	if user.Email != email {
		t.Logf("email wrong: expected \"%s\", got \"%s\"", email, user.Email)
		t.Fail()
	}
}

func TestUpdateUser(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	emailBefore := fmt.Sprintf("testuser-%d@preflight.com", rand.Int())
	emailAfter := fmt.Sprintf("testuser-%d@preflight.com", rand.Int())
	password := "password"

	p, err := New("localhost", "preflight-test")
	if err != nil {
		t.Fatal(err)
	}
	user, err := p.AddUser(emailBefore, password)
	if err != nil {
		t.Fatal(err)
	}
	user.Email = emailAfter
	err = p.UpdateUser(user)
	if err != nil {
		t.Fatal(err)
	}
	user, err = p.GetUser(user.GetId())
	if err != nil {
		t.Fatal(err)
	}

	if user.Email != emailAfter {
		t.Logf("email wrong: expected \"%s\", got \"%s\"", emailAfter, user.Email)
		t.Fail()
	}
}

func TestDeleteUser(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	email := fmt.Sprintf("testuser-%d@preflight.com", rand.Int())
	password := "password"

	p, err := New("localhost", "preflight-test")
	if err != nil {
		t.Fatal(err)
	}
	user, err := p.AddUser(email, password)
	if err != nil {
		t.Fatal(err)
	}

	err = p.DeleteUser(user)
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.GetUser(user.GetId())
	if err == nil {
		t.Logf("user not deleted: expected error on GetUser, got nil")
		t.Fail()
	}
}

func TestNoDuplicateEmails(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	email := fmt.Sprintf("testuser-%d@preflight.com", rand.Int())
	password := "password"

	p, err := New("localhost", "preflight-test")
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.AddUser(email, password)
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.AddUser(email, password)
	if err == nil {
		t.Log("test failure: expected error adding duplicate user, got nil")
		t.Fail()
	}
}

func TestNode(t *testing.T) {
	wrongSecret := "wrong"
	secretFile := "/etc/preflight/test/secret"
	p, err := New("localhost", "preflight-test")
	if err != nil {
		t.Fatal(err)
	}

	err = p.RegisterNode(secretFile)
	if err != nil {
		t.Fatal(err)
	}
	secret, err := p.GetNodeSecret(secretFile)
	if err != nil {
		t.Fatal(err)
	}

	secretValid, err := p.ValidateNodeSecret(secret)
	if err != nil {
		t.Log("error validating secret: " +
			"\n\t" + err.Error())
		t.Fail()
	}
	wrongSecretValid, err := p.ValidateNodeSecret(wrongSecret)
	if err != nil {
		t.Log("error validating secret: " +
			"\n\t" + err.Error())
		t.Fail()
	}

	if len(secret) != security.SECRET_BITS/4 {
		t.Log("secret has wrong length: expected length %d, got %s",
			security.SECRET_BITS/4, secret)
		t.Fail()
	}
	if ! secretValid {
		t.Log("secret incorrectly validated: expected true, got false")
		t.Fail()
	}
	if wrongSecretValid {
		t.Log("secret incorrectly validated: expected false, got true")
		t.Fail()
	}
}

func TestLogging(t *testing.T) {
	testDir := "/var/log/preflight/test/"
	testFile := testDir + "foo/test.log"
	testString := "test string"
	err := os.RemoveAll(testDir)
	if err != nil {
		t.Fatal(err)
	}
	s := ServerSettings{ErrLog: testFile}

	logger, pErr := s.GetLogger()
	if pErr != nil {
		pErr.Prepend("error getting logger: ")
		t.Fatal(pErr)
	}
	defer logger.Close()
	logger.Println(testString)

	readBytes, err := ioutil.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	readString := string(readBytes)

	if ! strings.Contains(readString, testString) {
		t.Logf("logged message incorrect: expected \"%s\", got \"%s\"",
			testString, readString)
		t.Fail()
	}
}

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
