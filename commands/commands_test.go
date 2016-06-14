package commands

import (
	"fmt"
	"github.com/jsutton9/preflight/persistence"
	"github.com/jsutton9/preflight/security"
	"math/rand"
	"testing"
)

func TestUserCommands(t *testing.T) {
	persister, err := persistence.New("localhost", "commands-test")
	email := fmt.Sprintf("testuser-%d@preflight.com", rand.Int())

	id, err := AddUser(email, "password", persister)
	if err != nil {
		t.Fatal(err)
	}

	idEmail, err := GetUserIdFromEmail(email, persister)
	if err != nil {
		t.Log("error getting user id from email: " +
			"\n\t" + err.Error())
		t.Fail()
	}
	if idEmail != id {
		t.Logf("incorrect id from GetUserIdFromEmail: " +
			"\n\t expected %s, got %s", id, idEmail)
		t.Fail()
	}

	token, err := AddToken(id, security.PermissionFlags{}, 24, "foo", persister)
	if err != nil {
		t.Fatal(err)
	}

	idToken, err := GetUserIdFromToken(token.Secret, persister)
	if err != nil {
		t.Log("error getting user id from token: " +
			"\n\t" + err.Error())
		t.Fail()
	}
	if idToken != id {
		t.Logf("incorrect id from GetUserIdFromToken: " +
			"\n\t expected %s, got %s", id, idToken)
		t.Fail()
	}
}
