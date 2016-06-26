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
	persister, err := persistence.New("localhost", "commands-test")
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

	id, err := AddUser(userReqString, persister)
	if err != nil {
		t.Fatal(err)
	}

	tokenString, err := AddToken(id, tokenReqString, persister)
	if err != nil {
		t.Fatal(err)
	}
	token := new(security.Token)
	err = json.Unmarshal([]byte(tokenString), token)
	if err != nil {
		t.Fatal(err)
	}

	idEmail, err := GetUserIdFromEmail(email, persister)
	if err != nil {
		t.Log("error getting user id from email: " +
			"\n\t" + err.Error())
		t.Fail()
	}

	idToken, err := GetUserIdFromToken(token.Secret, persister)
	if err != nil {
		t.Log("error getting user id from token: " +
			"\n\t" + err.Error())
		t.Fail()
	}

	passwordValidBefore, err := ValidatePassword(id, oldPassword, persister)
	if err != nil {
		t.Log("error validating password: " +
			"\n\t" + err.Error())
		t.Fail()
	}
	err = ChangePassword(id, newPassword, persister)
	if err != nil {
		t.Log("error changing password: " +
			"\n\t" + err.Error())
		t.Fail()
	}
	oldPasswordValidAfter, err := ValidatePassword(id, oldPassword, persister)
	if err != nil {
		t.Log("error validating password: " +
			"\n\t" + err.Error())
		t.Fail()
	}
	newPasswordValidAfter, err := ValidatePassword(id, newPassword, persister)
	if err != nil {
		t.Log("error validating password: " +
			"\n\t" + err.Error())
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
	if ! passwordValidBefore {
		t.Log("password validation before change incorrect: " +
			"\n\t expected true, got false")
		t.Fail()
	}
	if oldPasswordValidAfter {
		t.Log("old password validation after change incorrect: " +
			"\n\t expected false, got true")
		t.Fail()
	}
	if ! newPasswordValidAfter {
		t.Log("new password validation after change incorrect: " +
			"\n\t expected true, got false")
		t.Fail()
	}
}

func TestChecklistCommands(t *testing.T) {
	// setup
	rand.Seed(time.Now().UnixNano())
	persister, err := persistence.New("localhost", "commands-test")
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
	id, err := AddUser(userReqString, persister)
	if err != nil {
		t.Fatal(err)
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
	err = AddChecklist(id, checklistReqString, persister)
	if err != nil {
		t.Fatal(err)
	}
	checklistOut1String, err := GetChecklistString(id, name, persister)
	if err != nil {
		t.Log("error getting checklist string: " +
			"\n\t" + err.Error())
		t.Fail()
	}

	err = UpdateChecklist(id, name, checklistIn2String, persister)
	if err != nil {
		t.Log("error updating checklist: " +
			"\n\t" + err.Error())
		t.Fail()
	}
	checklistOut2String, err := GetChecklistString(id, name, persister)
	if err != nil {
		t.Log("error getting checklist string: " +
			"\n\t" + err.Error())
		t.Fail()
	}

	checklistsOutString, err := GetChecklistsString(id, persister)
	if err != nil {
		t.Log("error getting checklists string: " +
			"\n\t" + err.Error())
		t.Fail()
	}

	err = DeleteChecklist(id, name, persister)
	if err != nil {
		t.Log("error deleting checklist: " +
			"\n\t" + err.Error())
		t.Fail()
	}
	_, err = GetChecklistString(id, name, persister)
	if err == nil {
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
	persister, err := persistence.New("localhost", "commands-test")
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
	id, err := AddUser(userReqString, persister)
	if err != nil {
		t.Fatal(err)
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

	tokenString, err := AddToken(id, tokenReqString, persister)
	if err != nil {
		t.Fatal(err)
	}
	token := new(security.Token)
	err = json.Unmarshal([]byte(tokenString), token)
	if err != nil {
		t.Fatal(err)
	}

	tokensString1, err := GetTokens(id, persister)
	if err != nil {
		t.Log("error getting tokens: " +
			"\n\t" + err.Error())
		t.Fail()
	}
	err = DeleteToken(id, token.Id, persister)
	if err != nil {
		t.Log("error deleting token: " +
			err.Error())
		t.Fail()
	}
	tokensString2, err := GetTokens(id, persister)
	if err != nil {
		t.Log("error getting tokens: " +
			"\n\t" + err.Error())
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
	persister, err := persistence.New("localhost", "commands-test")
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
	id, err := AddUser(userReqString, persister)
	if err != nil {
		t.Fatal(err)
	}
	timezone := "Africa/Abidjan"

	err = SetGeneralSetting(id, "timezone", timezone, persister)
	if err != nil {
		t.Fatal(err)
	}
	settingsString, err := GetGeneralSettings(id, persister)
	if err != nil {
		t.Fatal(err)
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

//TODO: test Update, Invoke
