package commands

import (
	"fmt"
	"github.com/jsutton9/preflight/checklist"
	"github.com/jsutton9/preflight/persistence"
	"github.com/jsutton9/preflight/security"
	"encoding/json"
	"math/rand"
	"testing"
	"time"
)

func TestUserCommands(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	persister, pErr := persistence.New("localhost", "commands-test")
	email := fmt.Sprintf("testuser-%d@preflight.com", rand.Int())
	oldPassword := "old-pass"
	newPassword := "new-pass"
	tokenReq := tokenRequest{
		Permissions: security.PermissionFlags{},
		ExpiryHours: 24,
		Description: "foo",
	}
	tokenReqBytes, err := json.Marshal(tokenReq)
	if err != nil {
		t.Fatal(err)
	}
	tokenReqString := string(tokenReqBytes[:])
	userReq := userRequest{
		Email: email,
		Password: oldPassword,
	}
	userReqBytes, err := json.Marshal(userReq)
	if err != nil {
		t.Fatal(err)
	}
	userReqString := string(userReqBytes[:])

	id, pErr := AddUser(userReqString, persister)
	if pErr != nil {
		t.Fatal(pErr)
	}

	tokenString, pErr := AddToken(id, tokenReqString, persister)
	if pErr != nil {
		t.Fatal(pErr)
	}
	token := new(security.Token)
	err = json.Unmarshal([]byte(tokenString), token)
	if err != nil {
		t.Fatal(err)
	}

	idEmail, pErr := GetUserIdFromEmail(email, persister)
	if pErr != nil {
		t.Log("error getting user id from email: " +
			"\n\t" + pErr.Error())
		t.Fail()
	}

	idToken, pErr := GetUserIdFromToken(token.Secret, persister)
	if pErr != nil {
		t.Log("error getting user id from token: " +
			"\n\t" + pErr.Error())
		t.Fail()
	}

	passwordValidBefore := ValidatePassword(id, oldPassword, persister)
	pErr = ChangePassword(id, newPassword, persister)
	if pErr != nil {
		t.Log("error changing password: " +
			"\n\t" + pErr.Error())
		t.Fail()
	}
	oldPasswordValidAfter := ValidatePassword(id, oldPassword, persister)
	newPasswordValidAfter := ValidatePassword(id, newPassword, persister)

	pErr = DeleteUser(id, persister)
	if pErr != nil {
		t.Log("error deleting user: " +
			"\n\t" + pErr.Error())
		t.Fail()
	}

	if idEmail != id {
		t.Logf("incorrect id from GetUserIdFromEmail: " +
			"\n\t expected %s, got %s", id, idEmail)
		t.Fail()
	}
	if idToken != id {
		t.Logf("incorrect id from GetUserIdFromToken: " +
			"\n\t expected %s, got %s", id, idToken)
		t.Fail()
	}
	if passwordValidBefore != nil {
		t.Log("password validation before change incorrect: " +
			"\n\t expected nil, got error")
		t.Fail()
	}
	if oldPasswordValidAfter == nil {
		t.Log("old password validation after change incorrect: " +
			"\n\t expected error, got nil")
		t.Fail()
	}
	if newPasswordValidAfter != nil {
		t.Log("new password validation after change incorrect: " +
			"\n\t expected nil, got error")
		t.Fail()
	}
}

