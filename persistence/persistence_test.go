package persistence

import (
	"github.com/jsutton9/preflight/config"
	"os"
	"testing"
	"time"
)

func TestPersistence(t *testing.T) {
	err := os.Remove(os.Getenv("HOME")+"/.preflight/records/test.json")
	if (err != nil) && (! os.IsNotExist(err)) {
		t.Fatal(err)
	}

	p, err := Load("test")
	if err != nil {
		t.Fatal(err)
	}

	templates := make(map[string]config.Template)
	tasks := []string{"first","second"}
	templates["foo"] = config.Template{tasks,nil,nil}
	conf := config.Config{
		TodoistToken: "abc123",
		Templates: templates,
	}
	p.Config = conf

	updateTime := time.Now()
	addTime := updateTime.AddDate(0, 0, -2)
	ids := []int{12, 34}
	p.UpdateHistory["foo"] = UpdateRecord{
		Ids: ids,
		Time: updateTime,
		AddTime: addTime,
	}

	err = p.Save()
	if err != nil {
		t.Fatal(err)
	}

	p, err = Load("test")
	if err != nil {
		t.Fatal(err)
	}

	if p.Config.TodoistToken != "abc123" {
		t.Log("Config.TodoistToken not persisted correctly:")
		t.Logf("\texpected \"abc123\", got \"%s\"\n", p.Config.TodoistToken)
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
	} else if len(record.Ids) != 2 {
		t.Log("updateRecord.Ids incorrect:")
		t.Logf("\texpected %v, got %v", ids, record.Ids)
		t.Fail()
	} else if (record.Ids[0] != ids[0]) || (record.Ids[1] != ids[1]) {
		t.Log("updateRecord.Ids incorrect:")
		t.Logf("\texpected %v, got %v", ids, record.Ids)
		t.Fail()
	} else if delta:=record.Time.Unix()-updateTime.Unix(); delta > 60 || delta < -60 {
		t.Log("updateRecord.Time incorrect:")
		t.Logf("\texpected %v, got %v", updateTime, record.Time)
		t.Fail()
	} else if delta:=record.AddTime.Unix()-addTime.Unix(); delta > 60 || delta < -60 {
		t.Log("updateRecord.AddTime incorrect:")
		t.Logf("\texpected %v, got %v", addTime, record.AddTime)
		t.Fail()
	}
}
