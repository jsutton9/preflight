package persistence

import (
	"errors"
	"github.com/jsutton9/preflight/checklist"
	"github.com/jsutton9/preflight/security"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"os"
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

type Persister struct {
	Session *mgo.Session
	UserCollection *mgo.Collection
	NodeCollection *mgo.Collection
}

type Node struct {
	Secret string
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

	userCollection := session.DB(database).C("users")
	nodeCollection := session.DB(database).C("nodes")

	p := Persister{
		Session: session,
		UserCollection: userCollection,
		NodeCollection: nodeCollection,
	}

	return &p, nil
}

func (p Persister) Close() {
	p.Session.Close()
}

func (p Persister) InitializeNode() error {
	secret, err := security.GenerateNodeSecret()
	if err != nil {
		return errors.New("persistence.Persister.InitializeNode: error generating secret: " +
			"\n\t" + err.Error())
	}

	dir := os.Getenv("HOME")+"/preflight/"
	err = ioutil.WriteFile(dir+"secret", []byte(secret), 0600)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0700)
			if err != nil {
				return errors.New("persistence.Persister.InitializeNode: error making directory \"" +
					dir + "\": \n\t" + err.Error())
			}
			err = ioutil.WriteFile(dir+"secret", []byte(secret), 0600)
			if err != nil {
				return errors.New("persistence.Persister.InitializeNode: error writing secret file: " +
					"\n\t" + err.Error())
			}
		} else {
			return errors.New("persistence.Persister.InitializeNode: error writing secret file: " +
				"\n\t" + err.Error())
		}
	}

	node := Node{Secret:secret}
	err = p.NodeCollection.Insert(&node)
	if err != nil {
		return errors.New("persistence.Persister.InitializeNode: error adding node to db: " +
			"\n\t" + err.Error())
	}

	return nil
}

func (p Persister) GetNodeSecret() (string, error) {
	dir := os.Getenv("HOME")+"/preflight/"
	data, err := ioutil.ReadFile(dir+"secret")
	if err != nil {
		return "", errors.New("persistence.GetNodeSecret: error reading secret file: " +
			"\n\t" + err.Error())
	}

	return string(data[:]), nil
}

func (p Persister) ValidateNodeSecret(secret string) (bool, error) {
	n, err := p.NodeCollection.Find(bson.M{"secret": secret}).Count()
	if err != nil {
		return false, errors.New("persistence.Persister.ValidateNodeSecret: error querying db: " +
			"\n\t" + err.Error())
	}

	return n>0, nil
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
		Checklists: make(map[string]*checklist.Checklist),
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
