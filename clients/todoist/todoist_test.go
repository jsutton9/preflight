package todoist

import (
	"testing"
	"os"
)

func TestPostTasks(t *testing.T) {
	key := os.Getenv("TEST_API_KEY")
	if key == "" {
		t.Fatal("missing environment variable TEST_API_KEY")
	}

	c := New(Security{key})

	id1, err := c.PostTask("foo")
	if err != nil {
		t.Error(err)
	}
	id2, err := c.PostTask("bar")
	if err != nil {
		t.Error(err)
	}

	err = c.DeleteTask(id1)
	if err != nil {
		t.Error(err)
	}
	err = c.DeleteTask(id2)
	if err != nil {
		t.Error(err)
	}
}
