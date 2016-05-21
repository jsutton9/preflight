package trello

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	Url       string
	Key       string
	Token     string
	BoardName string
}

type board struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type card struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type list struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Cards []card `json:"cards"`
}

func (c Client) get(query string) ([]byte, error) {
	request := c.Url + query + "&key=" + c.Key + "&token=" + c.Token
	response, err := http.Get(request)
	if err != nil {
		return nil, errors.New("trello.Client.get: error getting " + request + ": " +
			"\n\t" + err.Error())
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.New("trello.Client.get: error reading response: " +
			"\n\t" + err.Error())
	}
	response.Body.Close()

	if response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf(
			"trello.Client.get: bad API response for \"%s\":\n"+
				"\t\tStatus: %s\n"+
				"\t\tBody: %s\n",
			request, response.Status, string(body)))
	}

	return body, nil
}

func (c Client) boardId() (string, error) {
	body, err := c.get("members/me/boards?fields=name")
	if err != nil {
		return "", errors.New("trello.Client.boardId: error getting boards: " +
			"\n\t" + err.Error())
	}

	boards := make([]board, 0)
	err = json.Unmarshal(body, &boards)
	if err != nil {
		return "", errors.New("trello.Client.boardId: error parsing body: " +
			"\n\t" + err.Error())
	}

	for _, board := range boards {
		if board.Name == c.BoardName {
			return board.Id, nil
		}
	}

	return "", errors.New("trello.Client.boardId: Board named \"" + c.BoardName + "\" not found.")
}

func (c Client) cardNames(boardId, listName string) ([]string, error) {
	body, err := c.get("boards/" + boardId + "/lists?fields=name&cards=all&card_fields=name")
	if err != nil {
		return nil, errors.New("trello.Client.cardNames: error getting lists: " +
			"\n\t" + err.Error())
	}

	lists := make([]list, 0)
	err = json.Unmarshal(body, &lists)
	if err != nil {
		return nil, errors.New("trello.Client.cardNames: error parsing response: " +
			"\n\t" + err.Error())
	}

	for _, l := range lists {
		if l.Name == listName {
			tasks := make([]string, len(l.Cards))
			for i, c := range l.Cards {
				tasks[i] = c.Name
			}
			return tasks, nil
		}
	}

	return nil, errors.New("trello.Client.cardNames: list named \"" + listName + "\" not found.")
}

func New(key, token, boardName string) Client {
	return Client{
		Url:       "https://api.trello.com/1/",
		Key:       key,
		Token:     token,
		BoardName: boardName,
	}
}

func (c Client) Tasks(key, token, boardName, listName string) ([]string, error) {
	newClient := c
	if key != "" {
		newClient.Key = key
	}
	if token != "" {
		newClient.Token = token
	}
	if boardName != "" {
		newClient.BoardName = boardName
	}

	boardId, err := newClient.boardId()
	if err != nil {
		return nil, errors.New("trello.Client.Tasks: error getting board ID: " +
			"\n\t" + err.Error())
	}
	tasks, err := newClient.cardNames(boardId, listName)
	if err != nil {
		return nil, errors.New("trello.Client.Tasks: error getting card names: " +
			"\n\t" + err.Error())
	}
	return tasks, nil
}
