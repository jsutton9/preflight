package security

import (
	"testing"
)

func TestPassword(t *testing.T) {
	sec, err := New("password")
	if err != nil {
		t.Fatal(err)
	}

	if ! sec.ValidatePassword("password") {
		t.Log("test failure, ValidatePassword: expected true, got false")
		t.Fail()
	}
	if sec.ValidatePassword("wrong") {
		t.Log("test failure, ValidatePassword: expected false, got true")
		t.Fail()
	}
}

func TestValidateToken(t *testing.T) {
	sec, err := New("password")
	if err != nil {
		t.Fatal(err)
	}

	permissions := PermissionFlags{ChecklistRead:true}
	noPermissions := PermissionFlags{}
	wrongPermissions := PermissionFlags{ChecklistWrite:true}

	token, err := sec.AddToken(permissions, 24, "foo")
	if err != nil {
		t.Fatal(err)
	}

	if ! sec.ValidateToken(token.Secret, permissions) {
		t.Log("test failure, ValidateToken: expected true, got false")
		t.Fail()
	}
	if ! sec.ValidateToken(token.Secret, noPermissions) {
		t.Log("test failure, ValidateToken: expected true, got false")
		t.Fail()
	}
	if sec.ValidateToken(token.Secret, wrongPermissions) {
		t.Log("test failure, ValidateToken: expected false, got true")
		t.Fail()
	}
	if sec.ValidateToken("wrong secret", permissions) {
		t.Log("test failure, ValidateToken: expected false, got true")
		t.Fail()
	}
}

func TestDeleteToken(t *testing.T) {
	sec, err := New("password")
	if err != nil {
		t.Fatal(err)
	}

	permissions := PermissionFlags{ChecklistRead:true}

	token, err := sec.AddToken(permissions, 24, "foo")
	if err != nil {
		t.Fatal(err)
	}
	err = sec.DeleteToken(token.Id)
	if err != nil {
		t.Fatal(err)
	}

	if sec.ValidateToken(token.Secret, permissions) {
		t.Log("test failure, ValidateToken: expected false, got true")
		t.Fail()
	}
}
