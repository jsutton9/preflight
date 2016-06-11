package security

import (
	//"errors"
	"github.com/jsutton9/preflight/clients/todoist"
	"github.com/jsutton9/preflight/clients/trello"
	//"golang.org/x/crypto/bcrypt"
	"time"
)

type SecurityInfo struct {
	PasswordHash []byte      `json:"-"`
	Tokens []Token           `json:"tokens"`
	Todoist todoist.Security `json:"todoistSecurity"`
	Trello trello.Security   `json:"trelloSecurity"`
}

type Token struct {
	Value string                `json:"value"`
	Permissions PermissionFlags `json:"permissions"`
	Expiry time.Time            `json:"expiry"`
	Description string          `json:"description"`
}

type PermissionFlags struct {
	ChecklistRead bool   `json:"checklistRead"`
	ChecklistWrite bool  `json:"checklistWrite"`
	ChecklistInvoke bool `json:"checklistInvoke"`
	GeneralRead bool     `json:"generalRead"`
	GeneralWrite bool    `json:"generalWrite"`
}

/*func New(password string) SecurityInfo, error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("security.New: error hashing password: " +
			"\n\t" + err.Error())
	}

	sec := SecurityInfo{PasswordHash:hash}
	return sec
}

func (s SecurityInfo) ValidatePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(s.PasswordHash, []byte(password))
	if err == nil {
		return true
	} else {
		return false
	}
}*/
