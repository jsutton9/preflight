package persistence

import (
	"github.com/jsutton9/todoistist/config"
	"os"
	"testing"
	"time"
)

func TestPersistence(t *testing.T) {
	err := os.Remove("/var/lib/todoistist/test.json")
	if (err != nil) && (! os.IsNotExist(err)) {
		t.Fatal(err)
	}

	p, err := Load("test")
	if err != nil {
		t.Fatal(err)
	}

	templates := make(map[string]config.Template)
	tasks := []string{"first","second"}
	templates["foo"] = config.Template{tasks,nil}
	conf := config.Config{
		ApiToken: "abc123",
		Templates: templates,
	}
	p.Config = conf

	updateTime := time.Now()
	uuids := []string{"ab12", "cd34"}
	p.UpdateHistory["foo"] = UpdateRecord{
		Uuids: uuids,
		Time: updateTime,
	}

	err = p.Save()
	if err != nil {
		t.Fatal(err)
	}

	p, err = Load("test")
	if err != nil {
		t.Fatal(err)
	}

	if p.Config.ApiToken != "abc123" {
		t.Log("Config.ApiToken not persisted correctly:")
		t.Logf("\texpected \"abc123\", got \"%s\"\n", p.Config.ApiToken)
		t.Fail()
	}

	template, found := p.Config.Templates["foo"]
	if ! found {
		t.Log("Template \"foo\" not found")
		t.Fail()
	} else if len(template.Tasks) != 2 {
		t.Log("template.Tasks incorrect:")
		t.Logf("\texpected %v, got %v", tasks, template.Tasks)
		t.Fail()
	} else if (template.Tasks[0] != tasks[0]) || (template.Tasks[1] != tasks[1]) {
		t.Log("template.Tasks incorrect:")
		t.Logf("\texpected %v, got %v", tasks, template.Tasks)
		t.Fail()
	}

	record, found := p.UpdateHistory["foo"]
	if ! found {
		t.Log("Update record for \"foo\" not found")
		t.Fail()
	} else if len(record.Uuids) != 2 {
		t.Log("updateRecord.Uuids incorrect:")
		t.Logf("\texpected %v, got %v", uuids, record.Uuids)
		t.Fail()
	} else if (record.Uuids[0] != uuids[0]) || (record.Uuids[1] != uuids[1]) {
		t.Log("updateRecord.Uuids incorrect:")
		t.Logf("\texpected %v, got %v", uuids, record.Uuids)
		t.Fail()
	}
}
