package security

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/jsutton9/preflight/clients/todoist"
	"github.com/jsutton9/preflight/clients/trello"
	"golang.org/x/crypto/bcrypt"
	"time"
)

const (
	ID_BITS = 64
	SECRET_BITS = 64
)

type SecurityInfo struct {
	PasswordHash []byte      `json:"-"`
	Tokens []Token           `json:"tokens"`
	Todoist todoist.Security `json:"todoistSecurity"`
	Trello trello.Security   `json:"trelloSecurity"`
}

type Token struct {
	Id string                   `json:"id"`
	Secret string               `json:"secret"`
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

func New(password string) SecurityInfo, error {
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
	}
	return false
}

func (s SecurityInfo) ValidateToken(secret string, permissions PermissionFlags) bool {
	for _, token in s.Tokens {
		if token.Secret == secret {
			now := time.Now()
			if now.After(token.Expiry) {
				return false
			}
			if permissions.ChecklistRead && (! token.Permissions.ChecklistRead) ||
				permissions.ChecklistWrite && (! token.Permissions.ChecklistWrite) ||
				permissions.ChecklistInvoke && (! token.Permissions.ChecklistInvoke) ||
				permissions.GeneralRead && (! token.Permissions.GeneralRead) ||
				permissions.GeneralWrite && (! token.Permissions.GeneralWrite) {
					return false
				}
			return true
		}
	}

	return false
}

func (s SecurityInfo) AddToken(permissions PermissionFlags, expiryHours int, description string) (Token, error) {
	now := time.Now()
	dur := time.Duration(expiryHours)*time.Hour
	expiry = now.Add(dur)

	intId, err := rand.Int(rand.Reader, big.NewInt(ID_BITS))
	if err != nil {
		return "", errors.New("security.AddToken: error generating id: " +
			"\n\t" + err.Error())
	}
	id := fmt.SPrintf("%0x", intId)

	intSecret, err := rand.Int(rand.Reader, big.NewInt(SECRET_BITS))
	if err != nil {
		return "", errors.New("security.AddToken: error generating secret: " +
			"\n\t" + err.Error())
	}
	secret := fmt.SPrintf("%0x", intSecret)

	token := Token{
		Id: id,
		Secret: secret,
		Permissions: permissions,
		Expiry: expiry,
		Description: description,
	}

	append(s.Tokens, token)

	return token, nil
}

func (s SecurityInfo) DeleteToken(id string) error {
	for i, token := range s.Tokens {
		if token.Id == id {
			l := len(s.Tokens)
			s.Tokens[i] = s.Tokens[l-1]
			s.Tokens = s.Tokens[:l-1]
			return nil
		}
	}

	return errors.New("security.DeleteToken: token \""+id+"\" not found")
}
