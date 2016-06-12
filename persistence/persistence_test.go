package persistence

import (
	"testing"
	"github.com/jsutton9/preflight/security"
)

func TestGetUser(t *testing.T) {
	email := "foo@bar"
	user := &User{Email: email}

	p, err := New("localhost", "preflight-test")
	if err != nil {
		t.Fatal(err)
	}
	err = p.AddUser(user)
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
	email := "foo@baz"
	user := &User{Email: email}

	p, err := New("localhost", "preflight-test")
	if err != nil {
		t.Fatal(err)
	}
	err = p.AddUser(user)
	if err != nil {
		t.Fatal(err)
	}
	user, err = p.GetUserByEmail(email)
	if err != nil {
		t.Fatal(err)
	}

	if user.Email != email {
		t.Logf("email wrong: expected \"%s\", got \"%s\"", email, user.Email)
		t.Fail()
	}
}

func TestGetUserByToken(t *testing.T) {
	email := "foo@abc"
	sec, err := security.New("password")
	if err != nil {
		t.Fatal(err)
	}
	permissions := security.PermissionFlags{ChecklistRead:true}
	token, err := sec.AddToken(permissions, 24, "persistence test")
	if err != nil {
		t.Fatal(err)
	}
	user := &User{
		Email: email,
		Security: sec,
	}

	p, err := New("localhost", "preflight-test")
	if err != nil {
		t.Fatal(err)
	}
	err = p.AddUser(user)
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
	emailBefore := "abc@foo"
	emailAfter := "def@bar"
	user := &User{Email: emailBefore}

	p, err := New("localhost", "preflight-test")
	if err != nil {
		t.Fatal(err)
	}
	err = p.AddUser(user)
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
