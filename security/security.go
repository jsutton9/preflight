package security

import (
	"time"
)

type SecurityInfo struct {
	Email string            `json:"email"`
	PasswordHash string     `json:"-"`
	PasswordSalt string     `json:"-"`
	Tokens []Token          `json:"tokens"`
	Todoist TodoistSecurity `json:"todoistSecurity"`
	Trello TrelloSecurity   `json:"trelloSecurity"`
}

type Token struct {
	Value string                `json:"value"`
	Permissions PermissionFlags `json:"permissions"`
	Expiry time.Time            `json:"expiry"`
	Description String          `json:"description"`
}

type PermissionFlags struct {
	ChecklistRead bool   `json:"checklistRead"`
	ChecklistWrite bool  `json:"checklistWrite"`
	ChecklistInvoke bool `json:"checklistInvoke"`
	GeneralRead bool     `json:"generalRead"`
	GeneralWrite bool    `json:"generalWrite"`
}

type TodoistSecurity struct {
	Token string `json:"token"`
}

type TrelloSecurity struct {
	Key string   `json:"key"`
	Token string `json:"token"`
}
