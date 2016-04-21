package client

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	Url   string
	Token string
}

type taskArgs struct {
	Content string `json:"content,omitempty"`
	Ids []int      `json:"ids,omitempty"`
}

type command struct {
	Type string    `json:"type"`
	Uuid string    `json:"uuid"`
	TempId string  `json:"temp_id,omitempty"`
	Args *taskArgs `json:"args"`
}

type ApiError struct {
	Command      string
	Status       string
	ResponseBody string
}

type AddResponse struct {
	SyncStatus map[string]string `json:"SyncStatus"`
	TempIdMapping map[string]int `json:"TempIdMapping"`
}

type DeleteResponse struct {
	SyncStatus map[string]string `json:"SyncStatus"`
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

func (c Client) PostTask(task string) (int, error) {
	task = url.QueryEscape(task)

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
		return 0, err
	}
	body := make([]byte, 10000)
	bodyLen, err := response.Body.Read(body)
	if err != nil {
		return 0, err
	}
	response.Body.Close()

	responseContent := new(AddResponse)
	err = json.Unmarshal(body[:bodyLen], responseContent)
	if err != nil {
		return 0, err
	}

	if (response.StatusCode != 200) || (responseContent.SyncStatus[uuid] != "ok") {
		return 0, ApiError{
			Command:      "item_add " + task,
			Status:       response.Status,
			ResponseBody: string(body),
		}
	}

	id := responseContent.TempIdMapping[tempId]

	return id, nil
}

func (c Client) DeleteTask(id int) error {
	uuid := strconv.FormatInt(rand.Int63(), 16)
	ids := []int{id}
	cmd := command{
		Type: "item_delete",
		Uuid: uuid,
		Args: &taskArgs{Ids: ids},
	}

	cmdBytes, _ := json.Marshal(cmd)
	request := c.Url + "?token=" + c.Token +
	        "&commands=[" + string(cmdBytes) + "]"

	response, err := http.Post(request, "", strings.NewReader(""))
	if err != nil {
		return err
	}
	body := make([]byte, 10000)
	bodyLen, err := response.Body.Read(body)
	if err != nil {
		return err
	}
	response.Body.Close()

	responseContent := new(DeleteResponse)
	err = json.Unmarshal(body[:bodyLen], responseContent)
	if err != nil {
		return err
	}

	syncStatus := responseContent.SyncStatus[uuid]
	if (response.StatusCode != 200) || (syncStatus != "ok") {
		return ApiError{
			Command:      fmt.Sprintf("item_delete %d", id),
			Status:       response.Status,
			ResponseBody: string(body),
		}
	}

	return nil
}
