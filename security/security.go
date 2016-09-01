package security

import (
	"crypto/rand"
	"fmt"
	"github.com/jsutton9/preflight/api/errors"
	"github.com/jsutton9/preflight/clients/todoist"
	"github.com/jsutton9/preflight/clients/trello"
	"golang.org/x/crypto/bcrypt"
	"math/big"
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
	Secret string               `json:"secret,omitempty"`
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

func New(password string) (*SecurityInfo, *errors.PreflightError) {
	sec := SecurityInfo{}
	err := sec.SetPassword(password)
	if err != nil {
		return nil, err.Prepend("security.New: error setting password: ")
	}
	return &sec, nil
}

func (s *SecurityInfo) ValidatePassword(password string) *errors.PreflightError {
	err := bcrypt.CompareHashAndPassword(s.PasswordHash, []byte(password))
	if err != nil {
		return &errors.PreflightError{
			Status: 401,
			InternalMessage: "security.ValidatePassword: error validating password: " +
				"\n\t" + err.Error(),
			ExternalMessage: "Invalid password",
		}
	} else {
		return nil
	}
}

func (s *SecurityInfo) SetPassword(newPassword string) *errors.PreflightError {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return &errors.PreflightError{
			Status: 500,
			InternalMessage: "security.SetPassword: error hashing password: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error setting the password.",
		}
	}
	s.PasswordHash = hash

	return nil
}

func (s *SecurityInfo) ValidateToken(secret string, permissions PermissionFlags) *errors.PreflightError {
	for _, token := range s.Tokens {
		if token.Secret == secret {
			now := time.Now()
			if now.After(token.Expiry) {
				return &errors.PreflightError{
					Status: 401,
					InternalMessage: "security.ValidateToken: token expired",
					ExternalMessage: "The user token is expired.",
				}
			}
			if permissions.ChecklistRead && (! token.Permissions.ChecklistRead) ||
				permissions.ChecklistWrite && (! token.Permissions.ChecklistWrite) ||
				permissions.ChecklistInvoke && (! token.Permissions.ChecklistInvoke) ||
				permissions.GeneralRead && (! token.Permissions.GeneralRead) ||
				permissions.GeneralWrite && (! token.Permissions.GeneralWrite) {
					return &errors.PreflightError{
						Status: 401,
						InternalMessage: "security.ValidateToken: " +
							"insufficient permissions",
						ExternalMessage: "The user token does not have sufficient permissions.",
					}
				}
			return nil
		}
	}

	return &errors.PreflightError{
		Status: 401,
		InternalMessage: "security.ValidateToken: token not found",
		ExternalMessage: "The user token was absent or not recognized.",
	}
}

func (s *SecurityInfo) AddToken(permissions PermissionFlags, expiryHours int, description string) (*Token, *errors.PreflightError) {
	now := time.Now()
	dur := time.Duration(expiryHours)*time.Hour
	expiry := now.Add(dur)

	idMax := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(ID_BITS), nil)
	intId, err := rand.Int(rand.Reader, idMax)
	if err != nil {
		return nil, &errors.PreflightError{
			Status: 500,
			InternalMessage: "security.AddToken: error generating id: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error creating the token.",
		}
	}
	idPattern := fmt.Sprintf("%%0%dx", ID_BITS/4)
	id := fmt.Sprintf(idPattern, intId)

	secretMax := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(SECRET_BITS), nil)
	intSecret, err := rand.Int(rand.Reader, secretMax)
	if err != nil {
		return nil, &errors.PreflightError{
			Status: 500,
			InternalMessage: "security.AddToken: error generating secret: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error creating the token.",
		}
	}
	secretPattern := fmt.Sprintf("%%0%dx", SECRET_BITS/4)
	secret := fmt.Sprintf(secretPattern, intSecret)

	token := Token{
		Id: id,
		Secret: secret,
		Permissions: permissions,
		Expiry: expiry,
		Description: description,
	}

	s.Tokens = append(s.Tokens, token)

	return &token, nil
}

func (s *SecurityInfo) DeleteToken(id string) *errors.PreflightError {
	for i, token := range s.Tokens {
		if token.Id == id {
			l := len(s.Tokens)
			s.Tokens[i] = s.Tokens[l-1]
			s.Tokens = s.Tokens[:l-1]
			return nil
		}
	}

	return &errors.PreflightError{
		Status: 404,
		InternalMessage: "security.DeleteToken: token \""+id+"\" not found",
		ExternalMessage: "Token not found",
	}
}

func GenerateNodeSecret() (string, *errors.PreflightError) {
	secretMax := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(SECRET_BITS), nil)
	intSecret, err := rand.Int(rand.Reader, secretMax)
	if err != nil {
		return "", &errors.PreflightError{
			Status: 500,
			InternalMessage: "security.GenerateNodeSecret: error generating secret: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error registering the node.",
		}
	}
	secretPattern := fmt.Sprintf("%%0%dx", SECRET_BITS/4)
	return fmt.Sprintf(secretPattern, intSecret), nil
}
