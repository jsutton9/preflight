package trello

import (
	"os"
	"testing"
)

/* This test expects a board named "Todo Test" with a list named "Test" 
 * with (only) cards named "foo", "bar", and "baz", in that order.
 */
func TestGetList(t *testing.T) {
	key := os.Getenv("TEST_TRELLO_KEY")
	if key == "" {
		t.Fatal("missing environment variable TEST_TRELLO_KEY")
	}
	token := os.Getenv("TEST_TRELLO_TOKEN")
	if token == "" {
		t.Fatal("missing environment variable TEST_TRELLO_TOKEN")
	}

	boardName := "Todo Test"
	listName := "Test"
	correctTasks := []string{"foo", "bar", "baz"}

	t.Log("testing with constructor params")
	c := New(Security{Token:token}, key, boardName)
	tasks, err := c.Tasks(&ListKey{Board:"",Name:listName})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != len(correctTasks) {
		t.Logf("tasks wrong: expected %v, got %v", correctTasks, tasks)
		t.Fail()
	} else {
		for i, task := range correctTasks {
			if tasks[i] != task {
				t.Logf("tasks wrong: expected %v, got %v",
					correctTasks, tasks)
				t.Fail()
				break
			}
		}
	}

	t.Log("testing with method params")
	c = New(Security{Token:token}, key, "wrong")
	tasks, err = c.Tasks(&ListKey{Board:boardName,Name:listName})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != len(correctTasks) {
		t.Logf("tasks wrong: expected %v, got %v", correctTasks, tasks)
		t.Fail()
	} else {
		for i, task := range correctTasks {
			if tasks[i] != task {
				t.Logf("tasks wrong: expected %v, got %v",
					correctTasks, tasks)
				t.Fail()
				break
			}
		}
	}
}
