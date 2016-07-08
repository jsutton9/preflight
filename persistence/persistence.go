package persistence

import (
	"github.com/jsutton9/preflight/api/errors"
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

func New(url, database string) (*Persister, *errors.PreflightError) {
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, &errors.PreflightError{
			Status: 500,
			InternalMessage: "persistence.New: " +
				"error dialing \""+url+"\": \n\t" + err.Error(),
			ExternalMessage: "There was an error connecting to the database.",
		}
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

func (p Persister) InitializeNode() *errors.PreflightError {
	secret, pErr := security.GenerateNodeSecret()
	if pErr != nil {
		return pErr.Prepend("persistence.Persister.InitializeNode: error generating secret: ")
	}

	dir := os.Getenv("HOME")+"/preflight/"
	err := ioutil.WriteFile(dir+"secret", []byte(secret), 0600)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0700)
			if err != nil {
				return &errors.PreflightError{
					Status: 500,
					InternalMessage: "persistence.Persister.InitializeNode: " +
						"error making directory \""+dir+"\": \n\t" + err.Error(),
					ExternalMessage: "There was an error initializing the node.",
				}
			}
			err = ioutil.WriteFile(dir+"secret", []byte(secret), 0600)
			if err != nil {
				return &errors.PreflightError{
					Status: 500,
					InternalMessage: "persistence.Persister.InitializeNode: " +
						"error writing secret file: \n\t" + err.Error(),
					ExternalMessage: "There was an error initializing the node.",
				}
			}
		} else {
			return &errors.PreflightError{
				Status: 500,
				InternalMessage: "persistence.Persister.InitializeNode: " +
					"error writing secret file: \n\t" + err.Error(),
				ExternalMessage: "There was an error initializing the node.",
			}
		}
	}

	node := Node{Secret:secret}
	err = p.NodeCollection.Insert(&node)
	if err != nil {
		return &errors.PreflightError{
			Status: 500,
			InternalMessage: "persistence.Persister.InitializeNode: " +
				"error adding node to db: \n\t" + err.Error(),
			ExternalMessage: "There was an error adding the node to the database.",
		}
	}

	return nil
}

func (p Persister) GetNodeSecret() (string, *errors.PreflightError) {
	dir := os.Getenv("HOME")+"/preflight/"
	data, err := ioutil.ReadFile(dir+"secret")
	if err != nil {
		return "", &errors.PreflightError{
			Status: 500,
			InternalMessage: "persistence.GetNodeSecret: error reading secret file: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was a server authentication error.",
		}
	}

	return string(data[:]), nil
}

func (p Persister) ValidateNodeSecret(secret string) (bool, *errors.PreflightError) {
	n, err := p.NodeCollection.Find(bson.M{"secret": secret}).Count()
	if err != nil {
		return false, &errors.PreflightError{
			Status: 500,
			InternalMessage: "persistence.Persister.ValidateNodeSecret: error querying db: " +
				"\n\t" + err.Error(),
			ExternalMessage: "There was an error querying the database.",
		}
	}

	return n>0, nil
}

func (p Persister) AddUser(email, password string) (*User, *errors.PreflightError) {
	existing_user, _ := p.GetUserByEmail(email)
	if existing_user != nil {
		return nil, &errors.PreflightError{
			Status: 409,
			InternalMessage: "persistence.Persister.AddUser: " +
				"\n\tuser with email " + email + " already exists",
			ExternalMessage: "There is already a user with email " + email,
		}
	}

	security, pErr := security.New(password)
	if pErr != nil {
		return nil, pErr.Prepend("persistence.Persister.AddUser: error creating security: ")
	}
	user := User{
		Id: bson.NewObjectId(),
		Email: email,
		Settings: GeneralSettings{},
		Security: security,
		Checklists: make(map[string]*checklist.Checklist),
	}
	err := p.UserCollection.Insert(&user)
	if err != nil {
		return nil, &errors.PreflightError{
			Status: 500,
			InternalMessage: "persistence.Persister.AddUser: " +
				"error inserting user:\n\t" + err.Error(),
			ExternalMessage: "There was an error adding the user to the database.",
		}
	}

	return &user, nil
}

func (p Persister) UpdateUser(user *User) *errors.PreflightError {
	err := p.UserCollection.Update(bson.M{"_id": user.Id}, user)
	if err != nil {
		return &errors.PreflightError{
			Status: 500,
			InternalMessage: "persistence.Persister.UpdateUser: " +
				"error updating user:\n\t" + err.Error(),
			ExternalMessage: "There was an error updating the user in the database.",
		}
	}

	return nil
}

func (p Persister) GetUser(id string) (*User, *errors.PreflightError) {
	user := &User{}
	err := p.UserCollection.FindId(bson.ObjectIdHex(id)).One(user)
	if err != nil {
		return nil, &errors.PreflightError{
			Status: 404,
			InternalMessage: "persistence.Persister.GetUser: " +
				"error finding user id=" + id + ": \n\t" + err.Error(),
			ExternalMessage: "User with id " + id + " not found",
		}
	}

	return user, nil
}

func (p Persister) GetUserByEmail(email string) (*User, *errors.PreflightError) {
	user := &User{}
	err := p.UserCollection.Find(bson.M{"email": email}).One(user)
	if err != nil {
		return nil, &errors.PreflightError{
			Status: 404,
			InternalMessage: "persistence.Persister.GetUserByEmail: " +
				"error finding user email=" + email + ": \n\t" + err.Error(),
			ExternalMessage: "User with email " + email + " not found",
		}
	}

	return user, nil
}

func (p Persister) GetUserByToken(secret string) (*User, *errors.PreflightError) {
	user := &User{}
	err := p.UserCollection.Find(bson.M{"security.tokens.secret": secret}).One(user)
	if err != nil {
		return nil, &errors.PreflightError{
			Status: 403,
			InternalMessage: "persistence.Persister.GetUserByToken: " +
				"error finding user by token: \n\t" + err.Error(),
			ExternalMessage: "No user with that token was found",
		}
	}

	return user, nil
}
