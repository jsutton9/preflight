package persistence

import (
	"testing"
	"fmt"
	"github.com/jsutton9/preflight/security"
	"math/rand"
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

	p, err := New("localhost", "preflight-test")
	if err != nil {
		t.Fatal(err)
	}
	user, err := p.AddUser(email, password)
	if err != nil {
		t.Fatal(err)
	}

	permissions := security.PermissionFlags{ChecklistRead:true}
	token, err := user.Security.AddToken(permissions, 24, "persistence test")
	if err != nil {
		t.Fatal(err)
	}
	err = p.UpdateUser(user)
	if err != nil {
		t.Fatal(err)
	}

	user, err = p.GetUserByToken(token.Secret)
	if err != nil {
		t.Fatal(err)
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
