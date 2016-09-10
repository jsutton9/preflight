package persistence

import (
	"testing"
	"fmt"
	"github.com/jsutton9/preflight/security"
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
	responseChannel := make(chan *User)
	go UserCache(readChannel, writeChannel)

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
	userWriteAfter := &User{Id: userWriteBefore.Id, Email: userWriteBefore.Email}
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
	responseChannel := make(chan *User)
	go UserCache(readChannel, writeChannel)

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
