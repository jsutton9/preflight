package client

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	Url   string
	Token string
}

type taskArgs struct {
	Content string `json:"content"`
}

type command struct {
	Type string    `json:"type"`
	Uuid string    `json:"uuid"`
	TempId string  `json:"temp_id"`
	Args *taskArgs `json:"args"`
}

type ApiError struct {
	Command      string
	Status       string
	ResponseBody string
}

func New(token string) Client {
	rand.Seed(time.Now().UnixNano())
	return Client{
		Url:   "https://todoist.com/API/v6/sync",
		Token: token,
	}
}

func (e ApiError) Error() string {
	return fmt.Sprintf(
		"Bad API response for \"%s\":\n"+
			"Status: %s\n"+
			"Body: %s\n",
		e.Command, e.Status, e.ResponseBody,
	)
}

func (c Client) PostTask(task string) (string, error) {
	uuid := strconv.FormatInt(rand.Int63(), 16)
	tempId := strconv.FormatInt(rand.Int63(), 16)
	cmd := command{
		Type: "item_add",
		Uuid: uuid,
		TempId: tempId,
		Args: &taskArgs{Content: task},
	}

	cmdBytes, _ := json.Marshal(cmd)
	request := c.Url + "?token=" + c.Token +
		"&commands=[" + string(cmdBytes) + "]"

	response, err := http.Post(request, "", strings.NewReader(""))
	if err != nil {
		return uuid, err
	}
	body := make([]byte, 10000)
	response.Body.Read(body)
	response.Body.Close()

	pattern := fmt.Sprintf(".*\"%s\":\\s*\"[oO][kK]\"", uuid)
	syncStatusOk, _ := regexp.Match(pattern, body)

	if (response.StatusCode != 200) || (! syncStatusOk) {
		return uuid, ApiError{
			Command:      "item_add " + task,
			Status:       response.Status,
			ResponseBody: string(body),
		}
	}

	return uuid, nil
}