func TestChecklistCommands(t *testing.T) {
	// setup
	rand.Seed(time.Now().UnixNano())
	persister, pErr := persistence.New("localhost", "commands-test")
	email := fmt.Sprintf("testuser-%d@preflight.com", rand.Int())
	name := "foo"
	userReq := userRequest{
		Email: email,
		Password: "password",
	}
	userReqBytes, err := json.Marshal(userReq)
	if err != nil {
		t.Fatal(err)
	}
	userReqString := string(userReqBytes[:])
	id, pErr := AddUser(userReqString, persister)
	if pErr != nil {
		t.Fatal(pErr)
	}

	// setup (checklist inputs)
	checklistIn1 := checklist.Checklist{
		TasksSource: "preflight",
		TasksTarget: "todoist",
		IsScheduled: true,
		Tasks: []string{"before"},
	}
	checklistReq := checklistRequest{
		Name: name,
		Checklist: checklistIn1,
	}
	checklistIn2 := checklist.Checklist{
		TasksSource: "preflight",
		TasksTarget: "todoist",
		IsScheduled: true,
		Tasks: []string{"after"},
	}
	checklistReqBytes, err := json.Marshal(checklistReq)
	if err != nil {
		t.Fatal(err)
	}
	checklistReqString := string(checklistReqBytes[:])
	checklistIn2Bytes, err := json.Marshal(checklistIn2)
	if err != nil {
		t.Fatal(err)
	}
	checklistIn2String := string(checklistIn2Bytes[:])

	// execute checklist commands
	_, pErr = AddChecklist(id, checklistReqString, persister)
	if pErr != nil {
		t.Fatal(pErr)
	}
	checklistOut1String, pErr := GetChecklistString(id, name, persister)
	if pErr != nil {
		t.Log("error getting checklist string: " +
			"\n\t" + pErr.Error())
		t.Fail()
	}

	pErr = UpdateChecklist(id, name, checklistIn2String, persister)
	if pErr != nil {
		t.Log("error updating checklist: " +
			"\n\t" + pErr.Error())
		t.Fail()
	}
	checklistOut2String, pErr := GetChecklistString(id, name, persister)
	if pErr != nil {
		t.Log("error getting checklist string: " +
			"\n\t" + pErr.Error())
		t.Fail()
	}

	checklistsOutString, pErr := GetChecklistsString(id, persister)
	if pErr != nil {
		t.Log("error getting checklists string: " +
			"\n\t" + pErr.Error())
		t.Fail()
	}

	pErr = DeleteChecklist(id, name, persister)
	if pErr != nil {
		t.Log("error deleting checklist: " +
			"\n\t" + pErr.Error())
		t.Fail()
	}
	_, pErr = GetChecklistString(id, name, persister)
	if pErr == nil {
		t.Log("test failure: expected error from GetChecklistString after " +
			"delete, got nil")
		t.Fail()
	}

	// validate outputs
	checklistOut1 := checklist.Checklist{}
	err = json.Unmarshal([]byte(checklistOut1String), &checklistOut1)
	if err != nil {
		t.Log("error unmarshalling checklist \"" + checklistOut1String + "\": " +
			err.Error())
		t.Fail()
	} else if len(checklistOut1.Tasks) != len(checklistIn1.Tasks) ||
			checklistOut1.Tasks[0] != checklistIn1.Tasks[0] {
		t.Logf("test failure: checklistOut1 tasks wrong: " +
			"\n\texpected %v, got %v", checklistIn1.Tasks, checklistOut1.Tasks)
		t.Fail()
	}

	checklistOut2 := checklist.Checklist{}
	err = json.Unmarshal([]byte(checklistOut2String), &checklistOut2)
	if err != nil {
		t.Log("error unmarshalling checklist \"" + checklistOut2String + "\": " +
			err.Error())
		t.Fail()
	} else if len(checklistOut2.Tasks) != len(checklistIn2.Tasks) ||
			checklistOut2.Tasks[0] != checklistIn2.Tasks[0] {
		t.Logf("test failure: checklistOut2 tasks wrong: " +
			"\n\texpected %v, got %v", checklistIn2.Tasks, checklistOut2.Tasks)
		t.Fail()
	}

	checklistsOut := make(map[string]checklist.Checklist)
	err = json.Unmarshal([]byte(checklistsOutString), &checklistsOut)
	if err != nil {
		t.Log("error unmarshalling checklists \"" + checklistsOutString + "\": " +
			err.Error())
		t.Fail()
	} else {
		_, found := checklistsOut[name]
		if ! found {
			t.Log("test failure: checklist not found in checklists")
			t.Fail()
		}
	}
}

