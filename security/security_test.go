package security

import (
	"testing"
)

func TestPassword(t *testing.T) {
	sec, err := New("password")
	if err != nil {
		t.Fatal(err)
	}

	if sec.ValidatePassword("password") != nil {
		t.Log("test failure, ValidatePassword: expected nil, got error")
		t.Fail()
	}
	if sec.ValidatePassword("wrong") == nil {
		t.Log("test failure, ValidatePassword: expected error, got nil")
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

	if len(token.Id) != SECRET_BITS/4 {
		t.Log("test failure, AddToken: ")
		t.Logf("\texpected %d char id, got %s", ID_BITS/4, token.Id)
		t.Fail()
	}
	if len(token.Secret) != SECRET_BITS/4 {
		t.Log("test failure, AddToken: ")
		t.Logf("\texpected %d char secret, got %s", SECRET_BITS/4, token.Secret)
		t.Fail()
	}

	if sec.ValidateToken(token.Secret, permissions) != nil {
		t.Log("test failure, ValidateToken: expected nil, got error")
		t.Fail()
	}
	if sec.ValidateToken(token.Secret, noPermissions) != nil {
		t.Log("test failure, ValidateToken: expected nil, got error")
		t.Fail()
	}
	if sec.ValidateToken(token.Secret, wrongPermissions) == nil {
		t.Log("test failure, ValidateToken: expected error, got nil")
		t.Fail()
	}
	if sec.ValidateToken("wrong secret", permissions) == nil {
		t.Log("test failure, ValidateToken: expected error, got nil")
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

	if sec.ValidateToken(token.Secret, permissions) == nil {
		t.Log("test failure, ValidateToken: expected error, got nil")
		t.Fail()
	}
}
