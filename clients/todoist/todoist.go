package todoist

import (
	"encoding/json"
	"fmt"
	"github.com/jsutton9/preflight/api/errors"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	Url      string
	Security Security
}

type Security struct {
	Token string `json:"token"`
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

type AddResponse struct {
	SyncStatus map[string]string `json:"SyncStatus"`
	TempIdMapping map[string]int `json:"TempIdMapping"`
}

type DeleteResponse struct {
	SyncStatus map[string]string `json:"SyncStatus"`
}

func New(security Security) Client {
	rand.Seed(time.Now().UnixNano())
	return Client{
		Url:   "https://todoist.com/API/v6/sync",
		Security: security,
	}
}

func buildApiError(function string, command string, status string, body string) *errors.PreflightError {
	return &errors.PreflightError{
		Status: 500,
		InternalMessage: fmt.Sprintf("%s: bad API response for \"%s\": \n" +
			"\t\tStatus: %s\n\t\tBody: %s\n", function, command, status, body),
		ExternalMessage: fmt.Sprintf("Todoist returned an error response: \n" +
			"\t\tStatus: %s\n\t\tBody: %s\n", status, body),
	}
}

func (c Client) PostTask(task string) (int, *errors.PreflightError) {
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
	request := c.Url + "?token=" + c.Security.Token +
		"&commands=[" + string(cmdBytes) + "]"

	response, err := http.Post(request, "", strings.NewReader(""))
	if err != nil {
		return 0, &errors.PreflightError{
			Status: 500,
			InternalMessage: "todoist.Client.PostTask: error posting task: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error posting to Todoist.",
		}
	}
	body := make([]byte, 10000)
	bodyLen, err := response.Body.Read(body)
	if err != nil {
		return 0, &errors.PreflightError{
			Status: 500,
			InternalMessage: "todoist.Client.PostTask: error reading response: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error posting to Todoist.",
		}
	}
	response.Body.Close()

	responseContent := new(AddResponse)
	err = json.Unmarshal(body[:bodyLen], responseContent)
	if err != nil {
		return 0, &errors.PreflightError{
			Status: 500,
			InternalMessage: "todoist.Client.PostTask: error parsing response \"" +
				string(body) + "\": \n\t" + err.Error(),
			ExternalMessage: "We recieved an unrecognized response from Todoist: " +
				"\n\t\"" + string(body) + "\"",
		}
	}

	if (response.StatusCode != 200) || (responseContent.SyncStatus[uuid] != "ok") {
		return 0, buildApiError("todoist.Client.PostTask", "item_add "+task,
			response.Status, string(body))
	}

	id := responseContent.TempIdMapping[tempId]

	return id, nil
}

func (c Client) DeleteTask(id int) *errors.PreflightError {
	uuid := strconv.FormatInt(rand.Int63(), 16)
	ids := []int{id}
	cmd := command{
		Type: "item_delete",
		Uuid: uuid,
		Args: &taskArgs{Ids: ids},
	}

	cmdBytes, _ := json.Marshal(cmd)
	request := c.Url + "?token=" + c.Security.Token +
	        "&commands=[" + string(cmdBytes) + "]"

	response, err := http.Post(request, "", strings.NewReader(""))
	if err != nil {
		return &errors.PreflightError{
			Status: 500,
			InternalMessage: "todoist.Client.DeleteTask: error posting deletion: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error posting to Todoist.",
		}
	}
	body := make([]byte, 10000)
	bodyLen, err := response.Body.Read(body)
	if err != nil {
		return &errors.PreflightError{
			Status: 500,
			InternalMessage: "todoist.Client.DeleteTask: error reading response: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error posting to Todoist.",
		}
	}
	response.Body.Close()

	responseContent := new(DeleteResponse)
	err = json.Unmarshal(body[:bodyLen], responseContent)
	if err != nil {
		return &errors.PreflightError{
			Status: 500,
			InternalMessage: "todoist.Client.DeleteTask: error parsing response \"" +
				string(body) + "\": \n\t" + err.Error(),
			ExternalMessage: "We recieved an unrecognized response from Todoist: " +
				"\n\t\"" + string(body) + "\"",
		}
	}

	syncStatus := responseContent.SyncStatus[uuid]
	if (response.StatusCode != 200) || (syncStatus != "ok") {
		return buildApiError("todoist.Client.DeleteTask", "item_delete "+string(id),
			response.Status, string(body))
	}

	return nil
}
