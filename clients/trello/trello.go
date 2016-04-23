package trello

import (
	"encoding/json"
	"errors"
)

type Client struct {
	Url       string
	Key       string
	Token     string
	BoardName string
}

type userResponse struct {
	//TODO
}

type listsResponse struct {
	//TODO
}

func (c Client) boardId(key, token, boardName string) (string, error) {
	//TODO: query /1/members/me, extract id from boardName
}

func (c Client) cardNames(key, token, boardId, listName string) ([]string, error) {
	//TODO: query /1/boards/[boardId]/lists, extract card names from listName
}

func (c Client) Tasks(key, token, boardName, listName string) ([]string, error) {
	if key == "" {
		key = c.Key
		if key == "" {
			return nil, errors.New("No Todoist application key supplied")
		}
	}
	if token == "" {
		token = c.Token
		if token == "" {
			return nil, errors.New("No Todoist auth token supplied")
		}
	}
	if boardName == "" {
		boardName = c.BoardName
		if boardName == "" {
			return nil, errors.New("No Todoist board name supplied")
		}
	}

	boardId, err := c.boardId(key, token, boardName)
	if err != nil {
		return nil, err
	}
	tasks, err := c.cardNames(key, token, boardId, listName)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (c Client) Tasks(listName string) ([]string, error) {
	return c.Tasks("", "", "", listName)
}
