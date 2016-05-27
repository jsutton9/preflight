package commands

import (
	"encoding/json"
	"github.com/jsutton9/preflight/config"
	"github.com/jsutton9/preflight/persistence"
	"testing"
)

func TestTemplateCommands(test *testing.T) {
	// ensure user 'test' exists
	err := SetConfig("test", "../config/test_config.json")
	if err != nil {
		test.Fatal(err)
	}

	// create template request string
	testTemplate := config.Template{
		Tasks:    []string{"foo", "bar"},
		Trello:   &config.Trello{Key: "abc123"},
		Schedule: &config.Schedule{Start: "09:30"},
	}
	testTemplateReq := templateRequest{
		Name:     "test-template",
		Template: testTemplate,
	}
	jsonBytes, err := json.Marshal(testTemplateReq)
	if err != nil {
		test.Fatal(err)
	}

	// pass templateRequest to AddTemplate
	err = AddTemplate("test", string(jsonBytes[:]))
	if err != nil {
		test.Log("test failure: error from AddTemplate: " +
			"\n\t" + err.Error())
		test.Fail()
	}

	// verify that the template was added
	persist, err := persistence.Load("test")
	if err != nil {
		test.Fatal(err)
	}
	template, found := persist.Config.Templates["test-template"]
	if ! found {
		test.Log("test failure: template from AddTemplate not found")
		test.Fail()
	}

	// make a change to the template and remarshal
	testTemplate.Tasks = []string{"changed"}
	jsonBytes, err = json.Marshal(testTemplate)
	if err != nil {
		test.Fatal(err)
	}

	// UpdateTemplate gets template object, not templateRequest
	err = UpdateTemplate("test", "test-template", string(jsonBytes[:]))
	if err != nil {
		test.Log("test failure: error from UpdateTemplate: " +
			"\n\t" + err.Error())
		test.Fail()
	}

	// verify that the template was updated correctly
	persist, err = persistence.Load("test")
	if err != nil {
		test.Fatal(err)
	}
	template, found = persist.Config.Templates["test-template"]
	if ! found {
		test.Log("test failure: template not found after UpdateTemplate")
		test.Fail()
	} else if (len(template.Tasks) != 1) || (template.Tasks[0] != testTemplate.Tasks[0]) {
		test.Logf("test failure: tasks wrong after UpdateTemplate: " +
			"\n\texpected %v, got %v", testTemplate.Tasks, template.Tasks)
		test.Fail()
	}

	// delete template
	err = DeleteTemplate("test", "test-template")
	if err != nil {
		test.Log("test failure: error from DeleteTemplate: " +
			"\n\t" + err.Error())
		test.Fail()
	}

	// verify that the template is gone
	persist, err = persistence.Load("test")
	if err != nil {
		test.Fatal(err)
	}
	_, found = persist.Config.Templates["test-template"]
	if found {
		test.Log("test failure: template still exists after DeleteTemplate")
		test.Fail()
	}
}

func TestSettingsCommands(test *testing.T) {
	// ensure user 'test' exists
	err := SetConfig("test", "../config/test_config.json")
	if err != nil {
		test.Fatal(err)
	}

	// set Trello config
	trello := config.Trello{Key: "updated key"}
	jsonBytes, err := json.Marshal(trello)
	if err != nil {
		test.Fatal(err)
	}
	err = SetGlobalSetting("test", "trello", string(jsonBytes))
	if err != nil {
		test.Log("test failure: error from SetGlobalSetting: " +
			"\n\t" + err.Error())
		test.Fail()
	}

	// get global settings
	settingsString, err := GetGlobalSettings("test")
	if err != nil {
		test.Log("test failure: error from GetGlobalSettings: " +
			"\n\t" + err.Error())
		test.Fail()
	}

	// unmarshal and check for change
	conf := config.GlobalSettings{}
	err = json.Unmarshal([]byte(settingsString), &conf)
	if err != nil {
		test.Log("test failure: error unmarshalling \"" + settingsString + "\": " +
			"\n\t" + err.Error())
	}
	if conf.Trello.Key != trello.Key {
		test.Logf("test failure: conf.Trello.Key wrong: " +
			"\n\t" + "expected \"%s\", got \"%s\"",
			trello.Key, conf.Trello.Key)
		test.Fail()
	}
}
