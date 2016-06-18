package persistence

import (
	"errors"
	"github.com/jsutton9/preflight/checklist"
	"github.com/jsutton9/preflight/security"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	Id bson.ObjectId                          `json:"id" bson:"_id,omitempty"`
	Email string                              `json:"email"`
	Settings GeneralSettings                  `json:"generalSettings"`
	Security *security.SecurityInfo           `json:"security"`
	Checklists map[string]checklist.Checklist `json:"checklists"`
}

type GeneralSettings struct {
	Timezone string    `json:"timezone"`
	TrelloBoard string `json:"trelloBoard"`
}

type Persister struct {
	Session *mgo.Session
	UserCollection *mgo.Collection
}

func (u *User) GetId() string {
	return u.Id.Hex()
}

func New(url, database string) (*Persister, error) {
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, errors.New("persistence.New: " +
			"error dialing \""+url+"\": \n\t" + err.Error())
	}

	collection := session.DB(database).C("users")

	p := Persister{
		Session: session,
		UserCollection: collection,
	}

	return &p, nil
}

func (p Persister) Close() {
	p.Session.Close()
}

func (p Persister) AddUser(email, password string) (*User, error) {
	existing_user, _ := p.GetUserByEmail(email)
	if existing_user != nil {
		return nil, errors.New("persistence.Persister.AddUser: " +
			"\n\t" + "user with email " + email + " already exists")
	}

	security, err := security.New(password)
	if err != nil {
		return nil, errors.New("persistence.Persister.AddUser: " +
			"error creating security:\n\t" + err.Error())
	}
	user := User{
		Id: bson.NewObjectId(),
		Email: email,
		Settings: GeneralSettings{},
		Security: security,
		Checklists: make(map[string]checklist.Checklist),
	}
	err = p.UserCollection.Insert(&user)
	if err != nil {
		return nil, errors.New("persistence.Persister.AddUser: " +
			"error inserting user:\n\t" + err.Error())
	}

	return &user, nil
}

func (p Persister) UpdateUser(user *User) error {
	err := p.UserCollection.Update(bson.M{"_id": user.Id}, user)
	if err != nil {
		return errors.New("persistence.Persister.UpdateUser: " +
			"error updating user:\n\t" + err.Error())
	}

	return nil
}

func (p Persister) GetUser(id string) (*User, error) {
	user := &User{}
	err := p.UserCollection.FindId(bson.ObjectIdHex(id)).One(user)
	if err != nil {
		return nil, errors.New("persistence.Persister.GetUser: " +
			"error finding user id=" + id + ": \n\t" + err.Error())
	}

	return user, nil
}

func (p Persister) GetUserByEmail(email string) (*User, error) {
	user := &User{}
	err := p.UserCollection.Find(bson.M{"email": email}).One(user)
	if err != nil {
		return nil, errors.New("persistence.Persister.GetUserByEmail: " +
			"error finding user email=" + email + ": \n\t" + err.Error())
	}

	return user, nil
}

func (p Persister) GetUserByToken(secret string) (*User, error) {
	user := &User{}
	err := p.UserCollection.Find(bson.M{"security.tokens.secret": secret}).One(user)
	if err != nil {
		return nil, errors.New("persistence.Persister.GetUserByToken: " +
			"error finding user by token: \n\t" + err.Error())
	}

	return user, nil
}
