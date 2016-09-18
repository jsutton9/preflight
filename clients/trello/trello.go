package trello

import (
	"encoding/json"
	"fmt"
	"github.com/jsutton9/preflight/api/errors"
	"io/ioutil"
	"net/http"
)

type Client struct {
	Url       string
	Security  Security
	BoardName string
	Key       string
}

type board struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type card struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Closed bool `json:"closed"`
}

type list struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Cards []card `json:"cards"`
}

type Security struct {
	Token string `json:"token"`
}

type ListKey struct {
	Board string `json:"board"`
	Name string  `json:"name"`
}

func (l *ListKey) Equals(b *ListKey) bool {
	if l==nil && b==nil {
		return true
	} else if l==nil || b==nil {
		return false
	}

	if l.Board != b.Board || l.Name != b.Name {
		return false
	}

	return true
}

func (c Client) get(query string) ([]byte, *errors.PreflightError) {
	request := c.Url + query + "&key=" + c.Key + "&token=" + c.Security.Token
	response, err := http.Get(request)
	if err != nil {
		return nil, &errors.PreflightError{
			Status: 500,
			InternalMessage: "trello.Client.get: error getting " +
				request + ": \n\t" + err.Error(),
			ExternalMessage: "There was an error querying Trello.",
		}
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, &errors.PreflightError{
			Status: 500,
			InternalMessage: "trello.Client.get: error reading response: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error querying Trello.",
		}
	}
	response.Body.Close()

	if response.StatusCode != 200 {
		return nil, &errors.PreflightError{
			Status: 500,
			InternalMessage: fmt.Sprintf(
				"trello.Client.get: bad API response getting %s: " +
				"\n\t\tStatus: %s\n\t\tBody: %s",
				request, response.Status, string(body)),
			ExternalMessage: fmt.Sprintf(
				"Trello returned an error response: " +
				"\n\t\tStatus: %s\n\t\tBody: %s",
				response.Status, string(body)),
			}
	}

	return body, nil
}

func (c Client) boardId(boardName string) (string, *errors.PreflightError) {
	body, pErr := c.get("members/me/boards?fields=name")
	if pErr != nil {
		return "", pErr.Prepend("trello.Client.boardId: error getting boards: ")
	}

	boards := make([]board, 0)
	err := json.Unmarshal(body, &boards)
	if err != nil {
		return "", &errors.PreflightError{
			Status: 500,
			InternalMessage: "trello.Client.boardId: error parsing response " +
				string(body) + "\n\t" + err.Error(),
			ExternalMessage: "There was an error querying Todoist.",
		}
	}

	for _, board := range boards {
		if board.Name == boardName {
			return board.Id, nil
		}
	}

	return "", &errors.PreflightError{
		Status: 404,
		InternalMessage: "trello.Client.boardId: Board named \"" + boardName + "\" not found.",
		ExternalMessage: "Trello board \"" + boardName + "\" not found.",
	}
}

func (c Client) cardNames(boardId, listName string) ([]string, *errors.PreflightError) {
	body, pErr := c.get("boards/" + boardId + "/lists?fields=name&cards=all&card_fields=name,closed")
	if pErr != nil {
		return nil, pErr.Prepend("trello.Client.cardNames: error getting lists: ")
	}

	lists := make([]list, 0)
	err := json.Unmarshal(body, &lists)
	if err != nil {
		return nil, &errors.PreflightError{
			Status: 500,
			InternalMessage: "trello.Client.cardName: error parsing response " +
				string(body) + "\n\t" + err.Error(),
			ExternalMessage: "There was an error querying Todoist.",
		}
	}

	for _, l := range lists {
		if l.Name == listName {
			tasks := make([]string, 0, len(l.Cards))
			for _, c := range l.Cards {
				if ! c.Closed {
					tasks = append(tasks, c.Name)
				}
			}
			return tasks, nil
		}
	}

	return nil, &errors.PreflightError{
		Status: 404,
		InternalMessage: "trello.Client.cardNames: List named \"" + listName + "\" not found.",
		ExternalMessage: "Trello list \"" + listName + "\" not found.",
	}
}

func New(security Security, key string, boardName string) Client {
	return Client{
		Url:       "https://api.trello.com/1/",
		Security:  security,
		BoardName: boardName,
		Key:       key,
	}
}

func (c Client) Tasks(listKey *ListKey) ([]string, *errors.PreflightError) {
	if listKey.Board == "" {
		listKey.Board = c.BoardName
	}

	boardId, err := c.boardId(listKey.Board)
	if err != nil {
		return nil, err.Prepend("trello.Client.Tasks: error getting board ID: ")
	}
	tasks, err := c.cardNames(boardId, listKey.Name)
	if err != nil {
		return nil, err.Prepend("trello.Client.Tasks: error getting card names: ")
	}
	return tasks, nil
}
