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
	Checklists []*checklist.Checklist          `json:"checklists"`
}

type GeneralSettings struct {
	Timezone string    `json:"timezone"`
	TrelloBoard string `json:"trelloBoard"`
}

type UserDelta struct {
	User *User
	RemoveUser bool
	Added []*checklist.Checklist
	Removed []*checklist.Checklist
	Updated []*checklist.Checklist
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
		Checklists: make([]*checklist.Checklist, 0),
	}

	return &newUser, nil
}

/*
func (u *User) Invoke(name, trelloKey string) *errors.PreflightError {
	//TODO
}
*/

func (oldUser *User) GetDelta(newUser *User) *UserDelta {
	if oldUser == nil && newUser == nil {
		return nil
	} else if oldUser == nil {
		return &UserDelta{
			User: newUser,
			Added: newUser.Checklists,
		}
	} else if newUser == nil {
		return &UserDelta{
			User: oldUser,
			RemoveUser: true,
			Removed: oldUser.Checklists,
		}
	}

	delta := &UserDelta{
		User: newUser,
		Added: make([]*checklist.Checklist, 0),
		Removed: make([]*checklist.Checklist, 0),
		Updated: make([]*checklist.Checklist, 0),
	}

	oldChecklists := make(map[string]*checklist.Checklist)
	matched := make(map[string]bool)
	for _, cl := range oldUser.Checklists {
		oldChecklists[cl.Id] = cl
		matched[cl.Id] = false
	}

	for _, cl := range newUser.Checklists {
		old, found := oldChecklists[cl.Id]
		if ! found {
			delta.Added = append(delta.Added, cl)
		} else {
			matched[cl.Id] = true
			if ! cl.Equals(old) {
				delta.Updated = append(delta.Updated, cl)
			}
		}
	}

	for _, cl := range oldUser.Checklists {
		if ! matched[cl.Id] {
			delta.Removed = append(delta.Removed, cl)
		}
	}

	return delta
}
