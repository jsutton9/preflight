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
