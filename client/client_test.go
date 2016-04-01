package client

import (
	"testing"
	"os"
)

func TestPostTasks(t *testing.T) {
	key := os.Getenv("TEST_API_KEY")
	if key == "" {
		t.Fatal("missing environment variable TEST_API_KEY")
	}

	c := New(key)

	_, err := c.PostTask("foo")
	if err != nil {
		t.Error(err)
	}
	_, err = c.PostTask("bar")
	if err != nil {
		t.Error(err)
	}
}
