package user

import (
	"github.com/jsutton9/preflight/api/errors"
	"github.com/jsutton9/preflight/checklist"
	"github.com/jsutton9/preflight/security"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	Id bson.ObjectId                           `json:"id" bson:"_id,omitempty"`
	Email string                               `json:"email"`
	Settings GeneralSettings                   `json:"generalSettings"`
	Security *security.SecurityInfo            `json:"security"`
	Checklists map[string]*checklist.Checklist `json:"checklists"`
}

type GeneralSettings struct {
	Timezone string    `json:"timezone"`
	TrelloBoard string `json:"trelloBoard"`
}

func (u *User) GetId() string {
	return u.Id.Hex()
}

func New(email, password string) (*User, *errors.PreflightError) {
	security, err := security.New(password)
	if err != nil {
		return nil, err.Prepend("persistence.Persister.AddUser: error creating security: ")
	}

	newUser := User{
		Id: bson.NewObjectId(),
		Email: email,
		Settings: GeneralSettings{},
		Security: security,
		Checklists: make(map[string]*checklist.Checklist),
	}

	return &newUser, nil
}
