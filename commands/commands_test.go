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
	rand.Seed(time.Now().Unix())
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

func TestChecklistCommands(t *testing.T) {
	// setup
	rand.Seed(time.Now().Unix())
	persister, err := persistence.New("localhost", "commands-test")
	email := fmt.Sprintf("testuser-%d@preflight.com", rand.Int())
	name := "foo"
	id, err := AddUser(email, "password", persister)
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

func testSettingsCommands(t *testing.T) {
	rand.Seed(time.Now().Unix())
	persister, err := persistence.New("localhost", "commands-test")
	email := fmt.Sprintf("testuser-%d@preflight.com", rand.Int())
	id, err := AddUser(email, "password", persister)
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