func TestTokenCommands(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	persister, pErr := persistence.New("localhost", "commands-test")
	email := fmt.Sprintf("testuser-%d@preflight.com", rand.Int())
	userReq := userRequest{
		Email: email,
		Password: "password",
	}
	userReqBytes, err := json.Marshal(userReq)
	if err != nil {
		t.Fatal(err)
	}
	userReqString := string(userReqBytes[:])
	id, pErr := AddUser(userReqString, persister)
	if pErr != nil {
		t.Fatal(pErr)
	}
	tokenReq := tokenRequest{
		Permissions: security.PermissionFlags{},
		ExpiryHours: 24,
		Description: "foo",
	}
	tokenReqBytes, err := json.Marshal(tokenReq)
	if err != nil {
		t.Fatal(err)
	}
	tokenReqString := string(tokenReqBytes[:])

	tokenString, pErr := AddToken(id, tokenReqString, persister)
	if pErr != nil {
		t.Fatal(pErr)
	}
	token := new(security.Token)
	err = json.Unmarshal([]byte(tokenString), token)
	if err != nil {
		t.Fatal(err)
	}

	tokensString1, pErr := GetTokens(id, persister)
	if pErr != nil {
		t.Log("error getting tokens: " +
			"\n\t" + pErr.Error())
		t.Fail()
	}
	pErr = DeleteToken(id, token.Id, persister)
	if pErr != nil {
		t.Log("error deleting token: " +
			pErr.Error())
		t.Fail()
	}
	tokensString2, pErr := GetTokens(id, persister)
	if pErr != nil {
		t.Log("error getting tokens: " +
			"\n\t" + pErr.Error())
		t.Fail()
	}

	tokens1 := []security.Token{}
	err = json.Unmarshal([]byte(tokensString1), &tokens1)
	if err != nil {
		t.Log("error unmarshalling tokens \"" + tokensString1 + "\": " +
			"\n\t" + err.Error())
		t.Fail()
	}
	found := false
	for _, tokenOut := range tokens1 {
		if tokenOut.Id == token.Id {
			found = true
			break
		}
	}
	if ! found {
		t.Log("added token not found in list")
		t.Fail()
	}

	tokens2 := []security.Token{}
	err = json.Unmarshal([]byte(tokensString2), &tokens2)
	if err != nil {
		t.Log("error unmarshalling tokens \"" + tokensString2 + "\": " +
			"\n\t" + err.Error())
		t.Fail()
	}
	for _, tokenOut := range tokens2 {
		if tokenOut.Id == token.Id {
			t.Log("deleted token found in list")
			t.Fail()
			break
		}
	}
}

func TestSettingsCommands(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	persister, pErr := persistence.New("localhost", "commands-test")
	email := fmt.Sprintf("testuser-%d@preflight.com", rand.Int())
	userReq := userRequest{
		Email: email,
		Password: "password",
	}
	userReqBytes, err := json.Marshal(userReq)
	if err != nil {
		t.Fatal(err)
	}
	userReqString := string(userReqBytes[:])
	id, pErr := AddUser(userReqString, persister)
	if pErr != nil {
		t.Fatal(pErr)
	}
	timezone := "Africa/Abidjan"

	pErr = SetGeneralSetting(id, "timezone", timezone, persister)
	if pErr != nil {
		t.Fatal(pErr)
	}
	settingsString, pErr := GetGeneralSettings(id, persister)
	if pErr != nil {
		t.Fatal(pErr)
	}
	settings := new(persistence.GeneralSettings)
	err = json.Unmarshal([]byte(settingsString), settings)
	if err != nil {
		t.Fatal(err)
	}

	if settings.Timezone != timezone {
		t.Log("test failure: timezone incorrect: " +
			"\n\texpected %s, got %s", timezone, settings.Timezone)
		t.Fail()
	}
}

//TODO: test Update, Invoke, ValidateToken
